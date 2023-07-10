---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "stax_api_token Resource - terraform-provider-stax"
subcategory: ""
description: |-
  Provides a Stax API Token resource. Stax API Token https://support.stax.io/hc/en-us/articles/4447315161231-About-API-Tokens are security credentials that can be used to authenticate to the Stax API. Stax will create and store them securely in a customers security AWS account using Systems Manager Parameter Store https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html
---

# stax_api_token (Resource)

Provides a Stax API Token resource. [Stax API Token](https://support.stax.io/hc/en-us/articles/4447315161231-About-API-Tokens) are security credentials that can be used to authenticate to the Stax API. Stax will create and store them securely in a customers security AWS account using [Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)

## Example Usage

```terraform
resource "stax_user" "cost-data-scientist" {
  name = "cost-data-scientist"
  role = "api_admin"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The first name of the stax api token
- `role` (String) The role of the stax api token, this can be one of `api_admin`, `api_user`, or `api_readonly`

### Optional

- `tags` (Map of String) The tags associated with the stax api token

### Read-Only

- `created_ts` (String) The created timestamp for the stax stax api
- `id` (String) API Token identifier
- `modified_ts` (String) The modified timestamp for the stax stax api
- `status` (String) The status of the stax stax api