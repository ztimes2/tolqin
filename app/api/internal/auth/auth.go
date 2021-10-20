package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/bcrypt"
)

const (
	defaultSaltByteSize = 16
	minPasswordLength   = 8
)

// PasswordSalter takes care of salting passwords.
type PasswordSalter struct {
	byteSize int
	reader   io.Reader
	encodeFn func([]byte) string
}

// NewPasswordSalter returns a new *PasswordSalter.
func NewPasswordSalter() *PasswordSalter {
	return &PasswordSalter{
		reader:   rand.Reader,
		byteSize: defaultSaltByteSize,
		encodeFn: base64.URLEncoding.EncodeToString,
	}
}

// GenerateSalt generates and returns a random salt.
func (p *PasswordSalter) GenerateSalt() (string, error) {
	b := make([]byte, p.byteSize)
	if _, err := p.reader.Read(b); err != nil {
		return "", err
	}

	return p.encodeFn(b), nil
}

// SaltPassword salts the given password with the given salt and returns a salted
// result.
func (p *PasswordSalter) SaltPassword(password, salt string) string {
	return password + salt
}

var (
	// ErrMismatchedPassword is used when hash and password are mismatched during
	// comparison.
	ErrMismatchedHashAndPassword = errors.New("mismatched hash and password")
)

// PasswordHasher takes care of hashing passwords.
type PasswordHasher struct {
	cost       int
	generateFn func(password []byte, cost int) ([]byte, error)
	compareFn  func(hash, password []byte) error
}

// NewPasswordHasher returns a new *PasswordHasher.
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		cost:       bcrypt.DefaultCost,
		generateFn: bcrypt.GenerateFromPassword,
		compareFn:  bcrypt.CompareHashAndPassword,
	}
}

// HashPassword hashes the given password.
func (p *PasswordHasher) HashPassword(password string) (string, error) {
	hash, err := p.generateFn([]byte(password), p.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CompareHashAndPassword compares the given hash and password, and returns nil
// if they match.
//
// ErrMismatchedHashAndPassword is returned when hash and password don't match.
func (p *PasswordHasher) CompareHashAndPassword(hash, password string) error {
	if err := p.compareFn([]byte(hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrMismatchedHashAndPassword
		}
		return err
	}
	return nil
}

// IsPassword checks if the given string is a valid password.
func IsPassword(password string) bool {
	if len(password) < minPasswordLength {
		return false
	}

	// TODO check if password consists of allowed character set
	return true
}
