package router

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ztimes2/tolqin/app/api/internal/auth"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/httputil"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/valerra"
)

type authService interface {
	Token(email, password string) (string, error)
}

type authHandler struct {
	service authService
}

func newAuthHandler(s authService) *authHandler {
	return &authHandler{
		service: s,
	}
}

func (h *authHandler) token(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httputil.WritePayloadError(w, r)
		return
	}

	token, err := h.service.Token(payload.Email, payload.Password)
	if err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			httputil.WriteValidationError(w, r, "Invalid credentials.")
			return
		}

		if errors.Is(err, auth.ErrUserNotFound) {
			httputil.WriteValidationError(w, r, "Invalid credentials.")
			return
		}

		httputil.WriteUnexpectedError(w, r, err)
		return
	}

	httputil.WriteOK(w, r, tokenResponse{
		AccessToken: token,
	})
}
