package validation

import "fmt"

type Error struct {
	Field string
}

func NewError(field string) *Error {
	return &Error{
		Field: field,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("invalid field: %s", e.Field)
}
