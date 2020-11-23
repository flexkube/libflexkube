package util

import (
	"fmt"
	"testing"
)

func TestValidateError(t *testing.T) {
	t.Parallel()

	err := func() error {
		errors := ValidateError{
			fmt.Errorf("first error"),
		}

		errors = append(errors, fmt.Errorf("second error"))

		return errors.Return()
	}

	if err == nil {
		t.Fatalf("Error shouldn't be nil")
	}
}
