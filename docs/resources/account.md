---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "stax_account Resource - terraform-provider-stax"
subcategory: ""
description: |-
  Account resource. Stax Accounts https://support.stax.io/hc/en-us/articles/4453778959503-About-Accounts allows you to securely and easily create, view and centrally manage your AWS Accounts and get started deploying applications.
---

# stax_account (Resource)

Account resource. [Stax Accounts](https://support.stax.io/hc/en-us/articles/4453778959503-About-Accounts) allows you to securely and easily create, view and centrally manage your AWS Accounts and get started deploying applications.

## Example Usage

```terraform
variable "account_type_id" {
  description = "the account type identifier used for these accounts"
}

resource "stax_account" "presentation-dev" {
  name            = "presentation-dev"
  account_type_id = var.account_type_id
  tags = {
    "environment" : "production"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the stax account

### Optional

- `account_type_id` (String) The account type identifier for the stax account
- `aws_account_alias` (String) The aws account alias for the stax account
- `tags` (Map of String) The tags associated with the stax account

### Read-Only

- `account_type` (String) The account type for the stax account
- `aws_account_id` (String) The aws account identifier for the stax account
- `id` (String) Account identifier
- `status` (String) Account Status
