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

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &KeboolaProvider{}
)

// KeboolaProvider is the provider implementation.
type KeboolaProvider struct{}

// New creates a new provider instance
func New() provider.Provider {
	return &KeboolaProvider{}
}

// Client wraps the Keboola Management API client and exposes services.
type Client struct {
	API *keboola.APIClient
}

// KeboolaProviderModel describes the provider data model.
type KeboolaProviderModel struct {
	URL   types.String `tfsdk:"url"`
	Token types.String `tfsdk:"token"`
}

func (p *KeboolaProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "keboola-management"
	resp.Version = "0.1.0"
}

func (p *KeboolaProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Interact with Keboola Management API.",
		MarkdownDescription: "The Keboola Management provider allows Terraform to manage Keboola resources through the [Management API](https://keboolamanagementapi.docs.apiary.io/).",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description:         "Keboola Management API URL.",
				MarkdownDescription: "The URL of the Keboola Management API. For example: `https://connection.keboola.com/manage`",
				Required:            true,
			},
			"token": schema.StringAttribute{
				Description:         "Keboola Management API Token.",
				MarkdownDescription: "The Management API token used for authentication. This is a sensitive value and should be handled securely.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *KeboolaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config KeboolaProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.URL.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"URL is required",
		)
		return
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Token is required",
		)
		return
	}

	// Create a new configuration for the Management API client
	apiConfig := keboola.NewConfiguration()
	apiConfig.Host = config.URL.ValueString()
	apiConfig.AddDefaultHeader("X-KBC-ManageApiToken", config.Token.ValueString())

	// Create the Management API client with the configured settings
	apiClient := keboola.NewAPIClient(apiConfig)

	// Verify the token
	_, _, err := apiClient.TokenVerificationAPI.TokenVerification(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to verify token",
			"An unexpected error occurred when verifying the token: "+err.Error(),
		)
		return
	}

	client := &Client{
		API: apiClient,
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *KeboolaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMaintainerResource,
	}
}

func (p *KeboolaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
