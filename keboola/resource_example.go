package keboola

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/keboola/keboola-sdk-go/v2/pkg/keboola/management"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource = &maintainerResource{}
)

// NewMaintainerResource is a helper function to simplify the provider implementation.
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

// Metadata returns the resource type name.
func (r *maintainerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintainer"
}

// Schema defines the schema for the resource.
func (r *maintainerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Keboola Maintainer.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Maintainer ID.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Maintainer name.",
			},
			"default_connection_redshift_id": schema.StringAttribute{
				Optional:    true,
				Description: "Default Redshift Connection ID.",
			},
			"default_connection_snowflake_id": schema.StringAttribute{
				Optional:    true,
				Description: "Default Snowflake Connection ID.",
			},
			"default_connection_synapse_id": schema.StringAttribute{
				Optional:    true,
				Description: "Default Synapse Connection ID.",
			},
			"default_connection_exasol_id": schema.StringAttribute{
				Optional:    true,
				Description: "Default Exasol Connection ID.",
			},
			"default_connection_teradata_id": schema.StringAttribute{
				Optional:    true,
				Description: "Default Teradata Connection ID.",
			},
			"default_file_storage_id": schema.StringAttribute{
				Optional:    true,
				Description: "Default File Storage ID.",
			},
			"zendesk_url": schema.StringAttribute{
				Optional:    true,
				Description: "Zendesk URL.",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *maintainerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *maintainerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan maintainerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API request body
	apiReq := r.client.API.MaintainersAPI.CreateAMaintainer(ctx)
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
	apiResp, _, err := apiReq.CreateAMaintainerRequest(body).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating maintainer",
			fmt.Sprintf("Could not create maintainer: %s", err),
		)
		return
	}

	if apiResp.Id == nil {
		resp.Diagnostics.AddError(
			"Error creating maintainer",
			"API did not return maintainer ID",
		)
		return
	}

	// Set ID in the Terraform state
	plan.ID = types.StringValue(fmt.Sprintf("%v", int(*apiResp.Id)))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *maintainerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state maintainerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing maintainer ID",
			fmt.Sprintf("Could not parse maintainer ID: %s", err),
		)
		return
	}

	apiResp, _, err := r.client.API.MaintainersAPI.RetrieveAMaintainer(ctx, int32(id)).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading maintainer",
			fmt.Sprintf("Could not read maintainer ID %d: %s", id, err),
		)
		return
	}

	if apiResp == nil || apiResp.Id == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with API response
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

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *maintainerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan maintainerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing maintainer ID",
			fmt.Sprintf("Could not parse maintainer ID: %s", err),
		)
		return
	}

	body := management.UpdateAMaintainerRequest{}
	name := plan.Name.ValueString()
	body.Name = &name

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

	_, _, err = r.client.API.MaintainersAPI.UpdateAMaintainer(ctx, int32(id)).UpdateAMaintainerRequest(body).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating maintainer",
			fmt.Sprintf("Could not update maintainer ID %d: %s", id, err),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *maintainerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state maintainerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing maintainer ID",
			fmt.Sprintf("Could not parse maintainer ID: %s", err),
		)
		return
	}

	_, err = r.client.API.MaintainersAPI.DeleteAMaintainer(ctx, int32(id)).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting maintainer",
			fmt.Sprintf("Could not delete maintainer ID %d: %s", id, err),
		)
		return
	}
}
