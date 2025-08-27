package hasher

import (
	crand "crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"golang.org/x/crypto/scrypt"
)

// ===============================
// Scrypt Provider (PHC-like format)
// ===============================

type Scrypt struct {
	N       int
	R       int
	P       int
	KeyLen  int
	SaltLen int
}

func DefaultScrypt() *Scrypt { return &Scrypt{N: 1 << 15, R: 8, P: 1, KeyLen: 32, SaltLen: 16} }

func (s *Scrypt) ID() string { return "scrypt" }

func (s *Scrypt) Hash(password []byte) (string, error) {
	salt := make([]byte, s.SaltLen)
	if _, err := crand.Read(salt); err != nil {
		return "", err
	}
	dk, err := scrypt.Key(password, salt, s.N, s.R, s.P, s.KeyLen)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("$scrypt$N=%d,r=%d,p=%d$%s$%s", s.N, s.R, s.P, b64Raw(salt), b64Raw(dk)), nil
}

var phcScryptRe = regexp.MustCompile(`^\$scrypt\$N=(\d+),r=(\d+),p=(\d+)\$([A-Za-z0-9_-]+)\$([A-Za-z0-9_-]+)$`)

func (s *Scrypt) Verify(password []byte, encoded string) (bool, error) {
	m := phcScryptRe.FindStringSubmatch(encoded)
	if m == nil {
		return false, errors.New("hasher/scrypt: invalid string")
	}
	N64, _ := strconv.ParseInt(m[1], 10, 0)
	R64, _ := strconv.ParseInt(m[2], 10, 0)
	P64, _ := strconv.ParseInt(m[3], 10, 0)
	salt, err := b64RawDecode(m[4])
	if err != nil {
		return false, err
	}
	sum, err := b64RawDecode(m[5])
	if err != nil {
		return false, err
	}
	calc, err := scrypt.Key(password, salt, int(N64), int(R64), int(P64), len(sum))
	if err != nil {
		return false, err
	}
	if subtle.ConstantTimeCompare(calc, sum) == 1 {
		return true, nil
	}
	return false, nil
}

func (s *Scrypt) NeedsRehash(encoded string) bool {
	m := phcScryptRe.FindStringSubmatch(encoded)
	if m == nil {
		return true
	}
	N64, _ := strconv.ParseInt(m[1], 10, 0)
	R64, _ := strconv.ParseInt(m[2], 10, 0)
	P64, _ := strconv.ParseInt(m[3], 10, 0)
	salt, err1 := b64RawDecode(m[4])
	sum, err2 := b64RawDecode(m[5])
	if err1 != nil || err2 != nil {
		return true
	}
	if int(N64) != s.N || int(R64) != s.R || int(P64) != s.P {
		return true
	}
	if len(salt) != s.SaltLen || len(sum) != s.KeyLen {
		return true
	}
	return false
}
