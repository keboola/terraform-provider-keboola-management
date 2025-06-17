package keboola

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a new provider server instance.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"keboola": providerserver.NewProtocol6WithError(New("test")()),
}

// TestAccProtoV6ProviderFactories returns the provider factories for use in other test files
func TestAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return testAccProtoV6ProviderFactories
}

// TestProviderSchema verifies the provider schema is correct.
func TestProviderSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	providerFunc := New("test")
	p := providerFunc()

	resp := &provider.SchemaResponse{}
	p.Schema(ctx, provider.SchemaRequest{}, resp)

	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "host")
	assert.Contains(t, resp.Schema.Attributes, "token")
}

// TestProviderConfigure verifies the provider can be configured with minimal settings.
func TestProviderConfigure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	providerFunc := New("test")
	p := providerFunc()

	// Use real Manage API token from environment if available
	token := os.Getenv("KBC_MANAGE_API_TOKEN")
	if token == "" {
		token = "test-token" // fallback for local/dev testing
	}

	resp := &provider.ConfigureResponse{}
	p.Configure(ctx, provider.ConfigureRequest{
		Config: testProviderConfig(t, map[string]string{
			"host":  "connection.keboola.com",
			"token": token,
		}),
	}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.ResourceData)
	assert.NotNil(t, resp.DataSourceData)
}

// testProviderConfig is a helper to create provider configuration for testing
func testProviderConfig(t *testing.T, values map[string]string) tfsdk.Config {
	t.Helper()

	// Create a data value using the schema
	val := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"host":  tftypes.String,
			"token": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"host":  tftypes.NewValue(tftypes.String, values["host"]),
		"token": tftypes.NewValue(tftypes.String, values["token"]),
	})

	// Create the config with the value
	cfg := tfsdk.Config{
		Raw: val,
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"host": schema.StringAttribute{
					Required: true,
				},
				"token": schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	return cfg
}
