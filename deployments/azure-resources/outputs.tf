output "app_client_id" {
  value = azuread_application_registration.boring_registry.client_id
}
output "app_object_id" {
  value = azuread_application_registration.boring_registry.object_id
}
output "app_key_id" {
  value = azuread_application_password.boring_registry.id
}
output "app_password" {
  value     = azuread_application_password.boring_registry.value
  sensitive = true
}
output "current_subscription_display_name" {
  value = data.azurerm_subscription.current.display_name
}
output "current_subscription_id" {
  value = data.azurerm_subscription.current.id
}

output "storage_account_tfstate_id" {
  value = azurerm_storage_account.tfstate.id
}
output "storage_account_boring_registry_id" {
  value = azurerm_storage_account.boring_registry.id
}

output "resource_group_id" {
  value = azurerm_resource_group.boring_registry.id
}
output "terraform_state_container_name" {
  value = azurerm_storage_container.tfstate.name
}
output "boring_registry_container_name" {
  value = azurerm_storage_container.boring_registry_container.name
}
output "default_credentials" {
  value = "az login --service-principal -u ${azuread_application_registration.boring_registry.client_id} -p ${nonsensitive(azuread_application_password.boring_registry.value)} --tenant ${var.tenant_id}"
}