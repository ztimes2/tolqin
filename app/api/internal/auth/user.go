package auth

import (
	"errors"
	"time"
)

type User struct {
	ID           string
	Email        string
	Name         string
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

var (
	ErrEmailAlreadyTaken = errors.New("email has already been taken")
)

type UserStore interface {
	UserByEmail(email string) (User, error)
	CreateUser(UserCreationEntry) (User, error)
}

type UserCreationEntry struct {
	Role         Role
	Email        string
	PasswordHash string
	PasswordSalt string
}
