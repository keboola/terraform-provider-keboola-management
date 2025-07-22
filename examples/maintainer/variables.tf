variable "hostname_suffix" {
  type        = string
  description = "Hostname suffix for the Keboola Stacks (e.g., keboola.com)"
}

variable "token" {
  type        = string
  description = "Keboola Management API token"
  sensitive   = true
} 