package util

import (
	"fmt"
	"testing"
)

func TestValidateErrors(t *testing.T) {
	t.Parallel()

	err := func() error {
		errors := ValidateErrors{
			fmt.Errorf("first error"),
		}

		errors = append(errors, fmt.Errorf("second error"))

		return errors.Return()
	}

	if err == nil {
		t.Fatalf("Error shouldn't be nil")
	}
}
