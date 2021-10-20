package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ztimes2/tolqin/app/api/internal/pkg/auth"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/jwt"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/valerrautil"
	"github.com/ztimes2/tolqin/app/api/pkg/valerra"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	passwordSalter passwordSalter
	passwordHasher passwordHasher
	jwtEncoder     jwtEncoder
	userStore      UserStore
}

type UserStore interface {
	auth.UserReader
}

type passwordSalter interface {
	SaltPassword(password, salt string) string
}

type passwordHasher interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(hash, password string) error
}

type jwtEncoder interface {
	EncodeJWT(auth.User) (string, error)
}

func NewService(
	ps *auth.PasswordSalter,
	ph *auth.PasswordHasher,
	j *jwt.EncodeDecoder,
	us UserStore) *Service {

	return &Service{
		passwordSalter: ps,
		passwordHasher: ph,
		jwtEncoder:     j,
		userStore:      us,
	}
}

func (s *Service) Token(email, password string) (string, error) {
	email = strings.TrimSpace(email)

	v := valerra.New()
	v.IfFalse(valerrautil.IsEmail(email), ErrInvalidCredentials)
	v.IfFalse(valerrautil.IsPassword(password), ErrInvalidCredentials)

	if err := v.Validate(); err != nil {
		return "", err
	}

	user, err := s.userStore.UserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("could not find user: %w", err)
	}

	salted := s.passwordSalter.SaltPassword(password, user.PasswordSalt)

	if err := s.passwordHasher.CompareHashAndPassword(user.PasswordHash, salted); err != nil {
		return "", fmt.Errorf("could not compare password: %w", err)
	}

	token, err := s.jwtEncoder.EncodeJWT(user)
	if err != nil {
		return "", fmt.Errorf("could not encode jwt: %w", err)
	}

	return token, nil
}
