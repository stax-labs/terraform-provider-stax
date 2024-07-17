---
layout: "stax"
page_title: "Managing Stax Resources with Terraform Cloud"
description: A Guide to managing Stax Resources with Terraform Cloud
---

# Managing Stax Resources with Terraform Cloud

[Terraform Cloud](https://www.terraform.io/) is a hosted service by HashiCorp that allows you to manage your Terraform configurations in a collaborative and secure manner. This guide will walk through how to manage Stax resources using Terraform Cloud.

Before you begin you should read through the [Get Started - Terraform Cloud](https://developer.hashicorp.com/terraform/tutorials/cloud-get-started) provided by hashicorp, this will familiarize you with the basics of Terraform Cloud.

**Please Note: The Stax Terraform Provider is no longer in Developer Preview and is now deprecated. Do not use this provider for production workloads.**

# Initial project setup

Before running Terraform, you'll need to:
1. Create a Terraform Cloud Organization and Workspace
2. Configure [Terraform Cloud workspace variables](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/variables) for your Stax credentials
3. Setup a git repository, which will store your Terraform configuration
4. [Connect the VCS Provider to Terraform Cloud](https://developer.hashicorp.com/terraform/cloud-docs/vcs)

# Configuring Credentials

A Stax [API Token](https://www.stax.io/developer/api-tokens/) is required to authenticate with the Stax API. You need to configure these credentials as [Terraform Cloud workspace variables](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/variables), using the values as follows:

| Key | Category | Sensitive |
|---|---|---|
| STAX_ACCESS_KEY | env | false |
| STAX_SECRET_KEY | env | true |

**Note:** 
1. The access key (`STAX_ACCESS_KEY`) is not considered `SENSITIVE` and enables customers to track which API token is being used in terraform cloud.
2. The secret key (`STAX_SECRET_KEY`) however **must** be marked as `SENSITIVE`.

# Provider configuration

Following the conventions in the terraform cloud documentation, you will define our provider configuration in the `provider.tf` file in our project repository:

```
terraform {
  cloud {
    organization = "ORGANIZATION_NAME"
    workspaces {
      name = "WORKSPACE_NAME"
    }
  }
  required_providers {
    stax = {
      source = "registry.terraform.io/stax-labs/stax"
    }
  }
  # enables the use of config-driven imports
  required_version = ">= 1.5.0"
}

# The credentials from the API token are exported as environment variables named as follows:
# * STAX_ACCESS_KEY
# * STAX_SECRET_KEY
# 
# This is typically how ci pipelines will run this infra code.
#
provider "stax" {
  installation = "au1"
}
```

You will need to replace `ORGANIZATION_NAME` with your organization name, as well as the `WORKSPACE_NAME` placeholder for workspace name, these are manually setup in terraform cloud when you [Create a workspace](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/creating).

Now you can add stax resources to the `main.tf` in your project and they will be provisioned in stax as a part of the Terraform Cloud workflow.