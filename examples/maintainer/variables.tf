variable "api_url" {
  type        = string
  description = "URL of the Keboola Management API"
}

variable "token" {
  type        = string
  description = "Keboola Management API token"
  sensitive   = true
} 