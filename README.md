## Intro

`terraform-provider-azurerm-example-gen` is a tool used to output the Terraform configuration example based on the acceptance tests. As the name implies, this tool only serves for the github.com/hashicorp/terraform-provider-azurerm project.

## Install

```
$ go install github.com/magodo/terraform-provider-azurerm-example-gen
```

Note that this tool depends on `go` command. So make sure you have `go` installed in your runtime environment.

## Example

```
$ terraform-provider-azurerm-example-gen ~/github/terraform-provider-azurerm ./internal/services/network TestAccSubnet_basic
```

## How it Works

Given the fixed code structure of the acctest in the provider codebase, the tool fetches out the setup part of a specified testcase (i.e. the construction of the testdata and the resource type), then add a following `fmt.println` statement to print the first test step's configuration to the stdout. With wrapping this as a new test case, the tool will then invoke `go test -v` internally and capture the stdout content, which is the terraform configuration that the acctest runs.
