package keboola

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	keboola "github.com/keboola/keboola-sdk-go/v2/pkg/keboola/management"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &keboolaProvider{version: "dev"}
)

// keboolaProvider is the provider implementation.
type keboolaProvider struct {
	version string
}

// keboolaProviderModel maps provider schema data to a Go type.
type keboolaProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

// New creates a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &keboolaProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *keboolaProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "keboola-management"
	resp.Version = p.version
}

// Provider returns the terraform resource provider for Keboola.
func (p *keboolaProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Interact with Keboola Management API (https://keboolamanagementapi.docs.apiary.io/).",
		MarkdownDescription: "Interact with Keboola Management API (https://keboolamanagementapi.docs.apiary.io/).",
		Blocks:              map[string]schema.Block{},
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "Keboola Management API Host.",
			},
			"token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Keboola Management API Token. This is a secret value and should be stored securely.",
			},
		},
	}
}

// Client wraps the Keboola Management API client and exposes services.
type Client struct {
	API *keboola.APIClient
}

// providerConfigure initializes the Keboola Management API client using the SDK.
func (p *keboolaProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse) {
	var cfg keboolaProviderModel
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new configuration for the Management API client
	config := keboola.NewConfiguration()
	config.Host = cfg.Host.ValueString()                                     // Set the Management API URL
	config.AddDefaultHeader("X-KBC-ManageApiToken", cfg.Token.ValueString()) // Set the API token as a default header

	// Create the Management API client with the configured settings
	apiClient := keboola.NewAPIClient(config)

	_, _, err := apiClient.TokenVerificationAPI.TokenVerification(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to verify token", err.Error())
		return
	}

	resp.DataSourceData = &Client{API: apiClient}
	resp.ResourceData = &Client{API: apiClient}
}

// DataSources defines the data sources implemented by the provider.
func (p *keboolaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// Resources defines the resources implemented by the provider.
func (p *keboolaProvider) Resources(_ context.Context) []func() resource.Resource {
	sResource := func() resource.Resource {
		return NewMaintainerResource()
	}

	return []func() resource.Resource{
		sResource,
	}
}
