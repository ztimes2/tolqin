package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ztimes2/tolqin/app/api/internal/pkg/logging"
)

func Write(w http.ResponseWriter, r *http.Request, statusCode int, v interface{}) {
	if v == nil {
		w.WriteHeader(statusCode)
		return
	}

	body, err := json.Marshal(v)
	if err != nil {
		WriteUnexpectedError(w, r, err)
		return
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}

func WriteUnexpectedError(w http.ResponseWriter, r *http.Request, err error) {
	if logger := logging.FromContext(r.Context()); logger != nil {
		logger.WithError(err).Errorf("unexpected error: %s", err)
	}

	body, _ := json.Marshal(NewErrorResponse("unexpected", "Something went wrong..."))

	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write(body)
}

func WriteNoContent(w http.ResponseWriter, r *http.Request) {
	Write(w, r, http.StatusNoContent, nil)
}

func WriteOK(w http.ResponseWriter, r *http.Request, v interface{}) {
	Write(w, r, http.StatusOK, v)
}

func WriteCreated(w http.ResponseWriter, r *http.Request, v interface{}) {
	Write(w, r, http.StatusCreated, v)
}

func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, errCode, errDesc string) {
	Write(w, r, statusCode, NewErrorResponse(errCode, errDesc))
}

func WriteValidationError(w http.ResponseWriter, r *http.Request, desc string) {
	Write(w, r, http.StatusBadRequest, NewValidationErrorResponse(desc))
}

func WriteFieldErrors(w http.ResponseWriter, r *http.Request, f *Fields) {
	Write(w, r, http.StatusBadRequest, NewFieldErrorResponse(f))
}

func WriteFieldError(w http.ResponseWriter, r *http.Request, key, reason string) {
	WriteFieldErrors(w, r, NewFields(Field{
		key:    key,
		reason: reason,
	}))
}

func WritePayloadError(w http.ResponseWriter, r *http.Request) {
	WriteValidationError(w, r, "Invalid payload.")
}

func WriteNotFoundError(w http.ResponseWriter, r *http.Request, desc string) {
	WriteError(w, r, http.StatusNotFound, "not_found", desc)
}

type ErrorResponse struct {
	Error ErrorResponseMeta `json:"error"`
}

type ErrorResponseMeta struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func NewErrorResponse(code, desc string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorResponseMeta{
			Code:        code,
			Description: desc,
		},
	}
}

type ValidationErrorResponse struct {
	Error ValidationErrorResponseMeta `json:"error"`
}

type ValidationErrorResponseMeta struct {
	ErrorResponseMeta
	Fields []ValidationErrorResponseField `json:"fields"`
}

type ValidationErrorResponseField struct {
	Key    string `json:"key"`
	Reason string `json:"reason"`
}

func NewValidationErrorResponse(desc string) ValidationErrorResponse {
	return ValidationErrorResponse{
		Error: ValidationErrorResponseMeta{
			ErrorResponseMeta: ErrorResponseMeta{
				Code:        "invalid_input",
				Description: desc,
			},
			Fields: make([]ValidationErrorResponseField, 0),
		},
	}
}

func NewFieldErrorResponse(f *Fields) ValidationErrorResponse {
	resp := NewValidationErrorResponse("Invalid input parameters.")

	for _, field := range f.fields {
		resp.Error.Fields = append(resp.Error.Fields, ValidationErrorResponseField{
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
