terraform {
  required_providers {
    stax = {
      source = "registry.terraform.io/stax-labs/stax"
    }
  }
}

provider "stax" {
}

output "stax-dev_permission" {
  value = data.stax_permission_sets.stax-dev
}
