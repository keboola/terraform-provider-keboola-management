package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/keboola/terraform-provider-keboola-management/keboola" // local provider package
)

// main is the entry point for the Terraform provider plugin.
func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: keboola.Provider,
	})
}
