# Stax Terraform Provider (Developer Preview)

This is the official Terraform provider for [Stax](https://www.stax.io/). 

**Please Note: The Stax Terraform Provider is currently in Developer Preview and is intended strictly for feedback purposes only. Do not use this provider for production workloads.**

# Usage

NOTE: This provider is built with the assumption that you are a Stax customer and are familiar with how to create and access [API Tokens](https://www.stax.io/developer/api-tokens/).


```terraform
variable "installation" {
  description = "Stax Short Installation ID for your Stax tenancy's control plane"
}

variable "api_token_access_key" {
  description = "Stax API Token Access Key"
}

variable "api_token_secret_key" {
  description = "Stax API Token Secret Key"
}

terraform {
  required_providers {
    stax = {
      source = "registry.terraform.io/stax-labs/stax"
    }
  }
}

provider "stax" {
  installation         = var.installation
  api_token_access_key = var.api_token_access_key
  api_token_secret_key = var.api_token_secret_key
}
```

The provider can also be configured using the equivalent environment variables as per the [Stax Provider](docs/index.md).

# Supported Resources

| Type | Resource | Data Source
|---|---|---|
| Account | ✅ | ✅ 
| AccountType | ✅ | ✅
| Permission Sets |
| Policies |
| APIToken |
| User | 
| Group | ✅ | ✅
| NetworkHub |
| Workload |
| Workload Manifest |

# Limitations 

1. The Terraform data sources currently only utilize the first page from the Stax API. However, this will change once we develop a strategy to manage large results while minimizing the impact on the Stax API.

# Development

To build the terraform provider locally you will require.

* Terraform >= 1.0
* Go >= 1.20

To build and install the provider use `make install`.

To run the acceptance tests use `make testacc`

To manually test the examples you will need to the provider path and setup some environment variables, then you will be able to use the targets in the `GNUMakefile` for testing.
## Overriding the provider for local development

You'll need to add a provider override to your `~/.terraformrc`. Documentation around using this file can be found [here](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers). See below for an example of a `.terraformrc` file set up for plugin override.

Get the path to the `bin` directory for your `GOPATH` used by the Go SDK to store tools and cache packages. The following handy shell command below will emit the correct value on Linux or OSX.

```shell
echo $(go env GOPATH)\bin
```

Update the `CHANGE_ME_TO_GO_BIN` to `bin` directory for your `GOPATH`.

```hcl
provider_installation {

        dev_overrides {
                "stax-labs/stax" = "CHANGE_ME_TO_GO_BIN"
        }

        direct {}
}

```

## Manual Testing Environment Variables

To provide the required secrets during development and integration testing some environment variables are required to run examples. These can be configured using an `.envrc` file which is loaded by [direnv](https://direnv.net/).

Example `.envrc` contents.

```bash
export STAX_ACCESS_KEY=whatever_access_key
export STAX_SECRET_KEY=whatever_secret_key
export STAX_INSTALLATION=au1
export TF_LOG=INFO

# used to test importing a stax account
export IMPORT_STAX_ACCOUNT_ID=whatever_uuid

# used to test importing a stax account type
export IMPORT_STAX_ACCOUNT_TYPE_ID=whatever_uuid

# used to test creating/updating a stax account
export ACCOUNT_TYPE_ID=whatever_uuid
```

# Contributing

For more information on contributing the to the Stax Go SDK, please see our [guide](https://github.com/stax-labs/terraform-provider-stax/blob/master/CONTRIBUTING.md).

# Getting Help

* If you're having trouble using the Stax SDK, please refer to our [documentation](https://www.stax.io/developer/api-tokens/).<br>
* If you've encountered an issue or found a bug, please [open an issue](https://github.com/stax-labs/terraform-provider-stax/issues).<br>
* For any other requests, please contact [Stax support](mailto:support@stax.io).