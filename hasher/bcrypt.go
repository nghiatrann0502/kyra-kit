package hasher

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// ===============================
// Bcrypt Provider (native format)
// ===============================

type Bcrypt struct{ Cost int }

func DefaultBcrypt() *Bcrypt { return &Bcrypt{Cost: 12} }

func (b *Bcrypt) ID() string { return "bcrypt" }

func (b *Bcrypt) Hash(password []byte) (string, error) {
	h, err := bcrypt.GenerateFromPassword(password, b.Cost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func (b *Bcrypt) Verify(password []byte, encoded string) (bool, error) {
	if isBcryptNative(encoded) {
		err := bcrypt.CompareHashAndPassword([]byte(encoded), password)
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return err == nil, err
	}
	return false, errors.New("hasher/bcrypt: invalid bcrypt string")
}

func (b *Bcrypt) NeedsRehash(encoded string) bool {
	if !isBcryptNative(encoded) {
		return true
	}
	cost, err := bcrypt.Cost([]byte(encoded))
	if err != nil {
		return true
	}
	return cost != b.Cost
}

func isBcryptNative(s string) bool {
	return strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$")
}
