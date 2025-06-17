package keboola

import (
	"context"
	"os"
	"testing"

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
