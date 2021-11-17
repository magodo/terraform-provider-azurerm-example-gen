package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/magodo/terraform-provider-azurerm-example-gen/examplegen"
)

const usage = `Generate example configuration for Terraform AzureRM provider from its AccTest.

Usage: terraform-provider-azurerm-example-gen -dir=rootdir -from=testname_regexp servicepkg...

Example: terraform-provider-azurerm-example-gen -dir=$HOME/github/terraform-provider-azurerm -from='TestAccSubnet_basic$' ./internal/services/network 
`

var (
	fromFlag = flag.String("from", "", "From which Acceptance test case")
	dirFlag  = flag.String("dir", ".", "The pat to the root directory of the provider code base")
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, usage)
		os.Exit(1)
	}
	if *fromFlag == "" {
		fmt.Fprintf(os.Stderr, usage)
		os.Exit(1)
	}

	src := examplegen.ExampleSource{
		RootDir:     *dirFlag,
		TestCase:    *fromFlag,
		ServicePkgs: args,
	}
	example, err := src.GenExample()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Printf(example)
}
