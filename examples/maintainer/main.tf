terraform {
  required_providers {
    keboola-management = {
      source  = "keboola/keboola-management"
      version = "0.1.2"
    }
    keboola = {
      source  = "keboola/keboola"
      version = "0.4.0"
    }
  }
}

provider "keboola-management" {
  hostname_suffix = var.hostname_suffix
  token           = var.token
}


resource "keboola-management_maintainer" "full" {
  name                            = "Full Example Maintainer"
  default_connection_snowflake_id = "1"
  default_file_storage_id         = "1"
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
  name                        = "Example Project2"
  organization_id             = keboola-management_organization.example.id # Reference to the organization resource
  type                        = "production"                               # or poc, demo
  default_backend             = "snowflake"                                # or redshift
  data_retention_time_in_days = "7"                                        # optional, e.g. 7 days

  token {
    description               = "Test Token" # Token description
    can_manage_buckets        = true         # Full permissions on tabular storage
    can_read_all_file_uploads = true         # Full permissions to files staging
    can_purge_trash           = true         # Allows permanently remove deleted configurations
    expires_in                = 7200         # Token lifetime in seconds
    bucket_permissions = {
      "in.c" = "string" # Example bucket permission
    }
    component_access = ["string"] # List of component IDs (as strings)
  }
}

# Example: Add project features
resource "keboola-management_project_feature" "data_streams" {
  project_id = keboola-management_project.example.id
  feature    = "data-streams"
}

resource "keboola-management_project_feature" "show_vault" {
  project_id = keboola-management_project.example.id
  feature    = "show-vault"
}

resource "keboola-management_project_feature" "syrup_jobs_limit_10" {
  project_id = keboola-management_project.example.id
  feature    = "syrup-jobs-limit-10"
}

resource "keboola-management_project_feature" "waii_integration" {
  project_id = keboola-management_project.example.id
  feature    = "waii-integration"
}

# Example: Project Invitation resource
resource "keboola-management_project_invitation" "example" {
  project_id = keboola-management_project.example.id
  email      = "role_dev_go+cicd@keboola.com"
  role       = "admin" # Change to the desired role
  # expiration_seconds = 86400 # Optional: 1 day
  reason = "CI/CD automation invitation"
}

# Output the storage token (sensitive, only available after creation)
output "project_storage_token" {
  value     = keboola-management_project.example.storage_token
  sensitive = true
}

# Example: Pass the storage token to the keboola provider (keboola/keboola)
provider "keboola" {
  host = "https://connection.${var.hostname_suffix}"
}

# Now you can use the keboola provider for other resources
# resource "keboola_some_resource" "example" {
#   # ... config ...
# }


resource "keboola_component_configuration" "generic_example" {
  component_id = "ex-generic-v2"
  name         = "My Generic Extractor Config"
  description  = "Created by Terraform"

  configuration = jsonencode({
    parameters = {
      # Add your extractor parameters here
      # Example:
      # api = {
      #   baseUrl = "https://api.example.com"
      # }
    }
  })
}