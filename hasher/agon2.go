package hasher

import (
	crand "crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"golang.org/x/crypto/argon2"
)

// ===============================
// Argon2id Provider (PHC format)
// ===============================

type Argon2idParams struct {
	Memory      uint32 // KiB
	Time        uint32 // iterations
	Parallelism uint8
	SaltLen     uint32
	KeyLen      uint32
}

type Argon2id struct{ params Argon2idParams }

func (a *Argon2id) ID() string { return "argon2id" }

// DefaultArgon2id returns a safe baseline for interactive logins.
func DefaultArgon2id() *Argon2id {
	return &Argon2id{params: Argon2idParams{
		Memory:      64 * 1024, // 64 MiB
		Time:        3,
		Parallelism: 2,
		SaltLen:     16,
		KeyLen:      32,
	}}
}

func (a *Argon2id) Hash(password []byte) (string, error) {
	salt := make([]byte, a.params.SaltLen)
	if _, err := crand.Read(salt); err != nil {
		return "", err
	}
	sum := argon2.IDKey(password, salt, a.params.Time, a.params.Memory, a.params.Parallelism, a.params.KeyLen)
	encSalt := b64Raw(salt)
	encSum := b64Raw(sum)
	// PHC: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	phc := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", a.params.Memory, a.params.Time, a.params.Parallelism, encSalt, encSum)
	return phc, nil
}

var phcArgon2Re = regexp.MustCompile(`^\$argon2id\$v=19\$m=(\d+),t=(\d+),p=(\d+)\$([A-Za-z0-9_-]+)\$([A-Za-z0-9_-]+)$`)

func (a *Argon2id) Verify(password []byte, encoded string) (bool, error) {
	m := phcArgon2Re.FindStringSubmatch(encoded)
	if m == nil {
		return false, errors.New("hasher/argon2id: invalid PHC string")
	}
	mem, _ := strconv.ParseUint(m[1], 10, 32)
	tim, _ := strconv.ParseUint(m[2], 10, 32)
	par, _ := strconv.ParseUint(m[3], 10, 8)
	salt, err := b64RawDecode(m[4])
	if err != nil {
		return false, err
	}
	sum, err := b64RawDecode(m[5])
	if err != nil {
		return false, err
	}
	calc := argon2.IDKey(password, salt, uint32(tim), uint32(mem), uint8(par), uint32(len(sum)))
	if subtle.ConstantTimeCompare(calc, sum) == 1 {
		return true, nil
	}
	return false, nil
}

func (a *Argon2id) NeedsRehash(encoded string) bool {
	m := phcArgon2Re.FindStringSubmatch(encoded)
	if m == nil {
		return true
	}
	mem, _ := strconv.ParseUint(m[1], 10, 32)
	tim, _ := strconv.ParseUint(m[2], 10, 32)
	par, _ := strconv.ParseUint(m[3], 10, 8)
	return uint32(mem) != a.params.Memory || uint32(tim) != a.params.Time || uint8(par) != a.params.Parallelism
}
