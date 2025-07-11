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

// AWS S3 File Storage resource (Create and Read only)
var (
	_ resource.Resource              = &fileStorageS3Resource{}
	_ resource.ResourceWithConfigure = &fileStorageS3Resource{}
)

func NewFileStorageS3Resource() resource.Resource {
	return &fileStorageS3Resource{}
}

type fileStorageS3Resource struct {
	client *Client
}

type fileStorageS3ResourceModel struct {
	ID          types.String `tfsdk:"id"`
	AwsKey      types.String `tfsdk:"aws_key"`
	AwsSecret   types.String `tfsdk:"aws_secret"`
	FilesBucket types.String `tfsdk:"files_bucket"`
	Region      types.String `tfsdk:"region"`
	Owner       types.String `tfsdk:"owner"`
}

func (r *fileStorageS3Resource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *fileStorageS3Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_storage_s3"
}

func (r *fileStorageS3Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages AWS S3 file storage. Only Create and Read are supported.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Storage ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aws_key": schema.StringAttribute{
				Description: "AWS access key.",
				Required:    true,
			},
			"aws_secret": schema.StringAttribute{
				Description: "AWS secret key.",
				Required:    true,
				Sensitive:   true,
			},
			"files_bucket": schema.StringAttribute{
				Description: "S3 bucket name.",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "AWS region.",
				Required:    true,
			},
			"owner": schema.StringAttribute{
				Description: "Associated AWS account owner.",
				Required:    true,
			},
		},
	}
}

func (r *fileStorageS3Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fileStorageS3ResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := management.CreateNewAWSS3StorageRequest{
		AwsKey:      plan.AwsKey.ValueString(),
		AwsSecret:   plan.AwsSecret.ValueString(),
		FilesBucket: plan.FilesBucket.ValueString(),
		Region:      plan.Region.ValueString(),
		Owner:       plan.Owner.ValueString(),
	}

	apiResp, _, err := r.client.API.SUPERFileStorageManagementAPI.CreateNewAWSS3Storage(ctx).CreateNewAWSS3StorageRequest(apiReq).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating AWS S3 file storage",
			fmt.Sprintf("Could not create AWS S3 file storage: %s", err.Error()),
		)
		return
	}
	plan.ID = types.StringValue(fmt.Sprintf("%v", apiResp.Id))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *fileStorageS3Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fileStorageS3ResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call ListStorages to get all S3 storages
	storages, _, err := r.client.API.SUPERFileStorageManagementAPI.ListStorages(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing S3 file storages",
			fmt.Sprintf("Could not list S3 file storages: %s", err.Error()),
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
			if v, ok := storage["awsKey"].(string); ok {
				state.AwsKey = types.StringValue(v)
			}
			if v, ok := storage["awsSecret"].(string); ok {
				state.AwsSecret = types.StringValue(v)
			}
			if v, ok := storage["filesBucket"].(string); ok {
				state.FilesBucket = types.StringValue(v)
			}
			if v, ok := storage["region"].(string); ok {
				state.Region = types.StringValue(v)
			}
			if v, ok := storage["owner"].(string); ok {
				state.Owner = types.StringValue(v)
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
func (r *fileStorageS3Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Delete not supported", "Deletion of AWS S3 file storage is not supported by the Keboola API.")
}

// Update is a no-op. Update is not supported by the API.
func (r *fileStorageS3Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Update not supported", "Update of AWS S3 file storage is not supported by the Keboola API.")
}
