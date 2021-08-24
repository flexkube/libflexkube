package util

import (
	"strings"
)

// ValidateErrors is a collection of errors, which can be used when
// performing validation of structs to collect all possible errors
// and return them in one batch.
type ValidateErrors []error

func (e ValidateErrors) Error() string {
	errors := []string{}

	for _, s := range e {
		errors = append(errors, s.Error())
	}

	return strings.Join(errors, ", ")
}

// Return returns nil, if no errors has been added.
func (e ValidateErrors) Return() error {
	if len(e) > 0 {
		return e
	}

	return nil
}
