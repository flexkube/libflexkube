package controlplane

import (
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/types"
)

// validator is a helper struct for validating common fields between controlplane
// configuration structs.
type validator struct {
	Common     *Common       `json:"common,omitempty"`
	Host       *host.Host    `json:"host,omitempty"`
	Kubeconfig client.Config `json:"kubeconfig,omitempty"`
	YAML       interface{}   `json:"yaml,omitempty"`
}

// validate validates validator struct. If validateKubeconfig is false,
// then validation of this field will be skipped.
func (v validator) validate(validateKubeconfig bool) error {
	var errors util.ValidateErrors

	if !validateKubeconfig {
		pki, err := utiltest.GeneratePKIErr()
		if err != nil {
			return fmt.Errorf("generating fake PKI for validation: %w", err)
		}

		v.Kubeconfig = client.Config{
			Server:            "localhost",
			CACertificate:     types.Certificate(pki.Certificate),
			ClientCertificate: types.Certificate(pki.Certificate),
			ClientKey:         types.PrivateKey(pki.PrivateKey),
		}
	}

	b, err := yaml.Marshal(v)
	if err != nil {
		return append(errors, fmt.Errorf("marshaling kubeconfig: %w", err))
	}

	if err := yaml.Unmarshal(b, &v); err != nil {
		return append(errors, fmt.Errorf("unmarshaling kubeconfig: %w", err))
	}

	if v.Common == nil {
		errors = append(errors, fmt.Errorf("common certificates must not defined"))
	}

	if validateKubeconfig {
		if _, err := v.Kubeconfig.ToYAMLString(); err != nil {
			errors = append(errors, fmt.Errorf("invalid kubeconfig: %w", err))
		}
	}

	errors = append(errors, v.validateHost()...)

	return errors.Return()
}

func (v validator) validateHost() util.ValidateErrors {
	var errors util.ValidateErrors

	if v.Host == nil {
		errors = append(errors, fmt.Errorf("host must be defined"))
	}

	if v.Host != nil {
		if err := v.Host.Validate(); err != nil {
			errors = append(errors, fmt.Errorf("validating host configuration: %w", err))
		}
	}

	return errors
}
