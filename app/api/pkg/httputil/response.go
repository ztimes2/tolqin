package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ztimes2/tolqin/app/api/pkg/log"
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

func writeData(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	write(w, r, statusCode, response{Data: data})
}

func writeError(w http.ResponseWriter, r *http.Request, statusCode int, errResp interface{}) {
	write(w, r, statusCode, response{Error: errResp})
}

// WriteError writes an error to the response using the given HTTP status code,
// error code, and error description.
func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, errCode, errDesc string) {
	writeError(w, r, statusCode, newErrorResponse(errCode, errDesc))
}

// WriteUnexpectedError writes a 500 Internal Server Error HTTP status code and
// an error using 'unexpected' error code and the static unexpected error description
// to the response. The given error gets additionally logged.
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

// WriteNoContent writes a 204 No Content HTTP status code to the response.
func WriteNoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// WriteOK writes a 200 OK HTTP status code and the given data to the response.
func WriteOK(w http.ResponseWriter, r *http.Request, data interface{}) {
	writeData(w, r, http.StatusOK, data)
}

// WriteCreated writes a 201 Created HTTP status code and the given data to the
// response.
func WriteCreated(w http.ResponseWriter, r *http.Request, data interface{}) {
	writeData(w, r, http.StatusCreated, data)
}

// WriteValidationError writes a 400 Bad Request HTTP status code and an error
// using 'invalid_input' error code and the given description to the response.
func WriteValidationError(w http.ResponseWriter, r *http.Request, desc string) {
	writeError(w, r, http.StatusBadRequest, newValidationErrorResponse(desc))
}

// WriteFieldErrors writes a 400 Bad Request HTTP status code and an error using
// 'invalid_input' error code, the static invalid input error description, and
// the given invalid fields to the response.
func WriteFieldErrors(w http.ResponseWriter, r *http.Request, f *InvalidFields) {
	writeError(w, r, http.StatusBadRequest, newFieldErrorResponse(f))
}

// WriteFieldError writes a 400 Bad Request HTTP status code and an error using
// 'invalid_input' error code, the static invalid parameters error description,
// and the given invalid field to the response.
func WriteFieldError(w http.ResponseWriter, r *http.Request, f InvalidField) {
	WriteFieldErrors(w, r, &InvalidFields{
		fields: []InvalidField{f},
	})
}

// WritePayloadError writes a 400 Bad Request HTTP status code and an error using
// 'invalid_input' error code and the static invalid payload error description.
func WritePayloadError(w http.ResponseWriter, r *http.Request) {
	WriteValidationError(w, r, "Invalid payload.")
}

// WriteNotFoundError writes a 404 Not Found HTTP status code and an error using
// 'not_found' error code and the given error description to the response.
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

func newFieldErrorResponse(f *InvalidFields) validationErrorResponse {
	resp := newValidationErrorResponse("Invalid input parameters.")

	for _, field := range f.fields {
		resp.Fields = append(resp.Fields, validationErrorResponseField(field))
	}

	return resp
}

// InvalidField holds details of an invalid field.
type InvalidField struct {
	Key    string
	Reason string
}

// NewInvalidField returns InvalidField using the given key and reason.
func NewInvalidField(key, reason string) InvalidField {
	return InvalidField{
		Key:    key,
		Reason: reason,
	}
}

// InvalidFields holds multiple invalid fields.
type InvalidFields struct {
	fields []InvalidField
}

// NewInvalidFields returns a new *InvalidFields.
func NewInvalidFields() *InvalidFields {
	return &InvalidFields{}
}

// Is adds the given field to the invalid fields if at least one of errors in the
// given err's chain matches the target.
func (f *InvalidFields) Is(err, target error, field InvalidField) {
	if !errors.Is(err, target) {
		return
	}
	f.fields = append(f.fields, field)
}
