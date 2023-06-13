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

output "dedicated_dev_account_types" {
  value = data.stax_account_types.dedicated_dev
}
