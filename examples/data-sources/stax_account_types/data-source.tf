variable "account_type_id" {
  description = "the account type identifier used for these accounts"
}

data "stax_account_types" "dedicated_dev" {
  filters = {
    ids = [var.account_type_id]
  }
}
