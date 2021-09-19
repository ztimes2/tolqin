package valerra

import "fmt"

// Validator can be used for validating multiple conditions and catching errors
// for those that fail.
type Validator struct {
	conditions []conditionFn
}

type conditionFn func() error

// New returns a new *Validator.
func New() *Validator {
	return &Validator{}
}

// IfFalse registers the given condition function to return the given error in
// case if the condition returns false.
func (v *Validator) IfFalse(condition func() bool, err error) {
	v.conditions = append(v.conditions, func() error {
		if !condition() {
			return err
		}
		return nil
	})
}

// Validate validates all the conditions registered to the validator and returns
// errors for the failed once as *Errors. The function returns nil if all conditions
// succeed.
func (v *Validator) Validate() error {
	var errs []error
	for _, fn := range v.conditions {
		if err := fn(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return NewErrors(errs...)
	}

	return nil
}

// IfFalse validates a single condition and immediately returns an error as *Errors
// in case if the condition returns false. The function returns nil if the condition
// returns true.
func IfFalse(condition func() bool, err error) error {
	v := New()
	v.IfFalse(condition, err)
	return v.Validate()
}

// Errors holds multiple errors.
type Errors struct {
	errs []error
}

// NewErrors wraps the given errors as *Errors.
func NewErrors(err ...error) *Errors {
	return &Errors{
		errs: err,
	}
}

// Errors returns underlying errors as slice.
func (e *Errors) Errors() []error {
	return e.errs
}

// Error implements the error interface.
func (e *Errors) Error() string {
	if len(e.errs) == 1 {
		return "1 rule failed validation"
	}
	return fmt.Sprintf("%d rules failed validation", len(e.errs))
}
