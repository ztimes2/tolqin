package auth

import (
	"errors"
	"strings"
	"time"
)

type User struct {
	ID           string
	Email        string
	CreatedAt    time.Time
	Role         Role
	PasswordHash string
	PasswordSalt string
}

const (
	roleNameAdmin = "admin"
)

type Role int

const (
	RoleUndefined Role = iota
	RoleAdmin
)

func NewRole(role string) Role {
	switch role {
	case roleNameAdmin:
		return RoleAdmin
	default:
		return RoleUndefined
	}
}

func (r Role) String() string {
	switch r {
	case RoleAdmin:
		return roleNameAdmin
	default:
		return ""
	}
}

func (r Role) MarshalJSON() ([]byte, error) {
	return []byte(`"` + r.String() + `"`), nil
}

func (r *Role) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	*r = NewRole(s)
	return nil
}

var (
	ErrEmailAlreadyTaken = errors.New("email has already been taken")
	ErrUserNotFound      = errors.New("user not found")
)

type UserReader interface {
	UserByEmail(email string) (User, error)
}

type UserWriter interface {
	CreateUser(UserCreationEntry) (User, error)
}

type UserCreationEntry struct {
	Role         Role
	Email        string
	PasswordHash string
	PasswordSalt string
}
