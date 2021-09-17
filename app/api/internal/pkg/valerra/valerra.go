package valerra

import "fmt"

type Validator struct {
	conditions []conditionFn
}

type conditionFn func() error

func New() *Validator {
	return &Validator{}
}

func (v *Validator) IfFalse(condition func() bool, err error) {
	v.conditions = append(v.conditions, func() error {
		if !condition() {
			return err
		}
		return nil
	})
}

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

func IfFalse(condition func() bool, err error) error {
	v := New()
	v.IfFalse(condition, err)
	return v.Validate()
}

type Errors struct {
	errs []error
}

func NewErrors(err ...error) *Errors {
	return &Errors{
		errs: err,
	}
}

func (e *Errors) Errors() []error {
	return e.errs
}

func (e *Errors) Error() string {
	if len(e.errs) == 1 {
		return "1 rule failed validation"
	}
	return fmt.Sprintf("%d rules failed validation", len(e.errs))
}
