package auth

import (
	"fmt"

	"github.com/ztimes2/tolqin/app/api/internal/auth"
)

type Service struct {
	passwordSalter passwordSalter
	passwordHasher passwordHasher
	tokener        tokener
	userStore      auth.UserStore
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
	t tokener,
	us auth.UserStore) *Service {

	return &Service{
		passwordSalter: ps,
		passwordHasher: ph,
		tokener:        t,
		userStore:      us,
	}
}

func (s *Service) Token(email, password string) (string, error) {
	// TODO sanitize and validate input

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
