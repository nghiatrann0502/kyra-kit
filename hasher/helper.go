package hasher

import (
	crand "crypto/rand"
	"encoding/base64"
	"errors"
	"math/big"
	"regexp"
)

// ===============================
// Helpers & Defaults
// ===============================

func b64Raw(b []byte) string                { return base64.RawStdEncoding.EncodeToString(b) }
func b64RawDecode(s string) ([]byte, error) { return base64.RawStdEncoding.DecodeString(s) }

var phcIDRe = regexp.MustCompile(`^\$([a-z0-9]+)\$`)

func extractPHCID(s string) string {
	m := phcIDRe.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	return m[1]
}

func cryptoRandIndex(n int) (int, error) {
	if n <= 0 {
		return 0, errors.New("invalid n")
	}
	max := new(big.Int).SetInt64(int64(n))
	r, err := crand.Int(crand.Reader, max)
	if err != nil {
		return 0, err
	}
	return int(r.Int64()), nil
}

func cryptoRandUint64N(n uint64) (uint64, error) {
	if n == 0 {
		return 0, errors.New("invalid n=0")
	}
	max := new(big.Int).SetUint64(n)
	r, err := crand.Int(crand.Reader, max)
	if err != nil {
		return 0, err
	}
	return r.Uint64(), nil
}

func MustDefaultManager() *Manager {
	return MustNewManager("argon2id", DefaultArgon2id(), DefaultBcrypt())
}

func MustExtendedManager() *Manager {
	return MustNewManager(
		"argon2id",
		DefaultArgon2id(),
		DefaultBcrypt(),
		DefaultScrypt(),
	)
}
