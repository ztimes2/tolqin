package auth

import (
	"fmt"

	"github.com/ztimes2/tolqin/app/api/internal/auth"
)

type Service struct {
	salter         auth.Salter
	passwordHasher auth.PasswordHasher
	tokener        auth.Tokener
	userStore      auth.UserStore
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


