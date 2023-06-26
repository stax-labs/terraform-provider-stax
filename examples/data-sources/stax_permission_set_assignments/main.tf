terraform {
  required_providers {
    stax = {
      source = "registry.terraform.io/stax-labs/stax"
    }
  }
}

provider "stax" {
}

output "stax-dev_permission_set_assignments" {
  value = data.stax_permission_set_assignments.stax-dev
}
