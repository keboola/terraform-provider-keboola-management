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
	_ resource.Resource                = &maintainerResource{}
	_ resource.ResourceWithConfigure   = &maintainerResource{}
	_ resource.ResourceWithImportState = &maintainerResource{}
)

// NewMaintainerResource is a helper function to simplify provider implementation.
func NewMaintainerResource() resource.Resource {
	return &maintainerResource{}
}

// maintainerResource is the resource implementation.
type maintainerResource struct {
	client *Client
}

// maintainerResourceModel maps the resource schema data.
type maintainerResourceModel struct {
	ID                           types.String `tfsdk:"id"`
	Name                         types.String `tfsdk:"name"`
	DefaultConnectionRedshiftID  types.String `tfsdk:"default_connection_redshift_id"`
	DefaultConnectionSnowflakeID types.String `tfsdk:"default_connection_snowflake_id"`
	DefaultConnectionSynapseID   types.String `tfsdk:"default_connection_synapse_id"`
	DefaultConnectionExasolID    types.String `tfsdk:"default_connection_exasol_id"`
	DefaultConnectionTeradataID  types.String `tfsdk:"default_connection_teradata_id"`
	DefaultFileStorageID         types.String `tfsdk:"default_file_storage_id"`
	ZendeskURL                   types.String `tfsdk:"zendesk_url"`
}

// Configure adds the provider configured client to the resource.
func (r *maintainerResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

// Metadata returns the resource type name.
func (r *maintainerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintainer"
}

// Schema defines the schema for the resource.
func (r *maintainerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Keboola maintainer.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Maintainer ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Maintainer name.",
				Required:    true,
			},
			"default_connection_redshift_id": schema.StringAttribute{
				Description: "Default Redshift Connection ID.",
				Optional:    true,
			},
			"default_connection_snowflake_id": schema.StringAttribute{
				Description: "Default Snowflake Connection ID.",
				Optional:    true,
			},
			"default_connection_synapse_id": schema.StringAttribute{
				Description: "Default Synapse Connection ID.",
				Optional:    true,
			},
			"default_connection_exasol_id": schema.StringAttribute{
				Description: "Default Exasol Connection ID.",
				Optional:    true,
			},
			"default_connection_teradata_id": schema.StringAttribute{
				Description: "Default Teradata Connection ID.",
				Optional:    true,
			},
			"default_file_storage_id": schema.StringAttribute{
				Description: "Default File Storage ID.",
				Optional:    true,
			},
			"zendesk_url": schema.StringAttribute{
				Description: "Zendesk URL.",
				Optional:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *maintainerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan maintainerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API request body
	body := management.CreateAMaintainerRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.DefaultConnectionRedshiftID.IsNull() {
		val := plan.DefaultConnectionRedshiftID.ValueString()
		body.DefaultConnectionRedshiftId = &val
	}
	if !plan.DefaultConnectionSnowflakeID.IsNull() {
		val := plan.DefaultConnectionSnowflakeID.ValueString()
		body.DefaultConnectionSnowflakeId = &val
	}
	if !plan.DefaultConnectionSynapseID.IsNull() {
		val := plan.DefaultConnectionSynapseID.ValueString()
		body.DefaultConnectionSynapseId = &val
	}
	if !plan.DefaultConnectionExasolID.IsNull() {
		val := plan.DefaultConnectionExasolID.ValueString()
		body.DefaultConnectionExasolId = &val
	}
	if !plan.DefaultConnectionTeradataID.IsNull() {
		val := plan.DefaultConnectionTeradataID.ValueString()
		body.DefaultConnectionTeradataId = &val
	}
	if !plan.DefaultFileStorageID.IsNull() {
		val := plan.DefaultFileStorageID.ValueString()
		body.DefaultFileStorageId = &val
	}
	if !plan.ZendeskURL.IsNull() {
		val := plan.ZendeskURL.ValueString()
		body.ZendeskUrl = &val
	}

	// Create new maintainer
	apiResp, _, err := r.client.API.MaintainersAPI.CreateAMaintainer(ctx).CreateAMaintainerRequest(body).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating maintainer",
			"Could not create maintainer, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	if apiResp.Id == nil {
		resp.Diagnostics.AddError(
			"Error creating maintainer",
			"API did not return maintainer ID",
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
func (r *maintainerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state maintainerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed maintainer value from API
	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting ID",
			"Could not convert ID to integer: "+err.Error(),
		)
		return
	}

	apiResp, _, err := r.client.API.MaintainersAPI.RetrieveAMaintainer(ctx, int32(id)).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading maintainer",
			"Could not read maintainer ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	if apiResp == nil || apiResp.Id == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(fmt.Sprintf("%v", int(*apiResp.Id)))
	// Name is not returned by API, keep local value
	if apiResp.DefaultConnectionRedshiftId != nil {
		state.DefaultConnectionRedshiftID = types.StringValue(fmt.Sprintf("%v", int(*apiResp.DefaultConnectionRedshiftId)))
	}
	if apiResp.DefaultConnectionSnowflakeId != nil {
		state.DefaultConnectionSnowflakeID = types.StringValue(fmt.Sprintf("%v", int(*apiResp.DefaultConnectionSnowflakeId)))
	}
	if apiResp.DefaultConnectionSynapseId != nil {
		state.DefaultConnectionSynapseID = types.StringValue(fmt.Sprintf("%v", int(*apiResp.DefaultConnectionSynapseId)))
	}
	if apiResp.DefaultConnectionExasolId != nil {
		state.DefaultConnectionExasolID = types.StringValue(fmt.Sprintf("%v", int(*apiResp.DefaultConnectionExasolId)))
	}
	if apiResp.DefaultConnectionTeradataId != nil {
		state.DefaultConnectionTeradataID = types.StringValue(fmt.Sprintf("%v", int(*apiResp.DefaultConnectionTeradataId)))
	}
	if apiResp.DefaultFileStorageId != nil {
		state.DefaultFileStorageID = types.StringValue(fmt.Sprintf("%v", int(*apiResp.DefaultFileStorageId)))
	}
	if apiResp.ZendeskUrl != nil {
		state.ZendeskURL = types.StringValue(*apiResp.ZendeskUrl)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *maintainerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan maintainerResourceModel
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

	// Build API request body
	body := management.UpdateAMaintainerRequest{}
	if !plan.Name.IsNull() {
		name := plan.Name.ValueString()
		body.Name = &name
	}
	if !plan.DefaultConnectionRedshiftID.IsNull() {
		val := plan.DefaultConnectionRedshiftID.ValueString()
		body.DefaultConnectionRedshiftId = &val
	}
	if !plan.DefaultConnectionSnowflakeID.IsNull() {
		val := plan.DefaultConnectionSnowflakeID.ValueString()
		body.DefaultConnectionSnowflakeId = &val
	}
	if !plan.DefaultConnectionSynapseID.IsNull() {
		val := plan.DefaultConnectionSynapseID.ValueString()
		body.DefaultConnectionSynapseId = &val
	}
	if !plan.DefaultConnectionExasolID.IsNull() {
		val := plan.DefaultConnectionExasolID.ValueString()
		body.DefaultConnectionExasolId = &val
	}
	if !plan.DefaultConnectionTeradataID.IsNull() {
		val := plan.DefaultConnectionTeradataID.ValueString()
		body.DefaultConnectionTeradataId = &val
	}
	if !plan.DefaultFileStorageID.IsNull() {
		val := plan.DefaultFileStorageID.ValueString()
		body.DefaultFileStorageId = &val
	}
	if !plan.ZendeskURL.IsNull() {
		val := plan.ZendeskURL.ValueString()
		body.ZendeskUrl = &val
	}

	// Update existing maintainer
	_, _, err = r.client.API.MaintainersAPI.UpdateAMaintainer(ctx, int32(id)).UpdateAMaintainerRequest(body).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating maintainer",
			"Could not update maintainer, unexpected error: "+err.Error(),
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
func (r *maintainerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state maintainerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing maintainer
	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting ID",
			"Could not convert ID to integer: "+err.Error(),
		)
		return
	}

	_, err = r.client.API.MaintainersAPI.DeleteAMaintainer(ctx, int32(id)).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting maintainer",
			"Could not delete maintainer, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *maintainerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
