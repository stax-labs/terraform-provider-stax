terraform {
  required_providers {
    stax = {
      source = "registry.terraform.io/stax-labs/stax"
    }
  }
}

provider "stax" {
}

output "stax_demo_api_tokens" {
  value = data.stax_api_tokens.stax_demo
}
