// Code generated by Terraform CLI; DO NOT EDIT.

//go:generate terraform fmt -recursive ./examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate

package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/keboola/terraform-provider-keboola-management/keboola" // local provider package
)

// Provider documentation generation.
func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/keboola/keboola-management",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), keboola.New, opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
