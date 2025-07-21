package keboola

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/keboola/keboola-sdk-go/v2/pkg/keboola/management"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

// NewProjectResource is a helper function to simplify provider implementation.
func NewProjectResource() resource.Resource {
	return &projectResource{}
}

// projectResource is the resource implementation.
type projectResource struct {
	client *Client
}

// projectResourceModel maps the resource schema data.
type projectResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	OrganizationID          types.String `tfsdk:"organization_id"`
	Type                    types.String `tfsdk:"type"`
	DefaultBackend          types.String `tfsdk:"default_backend"`
	DataRetentionTimeInDays types.String `tfsdk:"data_retention_time_in_days"`
	StorageToken            types.String `tfsdk:"storage_token"`
	Token                   *tokenModel  `tfsdk:"token"`
	// Add more fields as needed for project creation and management
}

// tokenModel represents the nested token block for storage token creation.
type tokenModel struct {
	Description           types.String `tfsdk:"description"`
	CanManageBuckets      types.Bool   `tfsdk:"can_manage_buckets"`
	CanReadAllFileUploads types.Bool   `tfsdk:"can_read_all_file_uploads"`
	CanPurgeTrash         types.Bool   `tfsdk:"can_purge_trash"`
	ExpiresIn             types.Number `tfsdk:"expires_in"`
	BucketPermissions     types.Map    `tfsdk:"bucket_permissions"`
	ComponentAccess       types.List   `tfsdk:"component_access"`
}

// Configure adds the provider configured client to the resource.
func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

// Metadata returns the resource type name.
func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the resource.
func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Keboola project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Project ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Project name.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "ID of the organization to which the project belongs.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Project type: one of production, poc, demo; default is production.",
				Required:    true,
			},
			"default_backend": schema.StringAttribute{
				Description: "Project default backend: snowflake or redshift; default is snowflake.",
				Optional:    true,
			},
			"data_retention_time_in_days": schema.StringAttribute{
				Description: "Data retention in days for Time Travel.",
				Optional:    true,
			},
			// Add more attributes as needed
			"storage_token": schema.StringAttribute{
				Description: "Storage token created for the project. Sensitive, only available after creation. Not available after refresh/import.",
				Computed:    true,
				Sensitive:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"token": schema.SingleNestedBlock{
				Description: "Optional block to define the storage token properties for the project.",
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						Description: "Token description.",
						Optional:    true,
					},
					"can_manage_buckets": schema.BoolAttribute{
						Description: "Token has full permissions on tabular storage. Defaults to true. Set to false to disable.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
					},
					"can_read_all_file_uploads": schema.BoolAttribute{
						Description: "Token has full permissions to files staging. Defaults to true. Set to false to disable.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
					},
					"can_purge_trash": schema.BoolAttribute{
						Description: "Allows permanently removing deleted configurations. Defaults to true. Set to false to disable.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
					},
					"expires_in": schema.NumberAttribute{
						Description: "Token lifetime in seconds.",
						Optional:    true,
					},
					"bucket_permissions": schema.MapAttribute{
						Description: "Map of bucket permissions, e.g., {\"in.c\": \"main: read\"}.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"component_access": schema.ListAttribute{
						Description: "List of component IDs to grant access for component configurations.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API request body for project creation
	name := plan.Name.ValueString()
	typeVal := plan.Type.ValueString()
	body := management.AddAProjectRequest{
		Name: name,    // required, plain string
		Type: typeVal, // required, plain string
	}
	if !plan.DefaultBackend.IsNull() && plan.DefaultBackend.ValueString() != "" {
		backend := plan.DefaultBackend.ValueString()
		body.DefaultBackend = &backend // pointer to string
	}
	if !plan.DataRetentionTimeInDays.IsNull() && plan.DataRetentionTimeInDays.ValueString() != "" {
		days := plan.DataRetentionTimeInDays.ValueString()
		body.DataRetentionTimeInDays = &days // pointer to string
	}

	apiResp, _, err := r.client.API.ProjectsAPI.AddAProject(ctx, plan.OrganizationID.ValueString()).AddAProjectRequest(body).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project",
			"Could not create project, unexpected error: "+err.Error(),
		)
		return
	}
	if apiResp == nil || apiResp.Id == nil {
		resp.Diagnostics.AddError(
			"Error creating project",
			"API did not return project ID",
		)
		return
	}
	plan.ID = types.StringValue(fmt.Sprintf("%v", *apiResp.Id))

	// Only create a storage token if the optional token block is provided.
	if plan.Token != nil {
		var tokenBody management.CreateStorageTokenRequest
		tokenBody = management.CreateStorageTokenRequest{
			Description: plan.Token.Description.ValueString(),
		}
		// Always set boolean fields if not null (default is true)
		if !plan.Token.CanManageBuckets.IsNull() {
			canManageBuckets := plan.Token.CanManageBuckets.ValueBool()
			tokenBody.CanManageBuckets = &canManageBuckets
		}
		if !plan.Token.CanReadAllFileUploads.IsNull() {
			canReadAllFileUploads := plan.Token.CanReadAllFileUploads.ValueBool()
			tokenBody.CanReadAllFileUploads = &canReadAllFileUploads
		}
		if !plan.Token.CanPurgeTrash.IsNull() {
			canPurgeTrash := plan.Token.CanPurgeTrash.ValueBool()
			tokenBody.CanPurgeTrash = &canPurgeTrash
		}
		if !plan.Token.ExpiresIn.IsNull() {
			// Convert Terraform number to float32 pointer
			bigVal := plan.Token.ExpiresIn.ValueBigFloat()
			f64, _ := bigVal.Float64()
			converted := float32(f64)
			tokenBody.ExpiresIn = &converted
		}
		if !plan.Token.BucketPermissions.IsNull() && !plan.Token.BucketPermissions.IsUnknown() {
			var perms map[string]string
			diags := plan.Token.BucketPermissions.ElementsAs(ctx, &perms, false)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				permissions := management.NewCreateStorageTokenRequestBucketPermissions()
				permissions.SetInC(perms["in.c"])
				tokenBody.BucketPermissions = permissions
			}
		}
		if !plan.Token.ComponentAccess.IsNull() && !plan.Token.ComponentAccess.IsUnknown() {
			var access []string
			diags := plan.Token.ComponentAccess.ElementsAs(ctx, &access, false)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				tokenBody.SetComponentAccess(access)
			}
		}

		// NOTE: The storage token is only available after creation and will NOT be available after refresh/import.
		// This is a one-time secret. Document this clearly for users.
		tokenResp, _, err := r.client.API.ProjectsAPI.CreateStorageToken(ctx, plan.ID.ValueString()).CreateStorageTokenRequest(tokenBody).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Error creating storage token", "Could not create storage token: "+err.Error())
			return
		}

		plan.StorageToken = types.StringValue(*tokenResp.Token)
	} else {
		// If the token block is not provided, we must explicitly set storage_token to null.
		// This avoids Terraform's 'unknown after apply' error for computed attributes.
		plan.StorageToken = types.StringNull()
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed project value from API
	apiResp, _, err := r.client.API.ProjectsAPI.ProjectDetail(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading project",
			"Could not read project ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	if apiResp == nil || apiResp.Id == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(fmt.Sprintf("%v", *apiResp.Id))
	if apiResp.Name != nil {
		state.Name = types.StringValue(*apiResp.Name)
	}
	// Add more fields as needed

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API request body
	body := management.UpdateAProjectRequest{}
	if !plan.Name.IsNull() {
		name := plan.Name.ValueString()
		body.Name = &name
	}
	// Add more fields as needed

	// Update existing project
	_, _, err := r.client.API.ProjectsAPI.UpdateAProject(ctx, plan.ID.ValueString()).UpdateAProjectRequest(body).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating project",
			"Could not update project, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch updated state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing project
	_, err := r.client.API.ProjectsAPI.DeleteAProject(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting project",
			"Could not delete project, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
