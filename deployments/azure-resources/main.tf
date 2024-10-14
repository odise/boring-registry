resource "random_id" "rg_name" {
  byte_length = 8
}
# Azure resource group
resource "azurerm_resource_group" "boring_registry" {
  location = "germanywestcentral"
  name     = "terraform-essentials-${random_id.rg_name.hex}-rg"
}

# service principal
resource "azuread_application_registration" "boring_registry" {
  display_name            = "Boring Registry"
  description             = "Boring Registry TF Deployment Application"
  sign_in_audience        = "AzureADMyOrg"
  group_membership_claims = ["All"]
}
resource "azuread_application_password" "boring_registry" {
  application_id = azuread_application_registration.boring_registry.id
  #  application_object_id = null
}

data "azuread_client_config" "current" {}
resource "azuread_service_principal" "boring_registry" {
  client_id                    = azuread_application_registration.boring_registry.client_id
  app_role_assignment_required = false

  owners = [
    data.azuread_client_config.current.object_id,
  ]
  lifecycle {
    ignore_changes = [
      owners,
    ]
  }
}

data "azurerm_subscription" "current" {}
resource "azurerm_role_assignment" "boring_registry" {
  scope                = data.azurerm_subscription.current.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = azuread_service_principal.boring_registry.object_id
}

# Storage
resource "random_id" "suffix" {
  byte_length = 2
}
# tf state
resource "azurerm_storage_account" "tfstate" {
  name                              = "terraformtfstate${random_id.suffix.hex}"
  resource_group_name               = azurerm_resource_group.boring_registry.name
  location                          = azurerm_resource_group.boring_registry.location
  account_tier                      = "Standard"
  account_replication_type          = "LRS"
  allow_nested_items_to_be_public   = false
  infrastructure_encryption_enabled = true

  tags = {}
}
resource "azurerm_storage_container" "tfstate" {
  name                  = "boring-registry-resources-tfstate"
  storage_account_name  = azurerm_storage_account.tfstate.name
  container_access_type = "private"
}

# boring registry BlobStorage
resource "azurerm_storage_account" "boring_registry" {
  name                              = "boringregistry${random_id.suffix.hex}"
  resource_group_name               = azurerm_resource_group.boring_registry.name
  location                          = azurerm_resource_group.boring_registry.location
  account_tier                      = "Standard"
  account_replication_type          = "LRS"
  allow_nested_items_to_be_public   = false
  infrastructure_encryption_enabled = true

  tags = {}
}
resource "azurerm_storage_container" "boring_registry_container" {
  name                  = "boring-registry-container"
  storage_account_name  = azurerm_storage_account.boring_registry.name
  container_access_type = "private"
}
