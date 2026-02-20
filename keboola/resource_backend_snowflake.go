package keboola

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/keboola/keboola-sdk-go/v2/pkg/keboola/management"
)

// This resource only supports Create. Read, Update, and Delete are not supported by the API.

var (
	_ resource.Resource              = &backendSnowflakeResource{}
	_ resource.ResourceWithConfigure = &backendSnowflakeResource{}
)

func NewBackendSnowflakeResource() resource.Resource {
	return &backendSnowflakeResource{}
}

type backendSnowflakeResource struct {
	client *Client
}

type backendSnowflakeResourceModel struct {
	ID                          types.String `tfsdk:"id"`
	Host                        types.String `tfsdk:"host"`
	Warehouse                   types.String `tfsdk:"warehouse"`
	Region                      types.String `tfsdk:"region"`
	Owner                       types.String `tfsdk:"owner"`
	TechnicalOwner              types.String `tfsdk:"technical_owner"`
	Username                    types.String `tfsdk:"username"`
	TechnicalOwnerContactEmails types.List   `tfsdk:"technical_owner_contact_emails"`
	UseDynamicBackends          types.Bool   `tfsdk:"use_dynamic_backends"`
	UseNetworkPolicies          types.Bool   `tfsdk:"use_network_policies"`
	UseSso                      types.Bool   `tfsdk:"use_sso"`
	Edition                     types.String `tfsdk:"edition"`
	SQLTemplate                 types.String `tfsdk:"sql_template"`
	IsEnabled                   types.Bool   `tfsdk:"is_enabled"`
	UserPublicKey               types.String `tfsdk:"user_public_key"`
	SecurityIntegrationKey      types.String `tfsdk:"security_integration_key"`
}

func (r *backendSnowflakeResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *backendSnowflakeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend_snowflake"
}

func (r *backendSnowflakeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Registers a Snowflake storage backend with RSA certificate authentication. Only Create is supported. Read, Update, and Delete are not supported by the API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Backend ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host": schema.StringAttribute{
				Description: "Snowflake host.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"warehouse": schema.StringAttribute{
				Description: "Snowflake warehouse name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Description: "Snowflake region.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"owner": schema.StringAttribute{
				Description: "Snowflake account owner.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"technical_owner": schema.StringAttribute{
				Description: "Technical owner of the backend.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Snowflake username.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"technical_owner_contact_emails": schema.ListAttribute{
				Description: "Contact emails for the technical owner.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"use_dynamic_backends": schema.BoolAttribute{
				Description: "Enable dynamic backends.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"use_network_policies": schema.BoolAttribute{
				Description: "Enable network policies.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"use_sso": schema.BoolAttribute{
				Description: "Enable SSO.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"edition": schema.StringAttribute{
				Description: "Snowflake edition.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sql_template": schema.StringAttribute{
				Description: "SQL template returned by the API.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_enabled": schema.BoolAttribute{
				Description: "Whether the backend is enabled.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"user_public_key": schema.StringAttribute{
				Description: "RSA public key for the Snowflake user. Use this to configure the Snowflake user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"security_integration_key": schema.StringAttribute{
				Description: "Security integration key returned by the API.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *backendSnowflakeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan backendSnowflakeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := management.CreateSnowflakeBackendWithCertRequest{
		Host:           plan.Host.ValueString(),
		Warehouse:      plan.Warehouse.ValueString(),
		Region:         plan.Region.ValueString(),
		Owner:          plan.Owner.ValueString(),
		TechnicalOwner: plan.TechnicalOwner.ValueString(),
	}

	if !plan.Username.IsNull() && !plan.Username.IsUnknown() {
		v := plan.Username.ValueString()
		apiReq.Username = &v
	}

	if !plan.TechnicalOwnerContactEmails.IsNull() && !plan.TechnicalOwnerContactEmails.IsUnknown() {
		var emails []string
		diags = plan.TechnicalOwnerContactEmails.ElementsAs(ctx, &emails, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiReq.TechnicalOwnerContactEmails = emails
	}

	if !plan.UseDynamicBackends.IsNull() && !plan.UseDynamicBackends.IsUnknown() {
		v := plan.UseDynamicBackends.ValueBool()
		apiReq.UseDynamicBackends = &v
	}

	if !plan.UseNetworkPolicies.IsNull() && !plan.UseNetworkPolicies.IsUnknown() {
		v := plan.UseNetworkPolicies.ValueBool()
		apiReq.UseNetworkPolicies = &v
	}

	if !plan.UseSso.IsNull() && !plan.UseSso.IsUnknown() {
		v := plan.UseSso.ValueBool()
		apiReq.UseSso = &v
	}

	if !plan.Edition.IsNull() && !plan.Edition.IsUnknown() {
		v := plan.Edition.ValueString()
		apiReq.Edition = &v
	}

	apiResp, _, err := r.client.API.SUPERStorageBackendsManagementAPI.CreateSnowflakeBackendWithCert(ctx).CreateSnowflakeBackendWithCertRequest(apiReq).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Snowflake backend",
			fmt.Sprintf("Could not create Snowflake backend: %s", err.Error()),
		)
		return
	}

	if apiResp.Id != nil {
		plan.ID = types.StringValue(fmt.Sprintf("%v", int64(*apiResp.Id)))
	}
	if apiResp.SqlTemplate != nil {
		plan.SQLTemplate = types.StringValue(*apiResp.SqlTemplate)
	} else {
		plan.SQLTemplate = types.StringNull()
	}
	if apiResp.IsEnabled != nil {
		plan.IsEnabled = types.BoolValue(*apiResp.IsEnabled)
	} else {
		plan.IsEnabled = types.BoolNull()
	}
	if apiResp.UserPublicKey != nil {
		plan.UserPublicKey = types.StringValue(*apiResp.UserPublicKey)
	} else {
		plan.UserPublicKey = types.StringNull()
	}
	if apiResp.SecurityIntegrationKey != nil {
		plan.SecurityIntegrationKey = types.StringValue(*apiResp.SecurityIntegrationKey)
	} else {
		plan.SecurityIntegrationKey = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *backendSnowflakeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddWarning("Read not supported", "Snowflake backend does not support read operation. State will not be refreshed.")
}

func (r *backendSnowflakeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Update not supported", "Snowflake backend does not support update operation. All fields require replacement.")
}

func (r *backendSnowflakeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Delete not supported", "Snowflake backend does not support delete operation. Resource will remain in state.")
}
