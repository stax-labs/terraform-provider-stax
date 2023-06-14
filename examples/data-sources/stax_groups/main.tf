terraform {
  required_providers {
    stax = {
      source = "registry.terraform.io/stax-labs/stax"
    }
  }
}

provider "stax" {
}

output "dedicated_dev_groups" {
  value = data.stax_groups.dedicated_dev
}
