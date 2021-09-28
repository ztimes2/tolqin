package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type PasswordSalter struct {
	byteSize int
	reader   io.Reader
	encodeFn func([]byte) string
}

func NewPasswordSalter(byteSize int) *PasswordSalter {
	return &PasswordSalter{
		reader:   rand.Reader,
		byteSize: byteSize,
		encodeFn: base64.URLEncoding.EncodeToString,
	}
}

func (p *PasswordSalter) SaltPassword(password string) (salted, salt string, err error) {
	b := make([]byte, p.byteSize)
	if _, err := p.reader.Read(b); err != nil {
		return "", "", err
	}

	salt = p.encodeFn(b)

	return password + salt, salt, nil
}

var (
	ErrMismatchedPassword = errors.New("mismatched password")
)

type PasswordHasher struct {
	cost              int
	hashPasswordFn    func(password []byte, cost int) ([]byte, error)
	comparePasswordFn func(hash, password []byte) error
}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		cost:              bcrypt.DefaultCost,
		hashPasswordFn:    bcrypt.GenerateFromPassword,
		comparePasswordFn: bcrypt.CompareHashAndPassword,
	}
}

func (p *PasswordHasher) HashPassword(password string) (string, error) {
	hash, err := p.hashPasswordFn([]byte(password), p.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (p *PasswordHasher) ComparePassword(hash, password string) error {
	if err := p.comparePasswordFn([]byte(hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrMismatchedPassword
		}
		return err
	}
	return nil
}

type Tokener interface { // TODO turn into struct
	Token(User) (string, error)
	ParseToken(token string) (TokenClaims, error)
}

type TokenClaims struct {
	jwt.StandardClaims

	Email string `json:"email,omitempty"`
	Role  string `json:"role,omitempty"`
}

func (t TokenClaims) UserInfo() UserInfo {
	return UserInfo{
		ID:    t.Subject,
		Email: t.Email,
		Role:  NewRole(t.Role),
	}
}

type UserInfo struct {
	ID    string
	Email string
	Role  Role
}
