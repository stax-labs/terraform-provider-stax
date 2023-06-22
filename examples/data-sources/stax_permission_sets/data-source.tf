data "stax_permission_sets" "stax-dev" {
  filters = {
    statuses = ["ACTIVE"]
  }
}
