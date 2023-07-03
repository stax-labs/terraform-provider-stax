variable "group_id" {
  description = "the group identifier to associate with these users"
}

variable "user_id" {
  description = "the identifier of the user which is a member of this group"
}

resource "stax_group_membership" "cost-data-scientist" {
  id = var.group_id
  user_ids = [
    var.user_id
  ]
}
