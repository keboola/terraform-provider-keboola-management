package keboola

import (
	"context"
	"os"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a new provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"keboola": providerserver.NewProtocol6WithError(New()),
}

func TestProvider_impl(t *testing.T) {
	var _ provider.Provider = &KeboolaProvider{}
}

func TestProvider(t *testing.T) {
	t.Run("schema validation", func(t *testing.T) {
		ctx := context.Background()
		p := &KeboolaProvider{}

		resp := &provider.SchemaResponse{}
		p.Schema(ctx, provider.SchemaRequest{}, resp)

		assert.NotNil(t, resp.Schema)
		assert.NotEmpty(t, resp.Schema.Attributes)
	})

	t.Run("configuration validation", func(t *testing.T) {
		if os.Getenv("KEBOOLA_API_URL") == "" {
			t.Skip("KEBOOLA_API_URL must be set for this test")
		}
		if os.Getenv("KEBOOLA_TOKEN") == "" {
			t.Skip("KEBOOLA_TOKEN must be set for this test")
		}

		ctx := context.Background()
		p := &KeboolaProvider{}

		configResp := &provider.ConfigureResponse{}
		p.Configure(ctx, provider.ConfigureRequest{}, configResp)

		assert.False(t, configResp.Diagnostics.HasError())
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("KEBOOLA_API_URL"); v == "" {
		t.Fatal("KEBOOLA_API_URL must be set for acceptance tests")
	}
	if v := os.Getenv("KEBOOLA_TOKEN"); v == "" {
		t.Fatal("KEBOOLA_TOKEN must be set for acceptance tests")
	}
}

func TestAccProvider_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("keboola_maintainer.test", "name", "test"),
				),
			},
		},
	})
}

const testAccProviderConfig_basic = `
provider "keboola" {
  api_url = "https://connection.keboola.com"
  token   = "your-token"
}

resource "keboola_maintainer" "test" {
  name = "test"
}
`

func TestBackendSnowflakeResource_schema(t *testing.T) {
	ctx := context.Background()
	r := NewBackendSnowflakeResource()

	resp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.Schema)

	attrs := resp.Schema.Attributes

	// Required attributes
	for _, name := range []string{"host", "warehouse", "region", "owner", "technical_owner"} {
		attr, ok := attrs[name]
		assert.True(t, ok, "attribute %s should exist", name)
		assert.True(t, attr.IsRequired(), "attribute %s should be required", name)
	}

	// Optional attributes
	for _, name := range []string{"username", "technical_owner_contact_emails", "use_dynamic_backends", "use_network_policies", "use_sso", "edition"} {
		attr, ok := attrs[name]
		assert.True(t, ok, "attribute %s should exist", name)
		assert.True(t, attr.IsOptional(), "attribute %s should be optional", name)
	}

	// Computed attributes
	for _, name := range []string{"id", "sql_template", "is_enabled", "user_public_key", "security_integration_key"} {
		attr, ok := attrs[name]
		assert.True(t, ok, "attribute %s should exist", name)
		assert.True(t, attr.IsComputed(), "attribute %s should be computed", name)
	}
}

func TestBackendSnowflakeResource_metadata(t *testing.T) {
	r := NewBackendSnowflakeResource()
	resp := &fwresource.MetadataResponse{}
	r.Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "keboola-management"}, resp)
	assert.Equal(t, "keboola-management_backend_snowflake", resp.TypeName)
}

func TestBackendSnowflakeActivateResource_schema(t *testing.T) {
	ctx := context.Background()
	r := NewBackendSnowflakeActivateResource()

	resp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.Schema)

	attrs := resp.Schema.Attributes

	// Required attributes
	attr, ok := attrs["backend_id"]
	assert.True(t, ok, "attribute backend_id should exist")
	assert.True(t, attr.IsRequired(), "attribute backend_id should be required")

	// Computed attributes
	for _, name := range []string{"id", "is_enabled"} {
		attr, ok := attrs[name]
		assert.True(t, ok, "attribute %s should exist", name)
		assert.True(t, attr.IsComputed(), "attribute %s should be computed", name)
	}
}

func TestBackendSnowflakeActivateResource_metadata(t *testing.T) {
	r := NewBackendSnowflakeActivateResource()
	resp := &fwresource.MetadataResponse{}
	r.Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "keboola-management"}, resp)
	assert.Equal(t, "keboola-management_backend_snowflake_activate", resp.TypeName)
}
