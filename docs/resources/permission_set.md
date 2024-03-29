---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "stax_permission_set Resource - terraform-provider-stax"
subcategory: ""
description: |-
  Provides a Stax Permission Set resource. Permission Sets https://support.stax.io/hc/en-us/articles/4453967433359-Permission-Sets allow customers to define their own AWS Access permissions, to which AWS accounts they apply and the groups of users who are subsequently granted access.
---

# stax_permission_set (Resource)

Provides a Stax Permission Set resource. [Permission Sets](https://support.stax.io/hc/en-us/articles/4453967433359-Permission-Sets) allow customers to define their own AWS Access permissions, to which AWS accounts they apply and the groups of users who are subsequently granted access.

## Example Usage

```terraform
resource "stax_permission_set" "data-scientist" {
  name                 = "data-scientist"
  max_session_duration = 28800
  description          = "Data Scientist Role. "
  aws_managed_policy_arns = [
    "arn:aws:iam::aws:policy/job-function/DataScientist",
    "arn:aws:iam::aws:policy/AWSBillingReadOnlyAccess"
  ]
  tags = {
    "owner" : "stax-demo@stax.io"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the stax Permission Set

### Optional

- `aws_managed_policy_arns` (List of String) A list of aws managed policy arns assigned to the Permission Set, see [aws managed policies](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_managed-vs-inline.html#aws-managed-policies) documentation for more information
- `description` (String) The description of the stax Permission Set
- `inline_policies` (Set of Object) The inline policies assigned to the Permission Set (see [below for nested schema](#nestedatt--inline_policies))
- `max_session_duration` (Number) The max session duration in seconds, used by this Permission Set when creating the AWS IAM role
- `tags` (Map of String) Permission Set tags

### Read-Only

- `created_by` (String) The identifier of the stax user who created the Permission Set
- `created_ts` (String) The Permission Set was creation timestamp
- `id` (String) Permission Set identifier
- `status` (String) The status of the stax Permission Set, can be ACTIVE or DELETED

<a id="nestedatt--inline_policies"></a>
### Nested Schema for `inline_policies`

Optional:

- `name` (String)
- `policy` (String)
