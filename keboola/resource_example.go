package keboola

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keboola/keboola-sdk-go/v2/pkg/keboola/management"
)

// resourceMaintainer defines a Keboola Maintainer resource.
func resourceMaintainer() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Maintainer ID.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Maintainer name.",
			},
			"default_connection_redshift_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default Redshift Connection ID.",
			},
			"default_connection_snowflake_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default Snowflake Connection ID.",
			},
			"default_connection_synapse_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default Synapse Connection ID.",
			},
			"default_connection_exasol_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default Exasol Connection ID.",
			},
			"default_connection_teradata_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default Teradata Connection ID.",
			},
			"default_file_storage_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default File Storage ID.",
			},
			"zendesk_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Zendesk URL.",
			},
		},
		CreateContext: resourceMaintainerCreate,
		ReadContext:   resourceMaintainerRead,
		UpdateContext: resourceMaintainerUpdate,
		DeleteContext: resourceMaintainerDelete,
	}
}

// resourceMaintainerCreate creates a maintainer using the Keboola Management API.
func resourceMaintainerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	// Build the request from schema fields
	req := client.API.MaintainersAPI.CreateAMaintainer(ctx)
	body := management.CreateAMaintainerRequest{
		Name: d.Get("name").(string),
	}
	if v, ok := d.GetOk("default_connection_redshift_id"); ok {
		val := v.(string)
		body.DefaultConnectionRedshiftId = &val
	}
	if v, ok := d.GetOk("default_connection_snowflake_id"); ok {
		val := v.(string)
		body.DefaultConnectionSnowflakeId = &val
	}
	if v, ok := d.GetOk("default_connection_synapse_id"); ok {
		val := v.(string)
		body.DefaultConnectionSynapseId = &val
	}
	if v, ok := d.GetOk("default_connection_exasol_id"); ok {
		val := v.(string)
		body.DefaultConnectionExasolId = &val
	}
	if v, ok := d.GetOk("default_connection_teradata_id"); ok {
		val := v.(string)
		body.DefaultConnectionTeradataId = &val
	}
	if v, ok := d.GetOk("default_file_storage_id"); ok {
		val := v.(string)
		body.DefaultFileStorageId = &val
	}
	if v, ok := d.GetOk("zendesk_url"); ok {
		val := v.(string)
		body.ZendeskUrl = &val
	}

	// Call the API
	resp, _, err := req.CreateAMaintainerRequest(body).Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the resource ID (API returns float32, convert to string)
	if resp.Id == nil {
		return diag.Errorf("API did not return maintainer ID")
	}
	d.SetId(fmt.Sprintf("%v", int(*resp.Id)))

	return resourceMaintainerRead(ctx, d, m)
}

// resourceMaintainerRead reads a maintainer from the Keboola Management API.
func resourceMaintainerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	resp, _, err := client.API.MaintainersAPI.RetrieveAMaintainer(ctx, int32(id)).Execute()
	if err != nil {
		return diag.FromErr(err)
	}
	if resp == nil || resp.Id == nil {
		d.SetId("") // Not found
		return nil
	}
	// Update state from API response
	d.Set("name", d.Get("name")) // Name is not returned by API, keep local
	if resp.DefaultConnectionRedshiftId != nil {
		d.Set("default_connection_redshift_id", fmt.Sprintf("%v", int(*resp.DefaultConnectionRedshiftId)))
	}
	if resp.DefaultConnectionSnowflakeId != nil {
		d.Set("default_connection_snowflake_id", fmt.Sprintf("%v", int(*resp.DefaultConnectionSnowflakeId)))
	}
	if resp.DefaultConnectionSynapseId != nil {
		d.Set("default_connection_synapse_id", fmt.Sprintf("%v", int(*resp.DefaultConnectionSynapseId)))
	}
	if resp.DefaultConnectionExasolId != nil {
		d.Set("default_connection_exasol_id", fmt.Sprintf("%v", int(*resp.DefaultConnectionExasolId)))
	}
	if resp.DefaultConnectionTeradataId != nil {
		d.Set("default_connection_teradata_id", fmt.Sprintf("%v", int(*resp.DefaultConnectionTeradataId)))
	}
	if resp.DefaultFileStorageId != nil {
		d.Set("default_file_storage_id", fmt.Sprintf("%v", int(*resp.DefaultFileStorageId)))
	}
	if resp.ZendeskUrl != nil {
		d.Set("zendesk_url", *resp.ZendeskUrl)
	}
	return nil
}

// resourceMaintainerUpdate updates a maintainer using the Keboola Management API.
func resourceMaintainerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	body := management.UpdateAMaintainerRequest{}
	if d.HasChange("name") {
		name := d.Get("name").(string)
		body.Name = &name
	}
	if d.HasChange("default_connection_redshift_id") {
		v := d.Get("default_connection_redshift_id").(string)
		body.DefaultConnectionRedshiftId = &v
	}
	if d.HasChange("default_connection_snowflake_id") {
		v := d.Get("default_connection_snowflake_id").(string)
		body.DefaultConnectionSnowflakeId = &v
	}
	if d.HasChange("default_connection_synapse_id") {
		v := d.Get("default_connection_synapse_id").(string)
		body.DefaultConnectionSynapseId = &v
	}
	if d.HasChange("default_connection_exasol_id") {
		v := d.Get("default_connection_exasol_id").(string)
		body.DefaultConnectionExasolId = &v
	}
	if d.HasChange("default_connection_teradata_id") {
		v := d.Get("default_connection_teradata_id").(string)
		body.DefaultConnectionTeradataId = &v
	}
	if d.HasChange("default_file_storage_id") {
		v := d.Get("default_file_storage_id").(string)
		body.DefaultFileStorageId = &v
	}
	if d.HasChange("zendesk_url") {
		v := d.Get("zendesk_url").(string)
		body.ZendeskUrl = &v
	}
	_, _, err = client.API.MaintainersAPI.UpdateAMaintainer(ctx, int32(id)).UpdateAMaintainerRequest(body).Execute()
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceMaintainerRead(ctx, d, m)
}

// resourceMaintainerDelete deletes a maintainer using the Keboola Management API.
func resourceMaintainerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.API.MaintainersAPI.DeleteAMaintainer(ctx, int32(id)).Execute()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
