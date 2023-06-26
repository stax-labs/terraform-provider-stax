variable "workload_id" {
  description = "the workload identifier used to filter workloads list"
}

data "stax_workloads" "stax-demo" {
  filters = {
    ids = [var.workload_id]
  }
}
