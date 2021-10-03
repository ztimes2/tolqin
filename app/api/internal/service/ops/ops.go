package ops

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ztimes2/tolqin/app/api/internal/auth"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/valerra"
	"github.com/ztimes2/tolqin/app/api/internal/valerrautil"
)

const (
	maxEmailLength    = 100
	maxPasswordLength = 50
)

var (
	ErrInvalidEmail    = errors.New("invalid e-mail")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidRole     = errors.New("invalid role")
)

type Service struct {
	passwordSalter passwordSalter
	passwordHasher passwordHasher
	userStore      UserStore
}

type UserStore interface {
	auth.UserWriter
}

type passwordSalter interface {
	SaltPassword(password string) (salted, salt string, err error)
}

type passwordHasher interface {
	HashPassword(password string) (string, error)
}

func NewService(
	ps *auth.PasswordSalter,
	ph *auth.PasswordHasher,
	us UserStore) *Service {

	return &Service{
		passwordSalter: ps,
		passwordHasher: ph,
		userStore:      us,
	}
}

func (s *Service) CreateUser(email, password string, role auth.Role) (auth.User, error) {
	email = strings.TrimSpace(email)

	v := valerra.New()
	v.IfFalse(valerrautil.IsEmail(email), ErrInvalidEmail)
	v.IfFalse(valerrautil.IsPassword(password), ErrInvalidPassword)
	v.IfFalse(valerrautil.IsRoleIn(role, auth.RoleAdmin), ErrInvalidRole)

	if err := v.Validate(); err != nil {
		return auth.User{}, err
	}

	salted, salt, err := s.passwordSalter.SaltPassword(password)
	if err != nil {
		return auth.User{}, fmt.Errorf("could not salt password: %w", err)
	}

	hash, err := s.passwordHasher.HashPassword(salted)
	if err != nil {
		return auth.User{}, fmt.Errorf("could not hash password: %w", err)
	}

	return s.userStore.CreateUser(auth.UserCreationEntry{
		Role:         role,
		Email:        email,
		PasswordHash: hash,
		PasswordSalt: salt,
	})
}
