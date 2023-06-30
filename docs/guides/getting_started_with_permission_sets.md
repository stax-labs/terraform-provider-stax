---
layout: "stax"
page_title: "Getting Started with Permission Sets"
description: A Getting Started with Permission Sets Guide
---

# Getting Started with Permission Sets Guide

Permission Sets are a feature of Stax, this guide was written to provide a walk through illustrating how they can be managed using terraform. Customers are encouraged to read through the [Stax Permission Sets documentation](https://support.stax.io/hc/en-us/articles/4453967433359-Permission-Sets) first.

## Create a Permission Set

The following terraform code will create a new permission named `phoenix-project-access` which is using the AWS Managed Policy `job-function/DataScientist`. Stax team recommends wherever possible customers should use the AWS Managed Policies.

```terraform
resource "stax_permission_set" "phoenix-project-access" {
  aws_managed_policy_arns = [
    "arn:aws:iam::aws:policy/job-function/DataScientist",
  ]
  description          = "Phoenix Project Access."
  inline_policies      = null
  max_session_duration = 28800 # 8 hours
  name                 = "phoenix-project-access"
  tags = {
    costCentre = "phoenix-project"
  }
  lifecycle {
    create_before_destroy = true
  }
}
```

## Create a Group

To create an Permission Set we need a group, the following terraform code will create a new one for our `phoenix-project-team`.

```terraform
resource "stax_group" "phoenix-project-team" {
  name = "phoenix-project-team"
  lifecycle {
    create_before_destroy = true
  }
}
```

## Create an Account Type

In Stax Account Types allow customers to describe a collection of accounts, then associate these to other constructs with in Stax. The following terraform code will create a new one for our `phoenix-project-accounts`.

```terraform
resource "stax_account_type" "phoenix-project-accounts" {
  name = "phoenix-project"
  lifecycle {
    create_before_destroy = true
  }
}
```

## Assign the Account Type to an Account

To ensure access is granted to accounts we need to associate a Stax Account to the Account Type. The following terraform code illustrates how this can be done.

```terraform
resource "stax_account" "innovation-dev" {
  account_type_id   = stax_account_type.phoenix-project-accounts.id # association of the new account type
  aws_account_alias = null
  name              = "innovation-dev"
  tags = {
    costCentre = "phoenix-project"
  }
}
```

## Create a Permission Set Assignment

The Stax Permission Set Assignment links all these resources together and provides Stax with the information required to deploy roles to a given set of AWS accounts. The following terraform code will create a new one linking our Permission Set to our Group, and Account Type.

```terraform
resource "stax_permission_set_assignment" "phoenix-project-access_phoenix-project-accounts_phoenix-project-team" {
  account_type_id   = stax_account_type.phoenix-project-accounts.id
  group_id          = stax_group.phoenix-project-team.id
  permission_set_id = stax_permission_set.phoenix-project-access.id
  lifecycle {
    create_before_destroy = true
  }
}
```

## Applying the Changes

To create the resources we have added to our terraform configuration we simply run `terraform apply`.
