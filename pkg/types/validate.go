package types

import (
	"strings"
)

// ValidateError is a collection of errors, which can be used when
// performing validation of structs to collect all possible errors
// and return them in one batch.
type ValidateError []error

func (e ValidateError) Error() string {
	errors := []string{}

	for _, s := range e {
		errors = append(errors, s.Error())
	}

	return strings.Join(errors, ", ")
}

// Return returns nil, if no errors has been added.
func (e ValidateError) Return() error {
	if len(e) > 0 {
		return e
	}

	return nil
}
