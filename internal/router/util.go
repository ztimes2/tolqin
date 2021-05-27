package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/ztimes2/tolqin/internal/logging"
)

func write(w http.ResponseWriter, r *http.Request, statusCode int, v interface{}) {
	if v == nil {
		w.WriteHeader(statusCode)
		return
	}

	body, err := json.Marshal(v)
	if err != nil {
		writeUnexpectedError(w, r, err)
		return
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}

func writeUnexpectedError(w http.ResponseWriter, r *http.Request, err error) {
	if logger := logging.FromContext(r.Context()); logger != nil {
		logger.WithError(err).Errorf("unexpected error: %s", err)
	}

	body, _ := json.Marshal(errorResponse{
		Description: "Something went wrong...",
	})

	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write(body)
}

func writeError(w http.ResponseWriter, r *http.Request, statusCode int, desc string) {
	write(w, r, statusCode, errorResponse{
		Description: desc,
	})
}

type errorResponse struct {
	Description string `json:"error_description"`
}

type listResponse struct {
	Items []interface{} `json:"items"`
}

func toListResponse(items []interface{}) listResponse {
	if len(items) == 0 {
		items = make([]interface{}, 0)
	}
	return listResponse{
		Items: items,
	}
}

func humanizeValidationErrors(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "Invalid input."
	}
	return fmt.Sprintf("Invalid %q field.", errs[0].Field())
}
