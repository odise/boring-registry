# `enercity-prod` stack

## `pre-commit` hooks

Make sure you have [pre-commit](https://pre-commit.com/) installed. Run `pre-commit install` to activate pre-commit 
hooks for this repository.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >=1.5 |
| <a name="requirement_azuread"></a> [azuread](#requirement\_azuread) | ~> 2.46.0 |
| <a name="requirement_azurerm"></a> [azurerm](#requirement\_azurerm) | >=3.11.0, <4.0 |
| <a name="requirement_random"></a> [random](#requirement\_random) | ~> 3.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_azuread"></a> [azuread](#provider\_azuread) | ~> 2.46.0 |
| <a name="provider_azurerm"></a> [azurerm](#provider\_azurerm) | >=3.11.0, <4.0 |
| <a name="provider_random"></a> [random](#provider\_random) | ~> 3.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [azuread_application_password.atlantis](https://registry.terraform.io/providers/hashicorp/azuread/latest/docs/resources/application_password) | resource |
| [azuread_application_registration.atlantis](https://registry.terraform.io/providers/hashicorp/azuread/latest/docs/resources/application_registration) | resource |
| [azuread_service_principal.atlantis](https://registry.terraform.io/providers/hashicorp/azuread/latest/docs/resources/service_principal) | resource |
| [azurerm_resource_group.terraform_essentials](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group) | resource |
| [azurerm_role_assignment.atlantis](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/role_assignment) | resource |
| [azurerm_storage_account.tfstate](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/storage_account) | resource |
| [azurerm_storage_container.tfstate](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/storage_container) | resource |
| [random_id.rg_name](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/id) | resource |
| [azuread_client_config.current](https://registry.terraform.io/providers/hashicorp/azuread/latest/docs/data-sources/client_config) | data source |

## Inputs

No inputs.

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_atlantis_app_client_id"></a> [atlantis\_app\_client\_id](#output\_atlantis\_app\_client\_id) | n/a |
| <a name="output_atlantis_app_key_id"></a> [atlantis\_app\_key\_id](#output\_atlantis\_app\_key\_id) | n/a |
| <a name="output_atlantis_app_object_id"></a> [atlantis\_app\_object\_id](#output\_atlantis\_app\_object\_id) | n/a |
| <a name="output_atlantis_app_password"></a> [atlantis\_app\_password](#output\_atlantis\_app\_password) | n/a |
<!-- END_TF_DOCS -->