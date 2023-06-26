variable "permission_set_id" {
  description = "the permission set identifier used for these assignments"
}


data "stax_permission_set_assignments" "stax-dev" {
  permission_set_id = var.permission_set_id
  filters = {
    statuses = ["DEPLOYMENT_COMPLETE"]
  }
}
