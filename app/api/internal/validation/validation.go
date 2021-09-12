package validation

import "fmt"

type ErrorOptionFn func(*Error)

func WithDescription(s string) ErrorOptionFn {
	return func(e *Error) {
		e.desc = s
	}
}

type Error struct {
	field string
	desc  string
}

func NewError(field string, opts ...ErrorOptionFn) *Error {
	e := &Error{
		field: field,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Error) Error() string {
	return fmt.Sprintf("invalid field: %q", e.field)
}

func (e *Error) Field() string {
	return e.field
}

func (e *Error) Description() string {
	if e.desc == "" {
		return fmt.Sprintf("Invalid %s.", e.field)
	}
	return e.desc
}
