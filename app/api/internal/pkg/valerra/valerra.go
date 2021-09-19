/*
Package valerra helps with validation of one or multiple conditions and catching
errors for those that fail. It is suitable for user input or form validation that
needs to catch all errors at once instead of returning only the first one.

For example, the following piece of code checks 3 conditions, maps them to specific
error types and returns errors for 2 conditions that fail validation:

	var (
		a = ""
		b = "bee"
		c = 123
	)
	v := valerra.New()
	v.IfFalse(valerra.StringNotEmpty(a), ErrInvalidA)
	v.IfFalse(valerra.StringLessOrEqual(b, 1), InvalidBError{B: b})
	v.IfFalse(func() bool {return c == 123}, errors.New("invalid c"))

	if err := v.Validate(); err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			for _, e := range vErr.Errors() {
				fmt.Println(e)
			}
		}
	}

	// Output:
	// invalid a
	// invalid b

Alternatively, if there is only one condition that needs to be checked, then the
following shorter approach can be used:

		valerra.IfFalse(valerra.StringNotEmpty(""), errors.New("invalid input")))
		if err := v.Validate(); err != nil {
			var vErr *valerra.Errors
			if errors.As(err, &vErr) {
				for _, e := range vErr.Errors() {
					fmt.Println(e)
				}
			}
		}

		// Output:
		// invalid input

The package is built around the Condition function primitive for checking validation
and comes with a set of trivial conditions out of the box. Additionally, it can support
any custom Condition for more specific use cases.
*/
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
func (v *Validator) IfFalse(c Condition, err error) {
	v.conditions = append(v.conditions, func() error {
		if !c() {
			return err
		}
		return nil
	})
}

// Condition is a function used for validation.
type Condition func() bool

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
func IfFalse(c Condition, err error) error {
	v := New()
	v.IfFalse(c, err)
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
