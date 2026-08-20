// Package password provides bcrypt password hashing and comparison utilities.
package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const defaultCost = bcrypt.DefaultCost + 2 // cost 12

// Hash returns a bcrypt hash of the given plaintext password.
func Hash(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), defaultCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// Compare returns true if plaintext matches the stored bcrypt hash.
func Compare(hash, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}
