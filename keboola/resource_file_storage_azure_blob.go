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

// Azure Blob Storage File Storage resource (Create and Read only)
var (
	_ resource.Resource              = &fileStorageAzureBlobResource{}
	_ resource.ResourceWithConfigure = &fileStorageAzureBlobResource{}
)

func NewFileStorageAzureBlobResource() resource.Resource {
	return &fileStorageAzureBlobResource{}
}

type fileStorageAzureBlobResource struct {
	client *Client
}

type fileStorageAzureBlobResourceModel struct {
	ID            types.String `tfsdk:"id"`
	AccountName   types.String `tfsdk:"account_name"`
	AccountKey    types.String `tfsdk:"account_key"`
	Owner         types.String `tfsdk:"owner"`
	ContainerName types.String `tfsdk:"container_name"`
}

func (r *fileStorageAzureBlobResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *fileStorageAzureBlobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_storage_azure_blob"
}

func (r *fileStorageAzureBlobResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Azure Blob Storage file storage. Only Create and Read are supported.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Storage ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_name": schema.StringAttribute{
				Description: "Azure storage account name.",
				Required:    true,
			},
			"account_key": schema.StringAttribute{
				Description: "Azure storage account key.",
				Required:    true,
				Sensitive:   true,
			},
			"owner": schema.StringAttribute{
				Description: "Associated Azure account owner.",
				Required:    true,
			},
			"container_name": schema.StringAttribute{
				Description: "Azure Blob container name.",
				Optional:    true,
			},
		},
	}
}

func (r *fileStorageAzureBlobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fileStorageAzureBlobResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := management.CreateNewAzureBlobStorageRequest{
		AccountName: plan.AccountName.ValueString(),
		Owner:       plan.Owner.ValueString(),
		AccountKey:  plan.AccountKey.ValueString(),
	}
	if !plan.ContainerName.IsNull() && plan.ContainerName.ValueString() != "" {
		container := plan.ContainerName.ValueString()
		apiReq.ContainerName = &container
	}

	apiResp, _, err := r.client.API.SUPERFileStorageManagementAPI.CreateNewAzureBlobStorage(ctx).CreateNewAzureBlobStorageRequest(apiReq).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Azure Blob file storage",
			fmt.Sprintf("Could not create Azure Blob file storage: %s", err.Error()),
		)
		return
	}
	plan.ID = types.StringValue(fmt.Sprintf("%v", apiResp.Id))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *fileStorageAzureBlobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fileStorageAzureBlobResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call ListAzureBlobStorage to get all Azure Blob storages
	storages, _, err := r.client.API.SUPERFileStorageManagementAPI.ListAzureBlobStorage(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing Azure Blob file storages",
			fmt.Sprintf("Could not list Azure Blob file storages: %s", err.Error()),
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
			if v, ok := storage["accountName"].(string); ok {
				state.AccountName = types.StringValue(v)
			}
			if v, ok := storage["accountKey"].(string); ok {
				state.AccountKey = types.StringValue(v)
			}
			if v, ok := storage["owner"].(string); ok {
				state.Owner = types.StringValue(v)
			}
			if v, ok := storage["containerName"].(string); ok {
				state.ContainerName = types.StringValue(v)
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
func (r *fileStorageAzureBlobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Delete not supported", "Deletion of Azure Blob file storage is not supported by the Keboola API.")
}

// Update is a no-op. Update is not supported by the API.
func (r *fileStorageAzureBlobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Update not supported", "Update of Azure Blob file storage is not supported by the Keboola API.")
}
