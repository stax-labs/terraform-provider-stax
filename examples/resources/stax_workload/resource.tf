variable "account_id" {
  description = "the account identifier used for this workload"
}

variable "catalog_id" {
  description = "the catalog identifier used for this workload"
}

variable "region" {
  description = "the region used for this workload"
}

variable "juma_root_account_id" {
  description = "the juma root account id used for this workload"
}

resource "stax_workload" "shared-workloads-bucket" {
  name       = "shared-workloads-bucket"
  account_id = var.account_id
  catalog_id = var.catalog_id
  region     = var.region

  parameters = [
    {
      "key" : "JumaRootAccountId", "value" : var.juma_root_account_id
    },
    {
      "key" : "JumaWorkloadName", "value" : "shared-workloads-bucket"
    }
  ]
  tags = {
    "provisioned_by" : "terraform"
  }
}
