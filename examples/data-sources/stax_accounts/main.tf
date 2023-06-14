terraform {
  required_providers {
    stax = {
      source = "registry.terraform.io/stax-labs/stax"
    }
  }
}

provider "stax" {
}

output "dedicated_dev_accounts" {
  value = data.stax_accounts.dedicated_dev
}
