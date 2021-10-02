package auth

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
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	passwordSalter passwordSalter
	passwordHasher passwordHasher
	tokener        tokener
	userStore      UserStore
}

type UserStore interface {
	auth.UserReader
}

type passwordSalter interface {
	SaltPassword(password string) (salted, salt string, err error)
}

type passwordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(hash, password string) error
}

type tokener interface {
	Token(auth.User) (string, error)
	ParseTokenClaims(token string) (auth.TokenClaims, error)
}

func NewService(
	ps *auth.PasswordSalter,
	ph *auth.PasswordHasher,
	t *auth.Tokener,
	us UserStore) *Service {

	return &Service{
		passwordSalter: ps,
		passwordHasher: ph,
		tokener:        t,
		userStore:      us,
	}
}

func (s *Service) Token(email, password string) (string, error) {
	email = strings.TrimSpace(email)

	v := valerra.New()
	v.IfFalse(valerra.StringLessOrEqual(email, maxEmailLength), ErrInvalidCredentials)
	v.IfFalse(valerrautil.IsEmail(email), ErrInvalidCredentials)
	v.IfFalse(valerra.StringLessOrEqual(password, maxEmailLength), ErrInvalidCredentials)

	if err := v.Validate(); err != nil {
		return "", err
	}

	user, err := s.userStore.UserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("could not find user: %w", err)
	}

	if err := s.passwordHasher.ComparePassword(user.PasswordHash, password); err != nil {
		return "", fmt.Errorf("could not compare password: %w", err)
	}

	token, err := s.tokener.Token(user)
	if err != nil {
		return "", fmt.Errorf("could not generate token: %w", err)
	}

	return token, nil
}
