package keboola

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/keboola/keboola-sdk-go/v2/pkg/keboola"
	"github.com/keboola/keboola-sdk-go/v2/pkg/keboola/management"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &projectTokenResource{}
	_ resource.ResourceWithConfigure   = &projectTokenResource{}
	_ resource.ResourceWithImportState = &projectTokenResource{}
)

// NewProjectTokenResource returns a new keboola_project_token resource.
func NewProjectTokenResource() resource.Resource {
	return &projectTokenResource{}
}

// projectTokenResource implements the keboola_project_token resource.
type projectTokenResource struct {
	client *Client
}

// projectTokenResourceModel maps the resource schema data.
type projectTokenResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	ProjectID             types.String `tfsdk:"project_id"`
	Description           types.String `tfsdk:"description"`
	CanManageBuckets      types.Bool   `tfsdk:"can_manage_buckets"`
	CanReadAllFileUploads types.Bool   `tfsdk:"can_read_all_file_uploads"`
	CanPurgeTrash         types.Bool   `tfsdk:"can_purge_trash"`
	ExpiresIn             types.Number `tfsdk:"expires_in"`
	BucketPermissions     types.Map    `tfsdk:"bucket_permissions"`
	ComponentAccess       types.List   `tfsdk:"component_access"`
	Token                 types.String `tfsdk:"token"`
}

// Configure adds the provider configured client to the resource.
func (r *projectTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

// Metadata returns the resource type name.
func (r *projectTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_token"
}

// Schema defines the schema for the resource.
func (r *projectTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Keboola project storage token. The token is a one-time secret and cannot be read after creation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Token ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "ID of the Keboola project.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // Changing project requires new token
				},
			},
			"description": schema.StringAttribute{
				Description: "Token description.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"can_manage_buckets": schema.BoolAttribute{
				Description: "Token has full permissions on tabular storage.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"can_read_all_file_uploads": schema.BoolAttribute{
				Description: "Token has full permissions to files staging.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"can_purge_trash": schema.BoolAttribute{
				Description: "Allows permanently removing deleted configurations.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"expires_in": schema.NumberAttribute{
				Description: "Token lifetime in seconds.",
				Optional:    true,
				PlanModifiers: []planmodifier.Number{
					numberplanmodifier.RequiresReplace(),
				},
			},
			"bucket_permissions": schema.MapAttribute{
				Description: "Map of bucket permissions, e.g., {\"in.c\": \"main: read\"}.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"component_access": schema.ListAttribute{
				Description: "List of component IDs to grant access for component configurations.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Description: "Token value.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create creates the storage token and sets the initial Terraform state.
func (r *projectTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan projectTokenResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API request body for token creation
	tokenBody := management.CreateStorageTokenRequest{
		Description: plan.Description.ValueString(),
	}
	if !plan.CanManageBuckets.IsNull() {
		canManageBuckets := plan.CanManageBuckets.ValueBool()
		tokenBody.CanManageBuckets = &canManageBuckets
	}
	if !plan.CanReadAllFileUploads.IsNull() {
		canReadAllFileUploads := plan.CanReadAllFileUploads.ValueBool()
		tokenBody.CanReadAllFileUploads = &canReadAllFileUploads
	}
	if !plan.CanPurgeTrash.IsNull() {
		canPurgeTrash := plan.CanPurgeTrash.ValueBool()
		tokenBody.CanPurgeTrash = &canPurgeTrash
	}
	if !plan.ExpiresIn.IsNull() {
		bigVal := plan.ExpiresIn.ValueBigFloat()
		f64, _ := bigVal.Float64()
		converted := float32(f64)
		tokenBody.ExpiresIn = &converted
	}
	if !plan.BucketPermissions.IsNull() && !plan.BucketPermissions.IsUnknown() {
		var perms map[string]string
		diags := plan.BucketPermissions.ElementsAs(ctx, &perms, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			permissions := management.NewCreateStorageTokenRequestBucketPermissions()
			permissions.SetInC(perms["in.c"])
			tokenBody.BucketPermissions = permissions
		}
	}
	if !plan.ComponentAccess.IsNull() && !plan.ComponentAccess.IsUnknown() {
		var access []string
		diags := plan.ComponentAccess.ElementsAs(ctx, &access, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			tokenBody.SetComponentAccess(access)
		}
	}

	// Create the storage token
	tokenResp, _, err := r.client.API.ProjectsAPI.CreateStorageToken(ctx, plan.ProjectID.ValueString()).CreateStorageTokenRequest(tokenBody).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage token", "Could not create storage token: "+err.Error())
		return
	}

	plan.ID = types.StringValue(*tokenResp.Id)
	plan.Token = types.StringValue(*tokenResp.Token)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state. The token is not available after creation.
func (r *projectTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state projectTokenResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update forces recreation on any change.
func (r *projectTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Any change requires recreation, so just call Create.
	r.Create(ctx, resource.CreateRequest{
		Plan: req.Plan,
	}, &resource.CreateResponse{
		State:       resp.State,
		Diagnostics: resp.Diagnostics,
	})
}

// Delete removes the resource from state. No API call is needed.
func (r *projectTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectTokenResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The token is required to authorize the deletion
	tokenID := state.ID.ValueString()
	if tokenID == "" {
		tflog.Info(ctx, "Token not found in state, skipping deletion (likely after refresh/import)")
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Info(ctx, "Creating authorized API client for token deletion")
	client, err := sdk.NewAuthorizedAPI(ctx, r.client.API.GetConfig().Host, state.Token.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating authorized API client",
			"Could not create authorized API client: "+err.Error(),
		)
		return
	}

	_, err = client.DeleteTokenRequest(tokenID).Send(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting storage token",
			"Could not delete storage token: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Storage token deleted successfully, removing resource from state")
	// Remove the resource from state after successful deletion
	resp.State.RemoveResource(ctx)
}

// ImportState imports an existing resource into Terraform.
func (r *projectTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID (token value, but not available after creation)
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
