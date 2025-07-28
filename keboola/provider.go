package keboola

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	keboola "github.com/keboola/keboola-sdk-go/v2/pkg/keboola/management"
)

// Environment variable names for configuration
const (
	KbcHostnameSuffix = "KBC_HOSTNAME_SUFFIX"
	KbcToken          = "KBC_MANAGE_TOKEN"
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
	HostnameSuffix types.String `tfsdk:"hostname_suffix"`
	Token          types.String `tfsdk:"token"`
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
			"hostname_suffix": schema.StringAttribute{
				Description:         "The hostname suffix for the Keboola Domain. For example: `keboola.com`. Can also be set via KBC_HOSTNAME_SUFFIX environment variable.",
				MarkdownDescription: "The hostname suffix for the Keboola Domain e.g `keboola.com`. The provider will construct the full URL as `https://connection.{hostname_suffix}`. Can also be set via KBC_HOSTNAME_SUFFIX environment variable.",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				Description:         "Keboola Management API Token. Can also be set via KBC_MANAGE_TOKEN environment variable.",
				MarkdownDescription: "The Management API token used for authentication. This is a sensitive value and should be handled securely. Can also be set via KBC_MANAGE_TOKEN environment variable.",
				Optional:            true,
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

	// Initialize variables to hold the final values
	var hostnameSuffix, token string

	// Check hostname suffix - use config if not unknown and not null, otherwise check environment
	if config.HostnameSuffix.IsUnknown() || config.HostnameSuffix.IsNull() {
		hostnameSuffix = os.Getenv(KbcHostnameSuffix) //nolint: forbidigo
	} else {
		hostnameSuffix = config.HostnameSuffix.ValueString()
	}

	// Check token - use config if not unknown and not null, otherwise check environment
	if config.Token.IsUnknown() || config.Token.IsNull() {
		token = os.Getenv(KbcToken) //nolint: forbidigo
	} else {
		token = config.Token.ValueString()
	}
	// Validate that we have the required values
	if hostnameSuffix == "" {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Hostname suffix is required. Set it in the provider configuration or via KBC_HOSTNAME_SUFFIX environment variable.",
		)
		return
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Token is required. Set it in the provider configuration or via KBC_MANAGE_TOKEN environment variable.",
		)
		return
	}

	// Construct the Management API URL from the hostname suffix
	apiURL := "connection." + hostnameSuffix

	// Create a new configuration for the Management API client
	apiConfig := keboola.NewConfiguration()
	apiConfig.Host = apiURL
	apiConfig.AddDefaultHeader("X-KBC-ManageApiToken", token)

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
		NewOrganizationResource,         // Register the organization resource
		NewProjectResource,              // Register the project resource
		NewProjectTokenResource,         // Register the project token resource
		NewProjectInvitationResource,    // Register the project invitation resource
		NewProjectFeatureResource,       // Register the project feature resource
		NewBackendResource,              // Register the backend resource
		NewBackendBigQueryResource,      // Register the BigQuery backend resource
		NewFileStorageS3Resource,        // Register the S3 file storage resource
		NewFileStorageGCSResource,       // Register the GCS file storage resource
		NewFileStorageAzureBlobResource, // Register the Azure Blob file storage resource
	}
}

func (p *KeboolaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
