package keboola

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/keboola/keboola-sdk-go/v2/pkg/keboola/management"
)

// This resource only supports Create. Read, Update, and Delete are not supported by the API.

var (
	_ resource.Resource              = &backendSnowflakeActivateResource{}
	_ resource.ResourceWithConfigure = &backendSnowflakeActivateResource{}
)

func NewBackendSnowflakeActivateResource() resource.Resource {
	return &backendSnowflakeActivateResource{}
}

type backendSnowflakeActivateResource struct {
	client *Client
}

type backendSnowflakeActivateResourceModel struct {
	ID        types.String `tfsdk:"id"`
	BackendID types.String `tfsdk:"backend_id"`
	IsEnabled types.Bool   `tfsdk:"is_enabled"`
}

func (r *backendSnowflakeActivateResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *backendSnowflakeActivateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend_snowflake_activate"
}

func (r *backendSnowflakeActivateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Activates a Snowflake storage backend after the Snowflake user has been configured with the public key. Only Create is supported.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Same as backend_id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"backend_id": schema.StringAttribute{
				Description: "ID of the Snowflake backend to activate.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_enabled": schema.BoolAttribute{
				Description: "Whether the backend is enabled after activation.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *backendSnowflakeActivateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan backendSnowflakeActivateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backendID := plan.BackendID.ValueString()

	apiResp, _, err := r.client.API.SUPERStorageBackendsManagementAPI.ActivateSnowflakeBackend(ctx, backendID).Execute()
	if err != nil {
		detail := err.Error()
		if apiErr, ok := err.(*management.GenericOpenAPIError); ok {
			detail = fmt.Sprintf("%s: %s", err.Error(), string(apiErr.Body()))
		}
		resp.Diagnostics.AddError(
			"Error activating Snowflake backend",
			fmt.Sprintf("Could not activate Snowflake backend '%s': %s", backendID, detail),
		)
		return
	}

	plan.ID = types.StringValue(backendID)
	if apiResp.IsEnabled != nil {
		plan.IsEnabled = types.BoolValue(*apiResp.IsEnabled)
	} else {
		plan.IsEnabled = types.BoolNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *backendSnowflakeActivateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddWarning("Read not supported", "Snowflake backend activation does not support read operation. State will not be refreshed.")
}

func (r *backendSnowflakeActivateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Update not supported", "Snowflake backend activation does not support update operation.")
}

func (r *backendSnowflakeActivateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Delete not supported", "Snowflake backend activation does not support delete operation. Resource will remain in state.")
}
