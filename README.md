# Keboola Terraform Provider

> **Note:** If you want to manage Keboola Storage API resources, use the official provider:
> - [GitHub Repository](https://github.com/keboola/terraform-provider-keboola)
> - [Terraform Registry](https://registry.terraform.io/providers/keboola/keboola/latest)
>
> This provider (`terraform-provider-keboola-management`) is for the Keboola **Management API** only.

This Terraform provider integrates **exclusively** with the [Keboola Management API](https://keboolamanagementapi.docs.apiary.io/#) using the official Keboola SDK (see `go.mod` for the exact version).

## Features
- Provider authentication via `url` and `token` (for the Management API).
- Ready for extension with Keboola Management API resources and data sources.

## API Token Generation
To use this provider, you need a Keboola Management API token. You can generate a token in the Keboola Connection Admin UI via your organization's admin interface or [URL](https://connection.HOSTNAME-SUFFIX.keboola.com/admin/account/access-tokens). For more details, see the [Keboola Management API documentation](https://keboolamanagementapi.docs.apiary.io/#).

## Usage Example
```hcl
provider "keboola" {
  url   = "https://connection.keboola.com" # Management API URL
  token = "your_management_api_token"      # Management API Token
}

resource "keboola_example" "test" {
  # Example resource usage (see below)
}
```

## Development
- Requires Go **1.24.3** or newer
- Uses [Terraform Plugin SDK v2](https://github.com/hashicorp/terraform-plugin-sdk) (v2.36.1)
- Implements Terraform provider protocol version **6.0** (see `terraform-registry-manifest.json`)
- Uses the official Keboola SDK (see `go.mod` for version)
- Tasks are managed via [Taskfile](https://taskfile.dev/)

## Compatibility
- This provider is compatible with Terraform protocol version **6.0** and requires Terraform CLI 1.5.0 or newer.

## Example Resource: keboola_example
This is a placeholder for the first resource using the Keboola SDK for the Management API. Replace this with actual resource logic as you expand the provider.

```hcl
resource "keboola_example" "test" {
  name = "example"
}
```

## Contributing
- Please follow Go best practices and Keboola conventions
- Add comments and documentation for all changes 