package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `Generate example configuration for Terraform AzureRM provider from its AccTest.

Usage: terraform-provider-azurerm-example-gen rootdir servicedir testname

Example: terraform-provider-azurerm-example-gen $HOME/github/terraform-provider-azurerm ./internal/services/network TestAccSubnet_basic
`

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, usage)
		os.Exit(1)
	}

	src := ExampleSource{
		RootDir:    args[0],
		ServiceDir: args[1],
		TestCase:   args[2],
	}
	example, err := src.GenExample()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println(example)
}
