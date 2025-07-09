terraform {
  required_providers {
    keboola-management = {
      source  = "keboola/keboola-management"
      version = "0.1.2"
    }
  }
}

provider "keboola-management" {
  url   = var.url
  token = var.token
}


resource "keboola-management_maintainer" "full" {
  name                            = "Full Example Maintainer"
  default_connection_redshift_id  = "123"
  default_connection_snowflake_id = "456"
  default_connection_synapse_id   = "789"
  default_connection_exasol_id    = "101"
  default_connection_teradata_id  = "102"
  default_file_storage_id         = "103"
  zendesk_url                     = "https://example.zendesk.com"
}

# Example: Organization resource
resource "keboola-management_organization" "example" {
  name          = "Example Organization"
  maintainer_id = keboola-management_maintainer.full.id # Reference to the maintainer resource
  # crm_id     = "optional-crm-id" # Uncomment and set if you want to specify a CRM ID
}

# Example: Project resource
resource "keboola-management_project" "example" {
  name                        = "Example Project"
  organization_id             = keboola-management_organization.example.id # Reference to the organization resource
  type                        = "production"                               # or poc, demo
  default_backend             = "snowflake"                                # or redshift
  data_retention_time_in_da3s = "7"                                        # optional, e.g. 7 days
}