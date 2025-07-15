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

// GCP Cloud Storage File Storage resource (Create and Read only)
var (
	_ resource.Resource              = &fileStorageGCSResource{}
	_ resource.ResourceWithConfigure = &fileStorageGCSResource{}
)

func NewFileStorageGCSResource() resource.Resource {
	return &fileStorageGCSResource{}
}

type fileStorageGCSResource struct {
	client *Client
}

type fileStorageGCSResourceModel struct {
	ID             types.String                    `tfsdk:"id"`
	FilesBucket    types.String                    `tfsdk:"files_bucket"`
	Owner          types.String                    `tfsdk:"owner"`
	Region         types.String                    `tfsdk:"region"`
	GcsCredentials *fileStorageGCSCredentialsModel `tfsdk:"gcs_credentials"`
}

type fileStorageGCSCredentialsModel struct {
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

func (r *fileStorageGCSResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *fileStorageGCSResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_storage_gcs"
}

func (r *fileStorageGCSResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages GCP Cloud Storage file storage. Only Create and Read are supported.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Storage ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"files_bucket": schema.StringAttribute{
				Description: "GCS bucket name.",
				Required:    true,
			},
			"owner": schema.StringAttribute{
				Description: "Associated GCP account owner.",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "GCP region.",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"gcs_credentials": schema.SingleNestedBlock{
				Description: "Service account credentials for GCS storage.",
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

func (r *fileStorageGCSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fileStorageGCSResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var creds *management.CreateNewGoogleCloudStorageRequestGcsCredentials
	if plan.GcsCredentials != nil {
		creds = &management.CreateNewGoogleCloudStorageRequestGcsCredentials{
			Type:                    plan.GcsCredentials.Type.ValueString(),
			ProjectId:               plan.GcsCredentials.ProjectID.ValueString(),
			PrivateKeyId:            plan.GcsCredentials.PrivateKeyID.ValueString(),
			PrivateKey:              plan.GcsCredentials.PrivateKey.ValueString(),
			ClientEmail:             plan.GcsCredentials.ClientEmail.ValueString(),
			ClientId:                plan.GcsCredentials.ClientID.ValueString(),
			AuthUri:                 plan.GcsCredentials.AuthURI.ValueString(),
			TokenUri:                plan.GcsCredentials.TokenURI.ValueString(),
			AuthProviderX509CertUrl: plan.GcsCredentials.AuthProviderX509CertURL.ValueString(),
			ClientX509CertUrl:       plan.GcsCredentials.ClientX509CertURL.ValueString(),
		}
	}

	apiReq := management.CreateNewGoogleCloudStorageRequest{
		FilesBucket:    plan.FilesBucket.ValueString(),
		Owner:          plan.Owner.ValueString(),
		Region:         plan.Region.ValueString(),
		GcsCredentials: creds,
	}

	apiResp, _, err := r.client.API.SUPERFileStorageManagementAPI.CreateNewGoogleCloudStorage(ctx).CreateNewGoogleCloudStorageRequest(apiReq).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating GCS file storage",
			fmt.Sprintf("Could not create GCS file storage: %s", err.Error()),
		)
		return
	}
	plan.ID = types.StringValue(fmt.Sprintf("%v", apiResp.Id))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *fileStorageGCSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fileStorageGCSResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call ListGoogleCloudStorage to get all GCS storages
	storages, _, err := r.client.API.SUPERFileStorageManagementAPI.ListGoogleCloudStorage(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing GCS file storages",
			fmt.Sprintf("Could not list GCS file storages: %s", err.Error()),
		)
		return
	}

	found := false
	for _, s := range storages {
		storage, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		id, ok := storage["id"].(float64)
		if !ok {
			continue
		}
		if state.ID.ValueString() == fmt.Sprintf("%v", id) {
			// Map fields
			if v, ok := storage["filesBucket"].(string); ok {
				state.FilesBucket = types.StringValue(v)
			}
			if v, ok := storage["region"].(string); ok {
				state.Region = types.StringValue(v)
			}
			if v, ok := storage["owner"].(string); ok {
				state.Owner = types.StringValue(v)
			}
			if creds, ok := storage["gcsCredentials"].(map[string]interface{}); ok {
				credModel := &fileStorageGCSCredentialsModel{}
				if v, ok := creds["type"].(string); ok {
					credModel.Type = types.StringValue(v)
				}
				if v, ok := creds["projectId"].(string); ok {
					credModel.ProjectID = types.StringValue(v)
				}
				if v, ok := creds["privateKeyId"].(string); ok {
					credModel.PrivateKeyID = types.StringValue(v)
				}
				if v, ok := creds["privateKey"].(string); ok {
					credModel.PrivateKey = types.StringValue(v)
				}
				if v, ok := creds["clientEmail"].(string); ok {
					credModel.ClientEmail = types.StringValue(v)
				}
				if v, ok := creds["clientId"].(string); ok {
					credModel.ClientID = types.StringValue(v)
				}
				if v, ok := creds["authUri"].(string); ok {
					credModel.AuthURI = types.StringValue(v)
				}
				if v, ok := creds["tokenUri"].(string); ok {
					credModel.TokenURI = types.StringValue(v)
				}
				if v, ok := creds["authProviderX509CertUrl"].(string); ok {
					credModel.AuthProviderX509CertURL = types.StringValue(v)
				}
				if v, ok := creds["clientX509CertUrl"].(string); ok {
					credModel.ClientX509CertURL = types.StringValue(v)
				}
				state.GcsCredentials = credModel
			}
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Delete is a no-op. Deletion is not supported by the API.
func (r *fileStorageGCSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Delete not supported", "Deletion of GCS file storage is not supported by the Keboola API.")
}

// Update is a no-op. Update is not supported by the API.
func (r *fileStorageGCSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Update not supported", "Update of GCS file storage is not supported by the Keboola API.")
}
