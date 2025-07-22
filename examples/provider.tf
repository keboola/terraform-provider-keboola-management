terraform {
  required_providers {
    keboola = {
      source = "keboola/keboola-management"
    }
  }
}

provider "keboola" {
  hostname_suffix = "keboola.com"
  token           = "xxx"
}