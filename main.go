package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/magodo/terraform-provider-azurerm-example-gen/examplegen"
)

const usage = `Generate example configuration for Terraform AzureRM provider from its AccTest.

Usage: terraform-provider-azurerm-example-gen rootdir servicepkg testname

Example: terraform-provider-azurerm-example-gen $HOME/github/terraform-provider-azurerm ./internal/services/network TestAccSubnet_basic
`

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, usage)
		os.Exit(1)
	}

	src := examplegen.ExampleSource{
		RootDir:    args[0],
		ServicePkg: args[1],
		TestCase:   args[2],
	}
	example, err := src.GenExample()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Printf(example)
}
