## Intro

`terraform-provider-azurerm-example-gen` is a tool used to output the Terraform configuration example based on the acceptance tests. As the name implies, this tool only serves for the github.com/hashicorp/terraform-provider-azurerm project.

## Install

```bash
$ go install github.com/magodo/terraform-provider-azurerm-example-gen
```

Note that this tool depends on `go` command. So make sure you have `go` installed in your runtime environment.

## Usage

```bash
$ terraform-provider-azurerm-example-gen -from=testname rootdir servicepkg...
```

Note that the invocation convention of this tool is similar to `go test`, where the `-from` supports suing regexp to specify the acctest names and the `servicepkg...` can be one or more packages (i.e. package-list mode), either in the directory form or import path form.

## Example

```bash
$ terraform-provider-azurerm-example-gen -dir=$HOME/github/terraform-provider-azurerm -from='TestAccSubnet_basic$' ./internal/services/network
# or
$ terraform-provider-azurerm-example-gen -dir=$HOME/github/terraform-provider-azurerm -from='TestAccSubnet_basic$' github.com/hashicorp/terraform-provider-azurerm/internal/services/network
```

Output:

```hcl
# Generated from AccTest TestAccSubnet_basic

provider "azurerm" {
  features {}
}
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-211102165905940679"
  location = "West Europe"
}
resource "azurerm_virtual_network" "test" {
  name                = "acctestvirtnet211102165905940679"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}
resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}
resource "azurerm_subnet" "test2" {
  name                 = "internal2"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.0.3.0/24"
}
```

## How it Works

Given the fixed code structure of the acctest in the provider codebase, the tool fetches out the setup part of a specified testcase (i.e. the construction of the testdata and the resource type), then add a following `fmt.println` statement to print the first test step's configuration to the stdout. With wrapping this as a new test case, the tool will then invoke `go test -v` internally and capture the stdout content, which is the terraform configuration that the acctest runs.
