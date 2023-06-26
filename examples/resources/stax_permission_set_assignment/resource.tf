variable "permission_set_id" {
  description = "the permission set identifier used for this assignment"
}

variable "group_id" {
  description = "the group identifier used for this assignment"
}

variable "account_type_id" {
  description = "the account type identifier used for this assignment"
}

resource "stax_permission_set_assignment" "data-scientist-production" {
  permission_set_id = var.permission_set_id
  group_id          = var.group_id
  account_type_id   = var.account_type_id
}
