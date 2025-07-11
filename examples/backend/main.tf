terraform {
  required_providers {
    keboola-management = {
      source  = "keboola/keboola-management"
      version = "0.1.2"
    }
  }
}

variable "url" {
  description = "Keboola Management API URL"
  type        = string
}

variable "token" {
  description = "Keboola Management API token"
  type        = string
}

provider "keboola-management" {
  url   = var.url
  token = var.token
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