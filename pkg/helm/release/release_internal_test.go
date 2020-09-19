package release

import (
	"fmt"
	"testing"
)

func TestRetryOnEtcdErrorRetry(t *testing.T) {
	calls := 0

	if err := retryOnEtcdError(func() error {
		calls++

		return fmt.Errorf("etcdserver: foo")
	}); err == nil {
		t.Errorf("retry should return error if all attempts failed")
	}

	if calls < 2 {
		t.Errorf("function should be at least called twice on etcd error")
	}
}

func TestRetryOnEtcdErrorDifferentError(t *testing.T) {
	calls := 0

	if err := retryOnEtcdError(func() error {
		calls++

		return fmt.Errorf("err")
	}); err == nil {
		t.Errorf("retry should return error if all attempts failed")
	}

	if calls != 1 {
		t.Errorf("function should be called only once if the error returned is not etcd error")
	}
}

func TestRetryOnEtcdErrorNoError(t *testing.T) {
	calls := 0

	if err := retryOnEtcdError(func() error {
		calls++

		return nil
	}); err != nil {
		t.Errorf("retry should not return error, got: %v", err)
	}

	if calls != 1 {
		t.Errorf("function should be called only once if no error is returned")
	}
}

func TestRetryOnEtcdErrorTranscientError(t *testing.T) {
	calls := 0

	if err := retryOnEtcdError(func() error {
		calls++

		if calls == 1 {
			return fmt.Errorf("etcdserver: foo")
		}

		return nil
	}); err != nil {
		t.Errorf("retry should retry and not return error, got: %v", err)
	}

	if calls != 2 {
		t.Errorf("function should return when no error is returned")
	}
}
