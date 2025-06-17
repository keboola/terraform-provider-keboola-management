terraform {
  required_providers {
    keboola = {
      source = "keboola/keboola-management"
    }
  }
}

provider "keboola" {
  url   = "https://connection.keboola.com"
  token = "xxx"
}