package keboola

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/keboola/keboola-sdk-go/v2/pkg/keboola/management"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &projectInvitationResource{}
	_ resource.ResourceWithConfigure   = &projectInvitationResource{}
	_ resource.ResourceWithImportState = &projectInvitationResource{}
)

// NewProjectInvitationResource is a helper function to simplify provider implementation.
func NewProjectInvitationResource() resource.Resource {
	return &projectInvitationResource{}
}

// projectInvitationResource is the resource implementation.
type projectInvitationResource struct {
	client *Client
}

// projectInvitationResourceModel maps the resource schema data.
type projectInvitationResourceModel struct {
	ID                types.String `tfsdk:"id"`
	ProjectID         types.String `tfsdk:"project_id"`
	Email             types.String `tfsdk:"email"`
	Role              types.String `tfsdk:"role"`
	ExpirationSeconds types.Number `tfsdk:"expiration_seconds"`
	Reason            types.String `tfsdk:"reason"`
	Status            types.String `tfsdk:"status"` // Computed field to track invitation status
}

// Configure adds the provider configured client to the resource.
func (r *projectInvitationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

// Metadata returns the resource type name.
func (r *projectInvitationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_invitation"
}

// Schema defines the schema for the resource.
func (r *projectInvitationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Keboola project invitation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Project invitation ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "ID of the project to which the invitation is sent.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address of the invited user.",
				Required:    true,
			},
			"role": schema.StringAttribute{
				Description: "Role to assign to the invited user.",
				Required:    true,
			},
			"expiration_seconds": schema.NumberAttribute{
				Description: "After how many seconds the invitation and membership of a user will expire.",
				Optional:    true,
			},
			"reason": schema.StringAttribute{
				Description: "Reason for inviting user.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the invitation (e.g., 'pending', 'accepted', 'expired').",
				Computed:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *projectInvitationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectInvitationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the API request
	apiReq := management.InviteAUserToAProjectRequest{
		Email: plan.Email.ValueString(),
	}
	if !plan.Role.IsNull() && plan.Role.ValueString() != "" {
		role := plan.Role.ValueString()
		apiReq.Role = &role
	}
	if !plan.ExpirationSeconds.IsNull() {
		f64, _ := plan.ExpirationSeconds.ValueBigFloat().Float64()
		f32 := float32(f64)
		apiReq.ExpirationSeconds = &f32
	}
	if !plan.Reason.IsNull() && plan.Reason.ValueString() != "" {
		reason := plan.Reason.ValueString()
		apiReq.Reason = &reason
	}

	// Call the API to send the invitation
	apiResp, _, err := r.client.API.ProjectsAPI.InviteAUserToAProject(ctx, plan.ProjectID.ValueString()).InviteAUserToAProjectRequest(apiReq).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project invitation",
			"Could not create project invitation: "+err.Error(),
		)
		return
	}
	if apiResp == nil || apiResp.Id == nil {
		resp.Diagnostics.AddError(
			"Error creating project invitation",
			"API did not return invitation ID",
		)
		return
	}
	plan.ID = types.StringValue(fmt.Sprintf("%v", *apiResp.Id))
	plan.Status = types.StringValue("pending")

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *projectInvitationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectInvitationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the API to get invitation details
	apiResp, _, err := r.client.API.ProjectsAPI.ProjectInvitationDetail(ctx, state.ProjectID.ValueString(), state.ID.ValueString()).Execute()
	if err != nil {
		// Check if the error is a 404 (invitation not found)
		// This can happen when the invitation was accepted or expired
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
			// Mark the invitation as accepted and keep it in state
			// This prevents Terraform from trying to recreate the invitation
			state.Status = types.StringValue("accepted")
			diags = resp.State.Set(ctx, &state)
			resp.Diagnostics.Append(diags...)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading project invitation",
			"Could not read invitation: "+err.Error(),
		)
		return
	}
	if apiResp == nil || apiResp.Id == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(fmt.Sprintf("%v", *apiResp.Id))
	if apiResp.User != nil && apiResp.User.Email != nil {
		state.Email = types.StringValue(*apiResp.User.Email)
	}
	if apiResp.Role != nil {
		state.Role = types.StringValue(*apiResp.Role)
	}
	if apiResp.Reason != nil {
		state.Reason = types.StringValue(*apiResp.Reason)
	}
	// Set status to pending if the invitation exists
	state.Status = types.StringValue("pending")
	// ExpirationSeconds is not directly available, so leave as is (API may provide Expires as a timestamp)
	// If needed, parse apiResp.Expires and apiResp.Created to calculate seconds

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *projectInvitationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get the current state to check if the invitation was accepted
	var state projectInvitationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the invitation was accepted, don't allow updates
	if !state.Status.IsNull() && state.Status.ValueString() == "accepted" {
		resp.Diagnostics.AddWarning(
			"Cannot update accepted invitation",
			"The invitation was accepted and can no longer be modified.",
		)
		// Keep the existing state
		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Project invitations cannot be updated. This method is required by the interface.
	// For now, we'll just keep the existing state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *projectInvitationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectInvitationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.API.ProjectsAPI.CancelProjectInvitation(ctx, state.ProjectID.ValueString(), state.ID.ValueString()).Execute()
	if err != nil {
		// Check if the error is a 404 (invitation not found)
		// This can happen when the invitation was already accepted or expired
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
			// The invitation no longer exists, which is fine for deletion
			// Just log this as info and continue
			return
		}

		resp.Diagnostics.AddError(
			"Error deleting project invitation",
			"Could not delete invitation: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *projectInvitationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
