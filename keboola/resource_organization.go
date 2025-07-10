package keboola

import (
	"context"
	"fmt"
	"strconv"

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
	_ resource.Resource                = &organizationResource{}
	_ resource.ResourceWithConfigure   = &organizationResource{}
	_ resource.ResourceWithImportState = &organizationResource{}
)

// NewOrganizationResource is a helper function to simplify provider implementation.
func NewOrganizationResource() resource.Resource {
	return &organizationResource{}
}

// organizationResource is the resource implementation.
type organizationResource struct {
	client *Client
}

// organizationResourceModel maps the resource schema data.
type organizationResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	MaintainerID            types.String `tfsdk:"maintainer_id"`
	AllowAutoJoin           types.String `tfsdk:"allow_auto_join"`
	CrmID                   types.String `tfsdk:"crm_id"`
	ActivityCenterProjectID types.String `tfsdk:"activity_center_project_id"`
	MfaRequired             types.String `tfsdk:"mfa_required"`
}

// Configure adds the provider configured client to the resource.
func (r *organizationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

// Metadata returns the resource type name.
func (r *organizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

// Schema defines the schema for the resource.
func (r *organizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Keboola organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Organization ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Organization name.",
				Required:    true,
			},
			"maintainer_id": schema.StringAttribute{
				Description: "Assign the organization to another maintainer.",
				Required:    true,
			},
			"allow_auto_join": schema.StringAttribute{
				Description: "Set whether superAdmins need approval to join the organization's projects (default true).",
				Optional:    true,
			},
			"crm_id": schema.StringAttribute{
				Description: "Set CRM ID. Only maintainer members and superadmins can change this.",
				Optional:    true,
			},
			"activity_center_project_id": schema.StringAttribute{
				Description: "Set ActivityCenter ProjectId. Only maintainer members and superadmins can change this.",
				Optional:    true,
			},
			"mfa_required": schema.StringAttribute{
				Description: "Toggle whether all members of or organization and its projects must have enabled multi-factor authentication (default false).",
				Optional:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan organizationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse maintainer_id for API call
	maintainerID, err := strconv.Atoi(plan.MaintainerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting maintainer_id",
			"Could not convert maintainer_id to integer: "+err.Error(),
		)
		return
	}

	// Build API request body
	name := plan.Name.ValueString()
	body := management.CreateAnOrganizationRequest{
		Name: &name,
	}
	if !plan.CrmID.IsNull() && plan.CrmID.ValueString() != "" {
		crmID := plan.CrmID.ValueString()
		body.CrmId = &crmID
	}

	// Create new organization under the specified maintainer
	apiResp, _, err := r.client.API.OrganizationsAPI.CreateAnOrganization(ctx, float32(maintainerID)).CreateAnOrganizationRequest(body).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating organization",
			"Could not create organization, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	if apiResp.Id == nil {
		resp.Diagnostics.AddError(
			"Error creating organization",
			"API did not return organization ID",
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%v", int(*apiResp.Id)))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state organizationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed organization value from API
	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting ID",
			"Could not convert ID to integer: "+err.Error(),
		)
		return
	}

	orgID := float32(id)
	apiResp, _, err := r.client.API.OrganizationsAPI.RetrieveAnOrganization(ctx, orgID).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading organization",
			"Could not read organization ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	if apiResp == nil || apiResp.Id == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(fmt.Sprintf("%v", int(*apiResp.Id)))
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
func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan organizationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting ID",
			"Could not convert ID to integer: "+err.Error(),
		)
		return
	}

	orgID := float32(id)
	// Build API request body
	body := management.UpdateAnOrganizationRequest{}
	if !plan.Name.IsNull() {
		name := plan.Name.ValueString()
		body.Name = &name
	}
	if !plan.MaintainerID.IsNull() && plan.MaintainerID.ValueString() != "" {
		maintainerID := plan.MaintainerID.ValueString()
		body.MaintainerId = &maintainerID
	}
	if !plan.AllowAutoJoin.IsNull() && plan.AllowAutoJoin.ValueString() != "" {
		allowAutoJoin := plan.AllowAutoJoin.ValueString()
		body.AllowAutoJoin = &allowAutoJoin
	}
	if !plan.CrmID.IsNull() && plan.CrmID.ValueString() != "" {
		crmID := plan.CrmID.ValueString()
		body.CrmId = &crmID
	}
	if !plan.ActivityCenterProjectID.IsNull() && plan.ActivityCenterProjectID.ValueString() != "" {
		acpID := plan.ActivityCenterProjectID.ValueString()
		body.ActivityCenterProjectId = &acpID
	}
	if !plan.MfaRequired.IsNull() && plan.MfaRequired.ValueString() != "" {
		mfaRequired := plan.MfaRequired.ValueString()
		body.MfaRequired = &mfaRequired
	}

	// Update existing organization
	_, _, err = r.client.API.OrganizationsAPI.UpdateAnOrganization(ctx, orgID).UpdateAnOrganizationRequest(body).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating organization",
			"Could not update organization, unexpected error: "+err.Error(),
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
func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state organizationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing organization
	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting ID",
			"Could not convert ID to integer: "+err.Error(),
		)
		return
	}

	orgID := float32(id)
	_, err = r.client.API.OrganizationsAPI.DeleteAnOrganization(ctx, orgID).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting organization",
			"Could not delete organization, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
