package auth

import (
	"github.com/dgrijalva/jwt-go"
)

type Salter interface {
	Salt() (string, error)
}

type PasswordHasher interface {
	HashPassword(password, salt string) (string, error)
	ComparePassword(hash, password, salt string) error
}

type Tokener interface {
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
