//go:build ignore
// +build ignore

/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		handleGenerate(os.Args[2:])
	case "backup-referencers":
		handleCommand(backupReferencers, "backed up")
	case "restore-referencers":
		handleCommand(restoreReferencers, "restored")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// printUsage displays the usage instructions for the script.
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  generate <folder1> [folder2] ...  - Generate cluster scope files")
	fmt.Println("  backup-referencers               - Backup referencers files")
	fmt.Println("  restore-referencers              - Restore referencers files")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  generate groups projects")
	fmt.Println("  generate clients")
}

// handleGenerate processes the generate command with the provided folders.
func handleGenerate(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "generate command requires at least one folder\n")
		printUsage()
		os.Exit(1)
	}

	if err := generateClusterScope(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully generated cluster scope for folders: %v\n", args)
}

// handleCommand executes the given function and handles errors uniformly.
func handleCommand(fn func() error, action string) {
	if err := fn(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully %s referencers files\n", action)
}

// generateClusterScope generates cluster-scoped resource files
// from namespaced resource files in the specified folders.
func generateClusterScope(folders []string) error {
	namespacedPath := "namespaced"
	clusterPath := "cluster"

	if err := validateAndSetup(namespacedPath, clusterPath, folders); err != nil {
		return err
	}

	for _, folder := range folders {
		if err := processFolder(namespacedPath, clusterPath, folder); err != nil {
			return fmt.Errorf("failed to process folder %s: %w", folder, err)
		}
	}

	return nil
}

// validateAndSetup checks the existence of the namespaced directory,
// creates the cluster directory if it doesn't exist, and cleans up
// any existing cluster folders that will be regenerated.
func validateAndSetup(namespacedPath, clusterPath string, folders []string) error {
	if _, err := os.Stat(namespacedPath); os.IsNotExist(err) {
		return fmt.Errorf("namespaced directory does not exist at %s", namespacedPath)
	}

	if err := os.MkdirAll(clusterPath, 0755); err != nil {
		return fmt.Errorf("failed to create cluster directory: %w", err)
	}

	for _, folder := range folders {
		targetFolder := filepath.Join(clusterPath, folder)
		if err := os.RemoveAll(targetFolder); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove cluster folder %s: %w", targetFolder, err)
		}
	}

	return nil
}

// processFolder processes all files in the specified folder, applying necessary transformations.
func processFolder(namespacedPath, clusterPath, folder string) error {
	folderPath := filepath.Join(namespacedPath, folder)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		fmt.Printf("Warning: folder %s does not exist, skipping\n", folderPath)
		return nil
	}

	return filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(d.Name(), "zz_") {
			return nil
		}

		targetPath := strings.Replace(path, namespacedPath, clusterPath, 1)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		if strings.HasSuffix(path, ".go") && !shouldSkipFile(path) {
			dir := filepath.Dir(targetPath)
			filename := filepath.Base(targetPath)
			targetPath = filepath.Join(dir, "zz_"+filename)
			return transformFile(path, targetPath)
		}

		return nil
	})
}

// shouldSkipFile checks if the file contains the +cluster-scope:skip-file annotation.
func shouldSkipFile(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), "+cluster-scope:skip-file")
}

// transformFile reads the source file, applies transformations,
// and writes the result to the destination file.
func transformFile(srcPath, dstPath string) error {
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", srcPath, err)
	}

	transformed := transformContent(string(content))

	if formatted, err := format.Source([]byte(transformed)); err == nil {
		transformed = string(formatted)
	}

	return os.WriteFile(dstPath, []byte(transformed), 0644)
}

// processDeleteAnnotations removes lines based on +cluster-scope:delete=N annotations.
func processDeleteAnnotations(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if idx := strings.Index(line, "+cluster-scope:delete="); idx != -1 {
			numStr := strings.TrimSpace(line[idx+len("+cluster-scope:delete="):])
			if num, err := strconv.Atoi(numStr); err == nil && num > 0 {
				i += num
				continue
			}
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// transformContent applies all necessary transformations to convert
// namespaced resource code to cluster-scoped resource code.
func transformContent(content string) string {
	content = processDeleteAnnotations(content)
	content = addGenerationComment(content)

	for _, r := range getReplacements() {
		content = strings.ReplaceAll(content, r.old, r.new)
	}

	return content
}

// addGenerationComment adds a code generation comment at the top of the file.
func addGenerationComment(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			result := make([]string, 0, len(lines)+2)
			result = append(result, lines[:i]...)
			result = append(result, "// Code generated by hack/generate-cluster-scope.go - DO NOT EDIT.", "")
			result = append(result, lines[i:]...)
			return strings.Join(result, "\n")
		}
	}
	return content
}

// replacement defines a string replacement rule.
type replacement struct {
	old, new string
}

// getReplacements returns a list of string replacements to convert
// This is important to keep a list structure to ensure priority of replacements.
func getReplacements() []replacement {
	return []replacement{
		{"m.crossplane.io", "crossplane.io"},
		{"xpv2.ManagedResourceSpec", "xpv1.ResourceSpec"},
		{"xpv1.NamespacedReference", "xpv1.Reference"},
		{"xpv1.NamespacedSelector", "xpv1.Selector"},
		{"xpv1.LocalSecretKeySelector", "xpv1.SecretKeySelector"},
		{"xpv1.LocalSecretReference", "xpv1.SecretReference"},
		{"LocalSecretReference:", "SecretReference:"},
		{".LocalSecretReference", ".SecretReference"},
		{"reference.NewAPINamespacedResolver", "reference.NewAPIResolver"},
		{"reference.NamespacedResolutionRequest", "reference.ResolutionRequest"},
		{"reference.NamespacedResolutionResponse", "reference.ResolutionResponse"},
		{"kubebuilder:resource:scope=Namespaced", "kubebuilder:resource:scope=Cluster"},
		{"/namespaced/", "/cluster/"},
		{"GetTokenValueFromLocalSecret", "GetTokenValueFromSecret"},
		{"TestCreateLocalSecretKeySelector", "TestCreateSecretKeySelector"},
		{"TestCreateLocalSecretReference", "TestCreateSecretReference"},
	}
}

// backupReferencers backs up all referencers.go files by renaming them with a .bak extension.
func backupReferencers() error {
	return filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(path, "referencers.go") && !strings.HasSuffix(path, ".bak") {
			return os.Rename(path, path+".bak")
		}
		return nil
	})
}

// restoreReferencers restores all referencers.go.bak files by renaming them back to referencers.go.
func restoreReferencers() error {
	return filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(path, "referencers.go.bak") {
			return os.Rename(path, strings.TrimSuffix(path, ".bak"))
		}
		return nil
	})
}
