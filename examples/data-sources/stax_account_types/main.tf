terraform {
  required_providers {
    stax = {
      source = "registry.terraform.io/stax-labs/stax"
    }
  }
}

provider "stax" {
}

output "dedicated_dev_account_types" {
  value = data.stax_account_types.dedicated_dev
}
