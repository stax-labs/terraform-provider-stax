variable "user_id" {
  description = "the user identifier used to filter users list"
}

data "stax_users" "stax_demo" {
  id = var.user_id
  # filters = {
  #   ids = [var.user_id]
  # }
}
