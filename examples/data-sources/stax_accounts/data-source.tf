data "stax_accounts" "dedicated_dev" {
  filters = {
    names = ["presentation-dev"]
  }
}
