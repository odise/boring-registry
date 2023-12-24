terraform {
  required_version = ">=1.5"
  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 2.47.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">=3.11.0, <4.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
  #  backend "azurerm" {
  #    resource_group_name  = "<resource group name from outputs>"
  #    storage_account_name = "<storage account name from outputs>"
  #    container_name       = "boring-registry-resources-tfstate"
  #    key                  = "terraform.tfstate"
  #  }
}

provider "azuread" {}
provider "azurerm" {
  features {}
}
provider "random" {}