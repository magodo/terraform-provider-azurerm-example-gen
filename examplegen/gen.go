package examplegen

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/imports"

	"golang.org/x/tools/go/packages"
)

const (
	testFileGen = "example_gen_test.go"
	testCaseGen = "TestExampleGen"
)

type ExampleSource struct {
	RootDir    string
	ServiceDir string
	TestCase   string
}

func (src ExampleSource) GenExample() (string, error) {
	cfg := &packages.Config{Mode: packages.LoadAllSyntax, Tests: true, Dir: src.RootDir}
	pkgs, err := packages.Load(cfg, src.ServiceDir)
	if err != nil {
		return "", fmt.Errorf("loading package: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return "", fmt.Errorf("loading package contains error")
	}

	var (
		targetFdecl *ast.FuncDecl
		targetFile  *ast.File
	)
	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			for _, decl := range f.Decls {
				fdecl, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if fdecl.Name.Name != src.TestCase {
					continue
				}
				targetFdecl = fdecl
				targetFile = f
				break
			}
		}
	}

	if targetFdecl == nil {
		return "", fmt.Errorf("no test case named %q", src.TestCase)
	}

	// We keep the content of the target function body until the invocation of the "ResourceTest", which is typically doing the setup.
	var resourceTestCall *ast.CallExpr
	var stmts []ast.Stmt
	for _, stmt := range targetFdecl.Body.List {
		exprstmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			stmts = append(stmts, stmt)
			continue
		}
		callexpr, ok := exprstmt.X.(*ast.CallExpr)
		if !ok {
			stmts = append(stmts, stmt)
			continue
		}
		selexpr, ok := callexpr.Fun.(*ast.SelectorExpr)
		if !ok {
			stmts = append(stmts, stmt)
			continue
		}
		if selexpr.Sel.Name != "ResourceTest" {
			stmts = append(stmts, stmt)
			continue
		}
		resourceTestCall = callexpr
		break
	}

	if resourceTestCall == nil {
		return "", fmt.Errorf("no ResourceTest call found")
	}

	// Look for the first test step and record the assignment to the "Config", which is then used as the config generation invocation.
	if len(resourceTestCall.Args) != 3 {
		return "", fmt.Errorf("ResourceTest doesn't have 3 arguments")
	}
	teststeps, ok := resourceTestCall.Args[2].(*ast.CompositeLit)
	if !ok {
		return "", fmt.Errorf("test steps are not defined as a composite literal")
	}
	if len(teststeps.Elts) == 0 {
		return "", fmt.Errorf("there is no test step defined")
	}
	firstteststep, ok := teststeps.Elts[0].(*ast.CompositeLit)
	if !ok {
		return "", fmt.Errorf("the first test step is not defined as a composite literal")
	}

	var configcallexpr *ast.CallExpr
	for _, elt := range firstteststep.Elts {
		elt, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := elt.Key.(*ast.Ident)
		if !ok {
			continue
		}
		if key.Name != "Config" {
			continue
		}
		callexpr, ok := elt.Value.(*ast.CallExpr)
		if !ok {
			continue
		}
		configcallexpr = callexpr
		break
	}

	if configcallexpr == nil {
		return "", fmt.Errorf(`no "Config" field found for the first test step`)
	}

	stmts = append(stmts,
		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "fmt",
					},
					Sel: &ast.Ident{
						Name: "Println",
					},
				},
				Args: []ast.Expr{
					configcallexpr,
				},
			},
		},
	)

	fset := token.NewFileSet()
	targetFdecl.Body.List = stmts
	targetFdecl.Name.Name = testCaseGen
	targetFile.Decls = []ast.Decl{targetFdecl}

	buf := bytes.NewBuffer([]byte{})
	if err := printer.Fprint(buf, fset, targetFile); err != nil {
		return "", fmt.Errorf("printing the source code: %v", err)
	}

	testFilePath := filepath.Join(src.RootDir, src.ServiceDir, testFileGen)
	content, err := imports.Process(testFilePath, buf.Bytes(), &imports.Options{Comments: false})
	if err != nil {
		return "", fmt.Errorf("imports processing the source code: %v", err)
	}

	testFile, err := os.Create(testFilePath)
	if err != nil {
		return "", fmt.Errorf("creating test file: %w", err)
	}
	defer os.Remove(testFilePath)
	if _, err := testFile.Write(content); err != nil {
		return "", fmt.Errorf("writing to the test file %q: %v", testFilePath, err)
	}

	// Run the test and fetch the printed Terraform configuration.
	cmd := exec.Command("go", "test", "-v", "-run="+testCaseGen)
	cmd.Dir = filepath.Join(src.RootDir, src.ServiceDir)

	// The acceptance.BuildTestData depends on the following environment variables:
	// - ARM_TEST_LOCATION
	// - ARM_TEST_LOCATION_ALT1
	// - ARM_TEST_LOCATION_ALT2
	cmd.Env = append(os.Environ(),
		"ARM_TEST_LOCATION=West Europe",
		"ARM_TEST_LOCATION_ALT1=East US",
		"ARM_TEST_LOCATION_ALT2=East US2",
	)

	var (
		stdout = bytes.NewBuffer([]byte{})
		stderr = bytes.NewBuffer([]byte{})
	)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		msg := fmt.Sprintf("running the test: %v", err)
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			msg += "\n" + stderr.String()
		}
		return "", fmt.Errorf(msg)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	var active bool
	var results []string
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "=== RUN"):
			active = true
			continue
		case strings.HasPrefix(line, "--- PASS"):
			active = false
			continue
		}
		if !active || line == "" {
			continue
		}
		results = append(results, line)
	}

	return strings.Join(results, "\n"), nil
}
