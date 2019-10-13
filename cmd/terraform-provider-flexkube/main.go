package main

import (
	"github.com/hashicorp/terraform/plugin"

	"github.com/invidian/flexkube/cmd/terraform-provider-flexkube/flexkube"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: flexkube.Provider})
}
