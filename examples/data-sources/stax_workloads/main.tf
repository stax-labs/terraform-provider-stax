terraform {
  required_providers {
    stax = {
      source = "registry.terraform.io/stax-labs/stax"
    }
  }
}

provider "stax" {
}

output "stax-demo_groups" {
  value = data.stax_workloads.stax-demo
}
