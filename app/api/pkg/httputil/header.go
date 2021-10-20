package httputil

import (
	"errors"
	"net/http"
	"strings"
)

const (
	headerAuth = "Authorization"

	authHeaderSchemeBearer = "Bearer"
)

func BearerAuthHeader(r *http.Request) (string, error) {
	values := strings.Split(r.Header.Get(headerAuth), " ")

	if len(values) != 2 {
		return "", errors.New("invalid format")
	}

	if values[0] != authHeaderSchemeBearer {
		return "", errors.New("unexpected scheme")
	}

	if values[1] == "" {
		return "", errors.New("empty credentials")
	}

	return values[1], nil
}
