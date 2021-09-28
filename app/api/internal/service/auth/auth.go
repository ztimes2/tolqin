package auth

import (
	"fmt"

	"github.com/ztimes2/tolqin/app/api/internal/auth"
)

type passwordSalter interface {
	SaltPassword(password string) (salted, salt string, err error)
}

type passwordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(hash, password string) error
}

type Service struct {
	passwordSalter passwordSalter
	passwordHasher passwordHasher
	tokener        auth.Tokener
	userStore      auth.UserStore
}

func NewService(
	passwordSalter *auth.PasswordSalter,
	passwordHasher *auth.PasswordHasher,
	userStore auth.UserStore) *Service {

	return &Service{
		passwordSalter: passwordSalter,
		passwordHasher: passwordHasher,
		userStore:      userStore,
	}
}

func (s *Service) Token(email, password string) (string, error) {
	// TODO sanitize and validate input

	user, err := s.userStore.UserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("could not find user: %w", err)
	}

	if err := s.passwordHasher.ComparePassword(
		user.PasswordHash,
		password,
		user.PasswordSalt,
	); err != nil {
		return "", fmt.Errorf("could not compare password: %w", err)
	}

	token, err := s.tokener.Token(user)
	if err != nil {
		return "", fmt.Errorf("could not generate token: %w", err)
	}

	return token, nil
}
