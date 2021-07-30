package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

func humanizeValidationErrors(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "Invalid input."
	}
	return fmt.Sprintf("Invalid %q field.", errs[0].Field())
}

func queryParamInt(r *http.Request, key string) (int, error) {
	if r.Form == nil {
		_ = r.ParseForm()
	}

	v := r.FormValue(key)
	if v == "" {
		return 0, nil
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}

	return i, nil
}
