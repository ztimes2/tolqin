package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ztimes2/tolqin/app/api/internal/pkg/log"
)

type response struct {
	Data  interface{} `json:"data,omitempty"`
	Error interface{} `json:"error,omitempty"`
}

func write(w http.ResponseWriter, r *http.Request, statusCode int, resp response) {
	body, err := json.Marshal(resp)
	if err != nil {
		WriteUnexpectedError(w, r, err)
		return
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}

func writeData(w http.ResponseWriter, r *http.Request, statusCode int, v interface{}) {
	write(w, r, statusCode, response{Data: v})
}

func writeError(w http.ResponseWriter, r *http.Request, statusCode int, v interface{}) {
	write(w, r, statusCode, response{Error: v})
}

func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, errCode, errDesc string) {
	writeError(w, r, statusCode, newErrorResponse(errCode, errDesc))
}

func WriteUnexpectedError(w http.ResponseWriter, r *http.Request, err error) {
	if logger := log.FromContext(r.Context()); logger != nil {
		logger.WithError(err).Errorf("unexpected error: %s", err)
	}

	body, _ := json.Marshal(response{
		Error: newErrorResponse("unexpected", "Something went wrong..."),
	})

	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write(body)
}

func WriteNoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func WriteOK(w http.ResponseWriter, r *http.Request, v interface{}) {
	writeData(w, r, http.StatusOK, v)
}

func WriteCreated(w http.ResponseWriter, r *http.Request, v interface{}) {
	writeData(w, r, http.StatusCreated, v)
}

func WriteValidationError(w http.ResponseWriter, r *http.Request, desc string) {
	writeError(w, r, http.StatusBadRequest, newValidationErrorResponse(desc))
}

func WriteFieldErrors(w http.ResponseWriter, r *http.Request, f *Fields) {
	writeError(w, r, http.StatusBadRequest, newFieldErrorResponse(f))
}

func WriteFieldError(w http.ResponseWriter, r *http.Request, key, reason string) {
	WriteFieldErrors(w, r, NewFields(Field{key: key, reason: reason}))
}

func WritePayloadError(w http.ResponseWriter, r *http.Request) {
	WriteValidationError(w, r, "Invalid payload.")
}

func WriteNotFoundError(w http.ResponseWriter, r *http.Request, desc string) {
	WriteError(w, r, http.StatusNotFound, "not_found", desc)
}

type errorResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func newErrorResponse(code, desc string) errorResponse {
	return errorResponse{
		Code:        code,
		Description: desc,
	}
}

type validationErrorResponse struct {
	errorResponse
	Fields []validationErrorResponseField `json:"fields"`
}

type validationErrorResponseField struct {
	Key    string `json:"key"`
	Reason string `json:"reason"`
}

func newValidationErrorResponse(desc string) validationErrorResponse {
	return validationErrorResponse{
		errorResponse: newErrorResponse("invalid_input", desc),
		Fields:        make([]validationErrorResponseField, 0),
	}
}

func newFieldErrorResponse(f *Fields) validationErrorResponse {
	resp := newValidationErrorResponse("Invalid input parameters.")

	for _, field := range f.fields {
		resp.Fields = append(resp.Fields, validationErrorResponseField{
			Key:    field.key,
			Reason: field.reason,
		})
	}

	return resp
}

type Field struct {
	key    string
	reason string
}

type Fields struct {
	fields []Field
}

func NewFields(f ...Field) *Fields {
	return &Fields{
		fields: f,
	}
}

func (f *Fields) Is(err, target error, key, reason string) {
	if !errors.Is(err, target) {
		return
	}
	f.fields = append(f.fields, Field{
		key:    key,
		reason: reason,
	})
}
