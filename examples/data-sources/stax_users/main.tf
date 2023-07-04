terraform {
  required_providers {
    stax = {
      source = "registry.terraform.io/stax-labs/stax"
    }
  }
}

provider "stax" {
}

output "stax_demo_users" {
  value = data.stax_users.stax_demo
}
