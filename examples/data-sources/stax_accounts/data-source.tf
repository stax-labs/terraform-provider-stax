data "stax_groups" "dedicated_dev" {
  filters = {
    names = ["presentation-dev"]
  }
}
