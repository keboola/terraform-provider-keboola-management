package keboola

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	keboola "github.com/keboola/keboola-sdk-go/v2/pkg/keboola/management"
)

// Provider returns the terraform resource provider for Keboola.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Keboola Management API URL.",
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Keboola Management API Token.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"keboola_maintainer": resourceMaintainer(), // Register maintainer resource
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// Client wraps the Keboola Management API client and exposes services.
type Client struct {
	API *keboola.APIClient
}

// providerConfigure initializes the Keboola Management API client using the SDK.
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	url := d.Get("url").(string)
	token := d.Get("token").(string)

	// Create a new configuration for the Management API client
	config := keboola.NewConfiguration()
	config.Host = url                                      // Set the Management API URL
	config.AddDefaultHeader("X-KBC-ManageApiToken", token) // Set the API token as a default header

	// Create the Management API client with the configured settings
	apiClient := keboola.NewAPIClient(config)

	_, _, err := apiClient.TokenVerificationAPI.TokenVerification(ctx).Execute()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return &Client{API: apiClient}, nil
}
