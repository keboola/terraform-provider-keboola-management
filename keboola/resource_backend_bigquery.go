package keboola

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/keboola/keboola-sdk-go/v2/pkg/keboola/management"
)

// This resource only supports Create and Update. Read and Delete are not supported by the API.
// The resource will not be refreshed or deleted by Terraform. Document this clearly to users.

var (
	_ resource.Resource              = &backendBigQueryResource{}
	_ resource.ResourceWithConfigure = &backendBigQueryResource{}
)

// NewBackendBigQueryResource returns a new BigQuery backend resource instance.
func NewBackendBigQueryResource() resource.Resource {
	return &backendBigQueryResource{}
}

// backendBigQueryResource implements the Terraform resource for BigQuery backend registration.
type backendBigQueryResource struct {
	client *Client
}

type backendBigQueryCredentialsModel struct {
	Type                    types.String `tfsdk:"type"`
	ProjectID               types.String `tfsdk:"project_id"`
	PrivateKeyID            types.String `tfsdk:"private_key_id"`
	PrivateKey              types.String `tfsdk:"private_key"`
	ClientEmail             types.String `tfsdk:"client_email"`
	ClientID                types.String `tfsdk:"client_id"`
	AuthURI                 types.String `tfsdk:"auth_uri"`
	TokenURI                types.String `tfsdk:"token_uri"`
	AuthProviderX509CertURL types.String `tfsdk:"auth_provider_x509_cert_url"`
	ClientX509CertURL       types.String `tfsdk:"client_x509_cert_url"`
}

type backendBigQueryResourceModel struct {
	ID          types.String                     `tfsdk:"id"`
	Owner       types.String                     `tfsdk:"owner"`
	FolderId    types.String                     `tfsdk:"folder_id"`
	Region      types.String                     `tfsdk:"region"`
	Credentials *backendBigQueryCredentialsModel `tfsdk:"credentials"`
}

func (r *backendBigQueryResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *backendBigQueryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend_bigquery"
}

func (r *backendBigQueryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Registers and updates a BigQuery backend. Only Create and Update are supported. Read and Delete are not supported by the API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Backend ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner": schema.StringAttribute{
				Description: "GCP account owner.",
				Required:    true,
			},
			"folder_id": schema.StringAttribute{
				Description: "GCP folder ID where the service account is created.",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "BigQuery region.",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"credentials": schema.SingleNestedBlock{
				Description: "Service account credentials for BigQuery backend.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Credential type.",
						Required:    true,
					},
					"project_id": schema.StringAttribute{
						Description: "GCP project ID.",
						Required:    true,
					},
					"private_key_id": schema.StringAttribute{
						Description: "Private key ID.",
						Required:    true,
					},
					"private_key": schema.StringAttribute{
						Description: "Private key.",
						Required:    true,
						Sensitive:   true,
					},
					"client_email": schema.StringAttribute{
						Description: "Client email.",
						Required:    true,
					},
					"client_id": schema.StringAttribute{
						Description: "Client ID.",
						Required:    true,
					},
					"auth_uri": schema.StringAttribute{
						Description: "Auth URI.",
						Required:    true,
					},
					"token_uri": schema.StringAttribute{
						Description: "Token URI.",
						Required:    true,
					},
					"auth_provider_x509_cert_url": schema.StringAttribute{
						Description: "Auth provider x509 cert URL.",
						Required:    true,
					},
					"client_x509_cert_url": schema.StringAttribute{
						Description: "Client x509 cert URL.",
						Required:    true,
					},
				},
			},
		},
	}
}

// Create registers a new BigQuery backend.
func (r *backendBigQueryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan backendBigQueryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var creds *management.CreateNewGoogleCloudStorageRequestGcsCredentials
	if plan.Credentials != nil {
		creds = &management.CreateNewGoogleCloudStorageRequestGcsCredentials{
			Type:                    plan.Credentials.Type.ValueString(),
			ProjectId:               plan.Credentials.ProjectID.ValueString(),
			PrivateKeyId:            plan.Credentials.PrivateKeyID.ValueString(),
			PrivateKey:              plan.Credentials.PrivateKey.ValueString(),
			ClientEmail:             plan.Credentials.ClientEmail.ValueString(),
			ClientId:                plan.Credentials.ClientID.ValueString(),
			AuthUri:                 plan.Credentials.AuthURI.ValueString(),
			TokenUri:                plan.Credentials.TokenURI.ValueString(),
			AuthProviderX509CertUrl: plan.Credentials.AuthProviderX509CertURL.ValueString(),
			ClientX509CertUrl:       plan.Credentials.ClientX509CertURL.ValueString(),
		}
	}

	apiReq := management.CreateANewBigQueryBackendRequest{
		Owner:       plan.Owner.ValueString(),
		FolderId:    plan.FolderId.ValueString(),
		Region:      plan.Region.ValueString(),
		Credentials: creds,
	}

	apiResp, _, err := r.client.API.SUPERStorageBackendsManagementAPI.CreateANewBigQueryBackend(ctx).CreateANewBigQueryBackendRequest(apiReq).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating BigQuery backend",
			fmt.Sprintf("Could not create BigQuery backend: %s", err.Error()),
		)
		return
	}
	plan.ID = types.StringValue(fmt.Sprintf("%v", apiResp.Id))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read is not supported by the API. Warn and do nothing.
func (r *backendBigQueryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddWarning("Read not supported", "BigQuery backend does not support read operation. State will not be refreshed.")
}

// Update updates the BigQuery backend.
func (r *backendBigQueryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan backendBigQueryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var creds *management.GCPCredentials
	if plan.Credentials != nil {
		creds = &management.GCPCredentials{
			Type:                    plan.Credentials.Type.ValueString(),
			ProjectId:               plan.Credentials.ProjectID.ValueString(),
			PrivateKeyId:            plan.Credentials.PrivateKeyID.ValueString(),
			PrivateKey:              plan.Credentials.PrivateKey.ValueString(),
			ClientEmail:             plan.Credentials.ClientEmail.ValueString(),
			ClientId:                plan.Credentials.ClientID.ValueString(),
			AuthUri:                 plan.Credentials.AuthURI.ValueString(),
			TokenUri:                plan.Credentials.TokenURI.ValueString(),
			AuthProviderX509CertUrl: plan.Credentials.AuthProviderX509CertURL.ValueString(),
			ClientX509CertUrl:       plan.Credentials.ClientX509CertURL.ValueString(),
		}
	}

	apiReq := management.BigQueryStorageBackendUpdate{
		Credentials: creds,
	}

	_, _, err := r.client.API.SUPERStorageBackendsManagementAPI.UpdateBigQueryBackend(ctx, plan.ID.ValueString()).BigQueryStorageBackendUpdate(apiReq).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating BigQuery backend",
			fmt.Sprintf("Could not update BigQuery backend '%s': %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete is not supported by the API. Warn and do nothing.
func (r *backendBigQueryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Delete not supported", "BigQuery backend does not support delete operation. Resource will remain in state.")
}
