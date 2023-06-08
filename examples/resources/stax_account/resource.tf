variable "account_type_id" {
  description = "the account type identifier used for these accounts"
}

resource "stax_account" "presentation-dev" {
  name            = "presentation-dev"
  account_type_id = var.account_type_id
  tags = {
    "environment" : "production"
  }
}