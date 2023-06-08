variable "group_id" {
  description = "the group identifier used to filter groups list"
}

data "stax_groups" "dedicated_dev" {
  filters = {
    ids = [var.group_id]
  }
}
