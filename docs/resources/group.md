---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "stax_group Resource - terraform-provider-stax"
subcategory: ""
description: |-
  Stax Group resource
---

# stax_group (Resource)

Stax Group resource

## Example Usage

```terraform
resource "stax_group" "cost-data-scientist" {
  name = "cost-data-scientist"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the account

### Read-Only

- `id` (String) Group identifier
- `type` (String) The type of Stax Group, this can be either `LOCAL` or `SCIM`. Note that groups with a type of `SCIM` cannot be modified by this provider.