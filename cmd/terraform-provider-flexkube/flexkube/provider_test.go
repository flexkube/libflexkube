package flexkube

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("Provider internal validation should succeed, got: %v", err)
	}
}
