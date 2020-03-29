package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConfigFilesMarshal(t *testing.T) {
	f := map[string]string{
		"/foo": "bar",
	}

	e := map[string]interface{}{
		"/foo": "bar",
	}

	if diff := cmp.Diff(configFilesMarshal(f, false), e); diff != "" {
		t.Fatalf("Unexpected diff: %s", diff)
	}
}

func TestConfigFilesMarshalSensitive(t *testing.T) {
	f := map[string]string{
		"/foo": "bar",
	}

	e := map[string]interface{}{
		"/foo": sha256sum([]byte("bar")),
	}

	if diff := cmp.Diff(configFilesMarshal(f, true), e); diff != "" {
		t.Fatalf("Unexpected diff: %s", diff)
	}
}

func TestConfigFilesMarshalDontChecksumEmpty(t *testing.T) {
	f := map[string]string{
		"/foo": "",
	}

	e := map[string]interface{}{
		"/foo": "",
	}

	if diff := cmp.Diff(configFilesMarshal(f, true), e); diff != "" {
		t.Fatalf("Unexpected diff: %s", diff)
	}
}

func TestConfigFilesUnmarshal(t *testing.T) {
	i := map[string]interface{}{
		"/foo": "bar",
	}

	e := map[string]string{
		"/foo": "bar",
	}

	if diff := cmp.Diff(configFilesUnmarshal(i), e); diff != "" {
		t.Fatalf("Unexpected diff: %s", diff)
	}
}

func TestConfigFilesUnmarshalEmpty(t *testing.T) {
	e := map[string]string{}

	if diff := cmp.Diff(configFilesUnmarshal(nil), e); diff != "" {
		t.Fatalf("Unexpected diff: %s", diff)
	}
}
