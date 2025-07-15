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

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource              = &backendResource{}
	_ resource.ResourceWithConfigure = &backendResource{}
)

// NewBackendResource returns a new backend resource instance.
func NewBackendResource() resource.Resource {
	return &backendResource{}
}

// backendResource implements the Terraform resource for storage backends.
type backendResource struct {
	client *Client
}

// backendResourceModel maps the resource schema data.
type backendResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Backend                   types.String `tfsdk:"backend"`
	Host                      types.String `tfsdk:"host"`
	Username                  types.String `tfsdk:"username"`
	Password                  types.String `tfsdk:"password"`
	Region                    types.String `tfsdk:"region"`
	Owner                     types.String `tfsdk:"owner"`
	Warehouse                 types.String `tfsdk:"warehouse"`
	Database                  types.String `tfsdk:"database"`
	UseSynapseManagedIdentity types.String `tfsdk:"use_synapse_managed_identity"`
	UseDynamicBackends        types.Bool   `tfsdk:"use_dynamic_backends"`
}

// Configure adds the provider configured client to the resource.
func (r *backendResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

// Metadata returns the resource type name.
func (r *backendResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend"
}

// Schema defines the schema for the resource.
func (r *backendResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Keboola storage backend (except BigQuery).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Backend ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"backend": schema.StringAttribute{
				Description: "Backend type (e.g., bigquery, snowflake, redshift, synapse, exasol, teradata).",
				Required:    true,
			},
			"host": schema.StringAttribute{
				Description: "Backend host.",
				Required:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username for backend.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for backend.",
				Required:    true,
				Sensitive:   true,
			},
			"region": schema.StringAttribute{
				Description: "Backend region.",
				Required:    true,
			},
			"owner": schema.StringAttribute{
				Description: "Associated AWS account owner.",
				Required:    true,
			},
			"warehouse": schema.StringAttribute{
				Description: "Warehouse (required for Snowflake).",
				Optional:    true,
			},
			"database": schema.StringAttribute{
				Description: "Database (required for Synapse and Teradata).",
				Optional:    true,
			},
			"use_synapse_managed_identity": schema.StringAttribute{
				Description: "Use Synapse Managed Identity (optional for Synapse).",
				Optional:    true,
			},
			"use_dynamic_backends": schema.BoolAttribute{
				Description: "Enable dynamic backends (only for supported backends, e.g., Snowflake).",
				Optional:    true,
			},
		},
	}
}

// Create creates the backend resource.
func (r *backendResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan backendResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the API request
	apiReq := management.CreateANewBackendRequest{
		Backend:  plan.Backend.ValueString(),
		Host:     plan.Host.ValueString(),
		Username: plan.Username.ValueString(),
		Password: plan.Password.ValueString(),
		Region:   plan.Region.ValueString(),
		Owner:    plan.Owner.ValueString(),
	}
	if !plan.Warehouse.IsNull() && plan.Warehouse.ValueString() != "" {
		w := plan.Warehouse.ValueString()
		apiReq.Warehouse = &w
	}
	if !plan.Database.IsNull() && plan.Database.ValueString() != "" {
		db := plan.Database.ValueString()
		apiReq.Database = &db
	}
	if !plan.UseSynapseManagedIdentity.IsNull() && plan.UseSynapseManagedIdentity.ValueString() != "" {
		id := plan.UseSynapseManagedIdentity.ValueString()
		apiReq.UseSynapseManagedIdentity = &id
	}
	if !plan.UseDynamicBackends.IsNull() {
		val := plan.UseDynamicBackends.ValueBool()
		apiReq.UseDynamicBackends = &val
	}

	// Call the API to create the backend
	apiResp, _, err := r.client.API.SUPERStorageBackendsManagementAPI.CreateANewBackend(ctx).CreateANewBackendRequest(apiReq).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating backend",
			fmt.Sprintf("Could not create backend: %s", err.Error()),
		)
		return
	}
	// apiResp.Id is float32, not a pointer
	plan.ID = types.StringValue(fmt.Sprintf("%v", apiResp.Id))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the backend resource state.
func (r *backendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state backendResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the API to get backend details
	apiResp, _, err := r.client.API.SUPERStorageBackendsManagementAPI.BackendDetail(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading backend",
			fmt.Sprintf("Could not read backend '%s': %s", state.ID.ValueString(), err.Error()),
		)
		return
	}
	if apiResp == nil || len(apiResp) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}
	// The response is []interface{}, map fields as needed (assume first element is the backend)
	backend, ok := apiResp[0].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing backend detail",
			"API response format unexpected.",
		)
		return
	}
	if id, ok := backend["id"].(float64); ok {
		state.ID = types.StringValue(fmt.Sprintf("%v", id))
	}
	if backendType, ok := backend["backend"].(string); ok {
		state.Backend = types.StringValue(backendType)
	}
	if host, ok := backend["host"].(string); ok {
		state.Host = types.StringValue(host)
	}
	if username, ok := backend["username"].(string); ok {
		state.Username = types.StringValue(username)
	}
	if region, ok := backend["region"].(string); ok {
		state.Region = types.StringValue(region)
	}
	if owner, ok := backend["owner"].(string); ok {
		state.Owner = types.StringValue(owner)
	}
	if warehouse, ok := backend["warehouse"].(string); ok {
		state.Warehouse = types.StringValue(warehouse)
	}
	if database, ok := backend["database"].(string); ok {
		state.Database = types.StringValue(database)
	}
	if useSynapse, ok := backend["useSynapseManagedIdentity"].(string); ok {
		state.UseSynapseManagedIdentity = types.StringValue(useSynapse)
	}
	if useDynamic, ok := backend["useDynamicBackends"].(bool); ok {
		state.UseDynamicBackends = types.BoolValue(useDynamic)
	}
	// Password is not returned for security reasons
	state.Password = types.StringNull()

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the backend resource.
func (r *backendResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan backendResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the API request
	apiReq := management.StorageBackendUpdate{}
	if !plan.Username.IsNull() && plan.Username.ValueString() != "" {
		u := plan.Username.ValueString()
		apiReq.Username = &u
	}
	if !plan.Password.IsNull() && plan.Password.ValueString() != "" {
		p := plan.Password.ValueString()
		apiReq.Password = &p
	}
	if !plan.UseDynamicBackends.IsNull() {
		val := plan.UseDynamicBackends.ValueBool()
		apiReq.UseDynamicBackends = &val
	}

	// Call the API to update the backend
	_, _, err := r.client.API.SUPERStorageBackendsManagementAPI.UpdateBackend(ctx, plan.ID.ValueString()).StorageBackendUpdate(apiReq).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating backend",
			fmt.Sprintf("Could not update backend '%s': %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the backend resource.
func (r *backendResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state backendResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the API to delete the backend
	_, err := r.client.API.SUPERStorageBackendsManagementAPI.DeleteBackend(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting backend",
			fmt.Sprintf("Could not delete backend '%s': %s", state.ID.ValueString(), err.Error()),
		)
		return
	}
}
