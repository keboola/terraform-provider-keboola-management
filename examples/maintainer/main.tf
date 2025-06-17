terraform {
  required_providers {
    keboola = {
      source = "keboola/keboola-management"
    }
  }
}

# Configure the provider
provider "keboola" {
  host  = var.host  # URL of your Keboola Management API
  token = var.token # Your Keboola Management API token
}

# Example 1: Minimal maintainer configuration
resource "keboola_maintainer" "minimal" {
  name = "Example Maintainer"
}

# Example 2: Full maintainer configuration with all optional fields
resource "keboola_maintainer" "full" {
  name                            = "Full Example Maintainer"
  default_connection_redshift_id  = "123"
  default_connection_snowflake_id = "456"
  default_connection_synapse_id   = "789"
  default_connection_exasol_id    = "101"
  default_connection_teradata_id  = "102"
  default_file_storage_id         = "103"
  zendesk_url                     = "https://example.zendesk.com"
}