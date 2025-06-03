package utils

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/stoewer/go-strcase"
)

// CreateSchema generates a JSON schema from any Go data structure,
// automatically detecting the module path and source location to include
// Go comments as descriptions in the schema.
func CreateSchema(dataStructure any) (map[string]any, error) {
	// Get the type information
	t := reflect.TypeOf(dataStructure)
	if t == nil {
		return nil, fmt.Errorf("dataStructure cannot be nil")
	}

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Create reflector

	r := &jsonschema.Reflector{}
	r.KeyNamer = strcase.SnakeCase

	// Try to add Go comments automatically

	if err := addGoCommentsAuto(r, t); err != nil {
		// If auto-detection fails, log the error but continue without comments

		fmt.Printf("Warning: Could not auto-detect source location for comments: %v\n", err)
	} else {
	}

	// Generate the schema

	schema := r.Reflect(dataStructure)
	if schema == nil {
		return nil, fmt.Errorf("failed to generate schema")
	}

	// Convert to map[string]any

	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	var result map[string]any

	if err := json.Unmarshal(schemaBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema to map: %w", err)
	}

	return result, nil
}

// addGoCommentsAuto automatically detects the module path and source directory
// for the given type and adds Go comments to the reflector.
func addGoCommentsAuto(r *jsonschema.Reflector, t reflect.Type) error {
	// Get the package path from reflection
	pkgPath := t.PkgPath()

	if pkgPath == "" {
		return fmt.Errorf("type has no package path (built-in type?)")
	}

	var sourceDir string
	var err error

	// Handle main package as a special case using AST parsing
	if pkgPath == "main" {

		sourceDir, err = findStructDefinitionInMain(t.Name())
		if err != nil {
			return fmt.Errorf("failed to find struct in main package: %w", err)
		}

	} else {
		// Find the actual source directory for this package

		sourceDir, err = findSourceDirectory(pkgPath)
		if err != nil {
			return fmt.Errorf("failed to find source directory: %w", err)
		}

	}

	// Verify the source directory exists

	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", sourceDir)
	}

	// Based on the internal logic of AddGoComments:
	// - It uses gopath.Join(base, path) to create package identifiers
	// - It walks the filesystem from 'path' and for each directory,
	//   creates a key using gopath.Join(base, relativePath)
	//
	// Strategy: Use the package path as base and "." as path
	// This way, when it processes the current directory,
	// gopath.Join(pkgPath, ".") will equal pkgPath

	// Change to the source directory temporarily to make relative path work
	oldDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(sourceDir); err != nil {
		return fmt.Errorf("failed to change to source directory: %w", err)
	}

	// Ensure we change back
	defer func() {
		os.Chdir(oldDir)
	}()

	// Now call AddGoComments with the package path as base and current dir as path

	if err := r.AddGoComments(pkgPath, "."); err != nil {
		return fmt.Errorf("failed to add Go comments: %w", err)
	}

	return nil
}

// findSourceDirectory finds the actual filesystem directory containing the source
// code for the given package path.
func findSourceDirectory(pkgPath string) (string, error) {
	// Try using build.Import first
	pkg, err := build.Import(pkgPath, "", build.FindOnly)
	if err == nil && pkg.Dir != "" {
		return pkg.Dir, nil
	}

	// Fallback: try to construct the path from module info

	moduleRoot, moduleName, err := findModuleInfo(pkgPath)
	if err != nil {
		return "", fmt.Errorf("failed to find module info: %w", err)
	}

	// Calculate relative path within the module
	relativePath := strings.TrimPrefix(pkgPath, moduleName)
	relativePath = strings.TrimPrefix(relativePath, "/")

	if relativePath == "" {
		return moduleRoot, nil
	}

	sourceDir := filepath.Join(moduleRoot, relativePath)

	return sourceDir, nil
}

// findModuleInfo finds the module root directory and module name for a given package path.
// It works by walking up the directory tree looking for go.mod files.
func findModuleInfo(pkgPath string) (moduleRoot, moduleName string, err error) {
	// Try to find the package in GOPATH or module cache first
	pkg, err := build.Import(pkgPath, "", build.FindOnly)
	if err != nil {
		// If build.Import fails, try using runtime info
		return findModuleInfoFromRuntime(pkgPath)
	}

	// Start from the package directory and walk up
	dir := pkg.Dir

	for {
		goModPath := filepath.Join(dir, "go.mod")

		if _, err := os.Stat(goModPath); err == nil {

			// Found go.mod, read the module name
			moduleName, err := readModuleName(goModPath)
			if err != nil {
				return "", "", fmt.Errorf("failed to read module name from %s: %w", goModPath, err)
			}

			return dir, moduleName, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory
			break
		}
		dir = parent
	}

	return "", "", fmt.Errorf("no go.mod found in any parent directory of %s", pkg.Dir)
}

// findModuleInfoFromRuntime uses runtime information to find module info
// when build.Import fails (e.g., for embedded packages or special cases).
func findModuleInfoFromRuntime(pkgPath string) (string, string, error) {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Walk up from current directory to find go.mod
	dir := wd

	for {
		goModPath := filepath.Join(dir, "go.mod")

		if _, err := os.Stat(goModPath); err == nil {

			moduleName, err := readModuleName(goModPath)
			if err != nil {
				return "", "", fmt.Errorf("failed to read module name: %w", err)
			}

			return dir, moduleName, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Try using go list as a fallback

	return findModuleInfoWithGoList()
}

// findModuleInfoWithGoList uses 'go list' command to find module information.
func findModuleInfoWithGoList() (string, string, error) {
	// Get module root

	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to run 'go list -m': %w", err)
	}
	moduleRoot := strings.TrimSpace(string(output))

	// Get module name

	cmd = exec.Command("go", "list", "-m", "-f", "{{.Path}}")
	output, err = cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get module name: %w", err)
	}
	moduleName := strings.TrimSpace(string(output))

	return moduleRoot, moduleName, nil
}

// findStructDefinitionInMain finds the exact file and line where a struct is defined
// within the main package using AST parsing. This solves the issue where the main
// package doesn't have a predictable directory structure.
func findStructDefinitionInMain(typeName string) (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	var foundFile string

	// Walk through all Go files in the module
	_ = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip vendor, .git, and other hidden directories
		if strings.Contains(path, "/vendor/") || strings.Contains(path, "/.") {
			return nil
		}

		// Parse the file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil // Skip files that can't be parsed
		}

		// Only check files in the main package
		if node.Name.Name != "main" {
			return nil
		}

		// Look for the struct declaration
		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				if x.Name.Name == typeName {
					pos := fset.Position(x.Pos())

					foundFile = filepath.Dir(pos.Filename)
					return false // Stop searching
				}
			}
			return true
		})

		if foundFile != "" {
			return fmt.Errorf("found") // Use error to break out of walk
		}

		return nil
	})

	if foundFile != "" {
		return foundFile, nil
	}

	return "", fmt.Errorf("struct %s not found in main package", typeName)
}

// readModuleName reads the module name from a go.mod file.
func readModuleName(goModPath string) (string, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("module declaration not found in go.mod")
}
