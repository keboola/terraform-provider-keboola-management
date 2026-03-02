terraform {
  required_providers {
    keboola-management = {
      source  = "keboola/keboola-management"
      version = "0.1.2"
    }
  }
}

variable "hostname_suffix" {
  description = "Hostname suffix for the Keboola Management API (e.g., keboola.com)"
  type        = string
}

variable "token" {
  description = "Keboola Management API token"
  type        = string
}

provider "keboola-management" {
  hostname_suffix = var.hostname_suffix
  token           = var.token
}

# ------------------------------------------------------------------------------
# Backend and File Storage Registration Examples
# ------------------------------------------------------------------------------

resource "keboola-management_backend" "example" {
  backend   = "snowflake"
  host      = "example.snowflakecomputing.com"
  username  = "terraform_user"
  password  = "supersecret" # Use a secure method for real deployments
  region    = "us-east-1"
  owner     = "aws-account-owner"
  warehouse = "COMPUTE_WH"
  # database  = "optional-db" # Uncomment for Synapse/Teradata
  # use_synapse_managed_identity = "false" # Optional for Synapse
  # use_dynamic_backends = true # Optional for supported backends
}

resource "keboola-management_backend_bigquery" "example" {
  owner     = "gcp-account-owner"
  folder_id = "gcp-folder-id"
  region    = "europe-west2"

  credentials {
    type                        = "service_account"
    project_id                  = "gcp-project-id"
    private_key_id              = "private-key-id"
    private_key                 = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n" # Sensitive
    client_email                = "service-account@example.iam.gserviceaccount.com"
    client_id                   = "client-id"
    auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
    token_uri                   = "https://oauth2.googleapis.com/token"
    auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
    client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/service-account%40example.iam.gserviceaccount.com"
  }
}

# ------------------------------------------------------------------------------
# Snowflake Backend with RSA Certificate Authentication
# This is a two-step process:
# 1. Register the backend (returns user_public_key and sql_template)
# 2. Configure the Snowflake user with the public key (external step)
# 3. Activate the backend
# ------------------------------------------------------------------------------

resource "keboola-management_backend_snowflake" "example" {
  host            = "example.snowflakecomputing.com"
  warehouse       = "KEBOOLA"
  username        = "KEBOOLA_STORAGE"
  region          = "us-east-1"
  owner           = "keboola"
  technical_owner = "keboola"

  # Optional fields:
  # technical_owner_contact_emails = ["admin@example.com"]
  # use_dynamic_backends           = true
  # use_network_policies           = true
  # use_sso                        = false
  # edition                        = "ENTERPRISE"
}

# After configuring the Snowflake user with the public key from the above resource:
resource "keboola-management_backend_snowflake_activate" "example" {
  backend_id = keboola-management_backend_snowflake.example.id
}

# Useful outputs for the intermediate Snowflake configuration step:
output "snowflake_user_public_key" {
  value       = keboola-management_backend_snowflake.example.user_public_key
  description = "RSA public key to configure on the Snowflake user"
}

output "snowflake_sql_template" {
  value       = keboola-management_backend_snowflake.example.sql_template
  description = "SQL template to execute on the Snowflake account"
}

resource "keboola_file_storage_s3" "example" {
  aws_key      = "AKIA..."
  aws_secret   = "..." # sensitive
  files_bucket = "my-bucket"
  region       = "eu-central-1"
  owner        = "my-aws-account-id"
}

resource "keboola_file_storage_gcs" "example" {
  files_bucket = "my-gcs-bucket"
  owner        = "gcp-account-owner"
  region       = "europe-west2"

  gcs_credentials {
    type                        = "service_account"
    project_id                  = "gcp-project-id"
    private_key_id              = "private-key-id"
    private_key                 = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n" # Sensitive
    client_email                = "service-account@example.iam.gserviceaccount.com"
    client_id                   = "client-id"
    auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
    token_uri                   = "https://oauth2.googleapis.com/token"
    auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
    client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/service-account%40example.iam.gserviceaccount.com"
  }
}

resource "keboola_file_storage_azure_blob" "example" {
  account_name   = "myazureaccount"
  account_key    = "..." # sensitive
  owner          = "azure-account-owner"
  container_name = "my-container" # optional
} 