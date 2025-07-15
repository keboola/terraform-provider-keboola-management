package keboola

import (
	"context"
	"fmt"

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
	_ resource.Resource                = &projectFeatureResource{}
	_ resource.ResourceWithConfigure   = &projectFeatureResource{}
	_ resource.ResourceWithImportState = &projectFeatureResource{}
)

// NewProjectFeatureAddResource registers the resource in the provider.
func NewProjectFeatureResource() resource.Resource {
	return &projectFeatureResource{}
}

// projectFeatureAddResource implements the resource logic.
type projectFeatureResource struct {
	client *Client
}

type projectFeatureResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Feature   types.String `tfsdk:"feature"`
}

func (r *projectFeatureResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *projectFeatureResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_feature"
}

func (r *projectFeatureResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Adds a feature to a Keboola project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique ID for this resource (project_id:feature).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "ID of the project.",
				Required:    true,
			},
			"feature": schema.StringAttribute{
				Description: "Feature to add to the project.",
				Required:    true,
			},
		},
	}
}

func (r *projectFeatureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectFeatureResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the API request
	apiReq := management.AddAProjectFeatureRequest{
		Feature: plan.Feature.ValueString(),
	}

	// Call the API to add the feature
	_, _, err := r.client.API.SUPERFeaturesAPI.AddAProjectFeature(ctx, plan.ProjectID.ValueString()).AddAProjectFeatureRequest(apiReq).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding feature to project",
			fmt.Sprintf("Could not add feature '%s' to project '%s': %s", plan.Feature.ValueString(), plan.ProjectID.ValueString(), err.Error()),
		)
		return
	}

	// Use project_id:feature as the resource ID
	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.ProjectID.ValueString(), plan.Feature.ValueString()))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *projectFeatureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectFeatureResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get project details to check if feature is present
	apiResp, _, err := r.client.API.ProjectsAPI.ProjectDetail(ctx, state.ProjectID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading project feature",
			fmt.Sprintf("Could not read project '%s': %s", state.ProjectID.ValueString(), err.Error()),
		)
		return
	}
	if apiResp == nil || apiResp.Features == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	found := false
	for _, f := range apiResp.Features { // Features is []string
		if f == state.Feature.ValueString() {
			found = true
			break
		}
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *projectFeatureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectFeatureResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove the feature from the project (no request body needed)
	_, err := r.client.API.SUPERFeaturesAPI.RemoveAProjectFeature(ctx, state.ProjectID.ValueString(), state.Feature.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing feature from project",
			fmt.Sprintf("Could not remove feature '%s' from project '%s': %s", state.Feature.ValueString(), state.ProjectID.ValueString(), err.Error()),
		)
		return
	}
}

// Update is a no-op because features cannot be updated, only added or removed.
func (r *projectFeatureResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update operation for project features.
}

func (r *projectFeatureResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID (project_id:feature)
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
