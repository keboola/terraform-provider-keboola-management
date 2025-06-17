variable "host" {
  type        = string
  description = "Keboola Management API Host"
}

variable "token" {
  type        = string
  description = "Keboola Management API Token"
  sensitive   = true
} 