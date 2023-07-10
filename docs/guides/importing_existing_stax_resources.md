---
layout: "stax"
page_title: "Importing Existing Stax Resources into Terraform"
description: A Guide to importing existing Stax Resources into Terraform
---

# Importing Existing Stax Resources into Terraform

One of the most common workflows when starting to manage Stax resources with Terraform is importing existing resources. This allows you to start managing resources that were already created within Stax.

This guide will walk through importing a [Stax Account](https://support.stax.io/hc/en-us/articles/4453778959503-About-Accounts) as an example.

## Importing an Account

With the release of [Terraform 1.5](https://www.hashicorp.com/blog/terraform-1-5-brings-config-driven-import-and-checks) Terraform introduced config-driven import. This means you can import resources by defining a block in configuration first.

To import a Stax Account into Terraform you first need to get the account ID. This can be found in the Stax Console under the Account details page. In the following block we define an `import` block for the account, passing the `id` of the Stax Account and a `to` pointing to the Terraform resource we want to generate.

Add the following code to a file name `import.tf`.

```terraform
import {
  id = "d531b5dd-bb09-4473-b022-d914d6a9572c"
  to = stax_account.innovation-dev
}
```

Next, run the import command:

```
terraform plan -generate-config-out=generated.tf
```

This command will generate a Terraform resource block for the imported resource and output it to `generated.tf`. The contents will look something like this:

```terraform
# __generated__ by Terraform
# Please review these resources and move them into your main configuration files.

# __generated__ by Terraform from "fb3cd13c-2bd0-46ed-881e-81c4e5912cf1"
resource "stax_account" "innovation-dev" {
  account_type_id   = "f7cb0794-cbaa-466d-888c-594b5222a6ec"
  aws_account_alias = null
  name              = "innovation-dev"
  tags = {
    costCentre = "phoenix-project"
  }
}
```

Move these resources to your `main.tf` file, then delete `generated.tf`. This will allow Terraform's `apply` command to apply them.

```terraform
# existing account resource imported into codebase
resource "stax_account" "innovation-dev" {
  account_type_id   = "f7cb0794-cbaa-466d-888c-594b5222a6ec"
  aws_account_alias = null
  name              = "innovation-dev"
  tags = {
    costCentre = "phoenix-project"
  }
}
```

Run Terraform Plan to preview any changes Terraform will make.

```
terraform plan
```

You should see the following summary at the end of the execution indicating that one resource is going to be imported into Terraform.

```
Plan: 1 to import, 0 to add, 0 to change, 0 to destroy.
```

Run `apply` to import the resource.

```
terraform apply
```

If the command is successful, you will see a result similar to the below:

```
Apply complete! Resources: 1 imported, 0 added, 0 changed, 0 destroyed.
```

## Recommendations

For those starting to use the Terraform provider, consider the following recommendations:

1. Just import the resources (Accounts, Account Types, Groups, etc.) that you need to manage. Don't import everything at once.
2. Keep import blocks in a separate `import.tf` file and generate the resources using `terraform plan -generate-config-out=generated.tf` then move them to your `main.tf`
3. Once imported, add tags to the resources to make them easier to manage and filter. Depending on your tagging strategy, consider adding a tag with a key of `managed_by` or similar and value of `terraform` to your resources. This will be useful later when filtering.

Now you're ready to start managing this Account resource with Terraform! Importing is a great way to get started with managing existing infrastructure. Let us know if you have any questions.