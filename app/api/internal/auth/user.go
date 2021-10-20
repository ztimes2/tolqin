package auth

import (
	"errors"
	"time"
)

// User represents the application's user.
type User struct {
	ID           string
	Email        string
	CreatedAt    time.Time
	Role         Role
	PasswordHash string
	PasswordSalt string
}

// Role represents a user role.
type Role int

// User roles supported by the application.
const (
	RoleUndefined Role = iota
	RoleAdmin
)

var (
	// ErrEmailAlreadyTaken is used when an e-mail address has already been taken
	// by an existing user.
	ErrEmailAlreadyTaken = errors.New("email has already been taken")

	// ErrUserNotFound is used when a user could not be found.
	ErrUserNotFound = errors.New("user not found")
)

// UserReader is a data storage from which users can be read.
type UserReader interface {
	// UserByEmail finds and returns a user by the given e-mail address.
	//
	// ErrUserNotFound is returned when a user could not be found.
	UserByEmail(email string) (User, error)
}

// UserWriter is a data storage containing users against which write operations
// can be performed.
type UserWriter interface {
	// CreateUser creates a new user using the given entry and returns it if the
	// creation succeeds.
	//
	// ErrEmailAlreadyTaken is returned when the provided e-mail address has already
	// been taken by another existing user.
	CreateUser(UserCreationEntry) (User, error)
}

// UserCreationEntry holds parameters for creating a new user in a data storage.
type UserCreationEntry struct {
	Role         Role
	Email        string
	PasswordHash string
	PasswordSalt string
}
