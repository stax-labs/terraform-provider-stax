# Stax Terraform Provider (Developer Preview)

This is the official Terraform provider for [Stax](https://www.stax.io/). 

**Please Note: The Stax Terraform Provider is currently in Developer Preview and is intended strictly for feedback purposes only. Do not use this provider for production workloads.**

# Usage

NOTE: This provider is built with the assumption that you are a Stax customer and are familiar with how to create and access [API Tokens](https://www.stax.io/developer/api-tokens/).


```terraform
variable "installation" {
  description = "installation name"
}

variable "api_token_access_key" {
  description = "api token access key"
}

variable "api_token_secret_key" {
  description = "api token secret key"
}

terraform {
  required_providers {
    stax = {
      source = "registry.terraform.io/stax/stax"
    }
  }
}

provider "stax" {
  installation         = var.installation
  api_token_access_key = var.api_token_access_key
  api_token_secret_key = var.api_token_secret_key
}
```

# Development

To provide the required secrets during development and integration testing some environment variables are required to run examples. These can be configured using an `.envrc` file which is loaded by [direnv](https://direnv.net/).

Example `.envrc` contents.

```bash
export STAX_ACCESS_KEY=whatever_access_key
export STAX_SECRET_KEY=whatever_secret_key
export STAX_INSTALLATION=au1
export TF_LOG=INFO

# used to test importing a stax account
export IMPORT_STAX_ACCOUNT_ID=whatever_uuid

# used to test creating/updating a stax account
export ACCOUNT_TYPE_ID=whatever_uuid
```

# Contributing

For more information on contributing the to the Stax Go SDK, please see our [guide](https://github.com/stax-labs/terraform-provider-stax/blob/master/CONTRIBUTING.md).

# Getting Help

* If you're having trouble using the Stax SDK, please refer to our [documentation](https://www.stax.io/developer/api-tokens/).<br>
* If you've encountered an issue or found a bug, please [open an issue](https://github.com/stax-labs/stax-golang-sdk/issues).<br>
* For any other requests, please contact [Stax support](mailto:support@stax.io).