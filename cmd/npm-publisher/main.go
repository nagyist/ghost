// Generates and publishes npm packages for the ghost CLI binary.
//
// Creates 6 packages:
//
//	@ghost.build/cli              - main package with JS wrapper + optionalDependencies
//	@ghost.build/cli-linux-x64    - linux x64 binary
//	@ghost.build/cli-linux-arm64  - linux arm64 binary
//	@ghost.build/cli-darwin-x64   - darwin x64 binary
//	@ghost.build/cli-darwin-arm64 - darwin arm64 binary
//	@ghost.build/cli-win32-x64   - windows x64 binary
//
// Usage:
//
//	go run ./cmd/npm-publisher --version 0.6.1 --dist-dir ./dist [--dry-run] [--provenance]
//
// The dist-dir should be the GoReleaser dist directory containing the built binaries.
// Each platform's binary should be at: <dist-dir>/ghost_<os>_<goarch>_<variant>/ghost[.exe]
package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed wrapper.cjs
var wrapperScript []byte

const scope = "@ghost.build"

type platformTarget struct {
	NpmOS      string // npm "os" field (e.g. "linux", "darwin", "win32")
	NpmCPU     string // npm "cpu" field (e.g. "x64", "arm64")
	DistDir    string // GoReleaser dist directory name (e.g. "ghost_linux_amd64_v1")
	BinaryName string // Binary filename (e.g. "ghost" or "ghost.exe")
}

func (p platformTarget) PackageSuffix() string {
	return fmt.Sprintf("cli-%s-%s", p.NpmOS, p.NpmCPU)
}

func (p platformTarget) PackageName() string {
	return fmt.Sprintf("%s/%s", scope, p.PackageSuffix())
}

var platforms = []platformTarget{
	{NpmOS: "linux", NpmCPU: "x64", DistDir: "ghost_linux_amd64_v1", BinaryName: "ghost"},
	{NpmOS: "linux", NpmCPU: "arm64", DistDir: "ghost_linux_arm64_v8.0", BinaryName: "ghost"},
	{NpmOS: "darwin", NpmCPU: "x64", DistDir: "ghost_darwin_amd64_v1", BinaryName: "ghost"},
	{NpmOS: "darwin", NpmCPU: "arm64", DistDir: "ghost_darwin_arm64_v8.0", BinaryName: "ghost"},
	{NpmOS: "win32", NpmCPU: "x64", DistDir: "ghost_windows_amd64_v1", BinaryName: "ghost.exe"},
}

func main() {
	version := flag.String("version", "", "Version to publish (required)")
	assetsDir := flag.String("assets-dir", "", "Alias for --dist-dir (deprecated)")
	distDir := flag.String("dist-dir", "", "GoReleaser dist directory containing built binaries (required)")
	dryRun := flag.Bool("dry-run", false, "Perform a dry run without actually publishing")
	provenance := flag.Bool("provenance", false, "Publish with npm provenance (requires OIDC)")
	flag.Parse()

	if *version == "" {
		log.Fatal("--version is required")
	}
	if *distDir == "" {
		*distDir = *assetsDir // support deprecated --assets-dir
	}
	if *distDir == "" {
		log.Fatal("--dist-dir is required")
	}

	resolvedDistDir, err := filepath.Abs(*distDir)
	if err != nil {
		log.Fatalf("resolving dist dir: %v", err)
	}

	workDir, err := os.MkdirTemp("", "ghost-npm-publish-*")
	if err != nil {
		log.Fatalf("creating temp dir: %v", err)
	}
	defer os.RemoveAll(workDir)

	dryRunLabel := ""
	if *dryRun {
		dryRunLabel = " (dry-run)"
	}
	fmt.Printf("Publishing ghost %s to npm%s\n", *version, dryRunLabel)
	fmt.Printf("  Dist: %s\n", resolvedDistDir)
	fmt.Printf("  Work dir: %s\n\n", workDir)

	// Publish platform packages first
	fmt.Println("Publishing platform packages...")
	for _, platform := range platforms {
		if err := publishPlatformPackage(platform, *version, resolvedDistDir, workDir, *dryRun, *provenance); err != nil {
			log.Fatalf("%s: %v", platform.PackageName(), err)
		}
	}
	fmt.Println()

	// Publish main package
	fmt.Println("Publishing main package...")
	if err := publishMainPackage(*version, workDir, *dryRun, *provenance); err != nil {
		log.Fatalf("%s/cli: %v", scope, err)
	}

	fmt.Printf("\nDone! Published %d packages%s.\n", len(platforms)+1, dryRunLabel)
}

func publishPlatformPackage(platform platformTarget, version, distDir, workDir string, dryRun, provenance bool) error {
	packageDir := filepath.Join(workDir, platform.PackageSuffix())
	binDir := filepath.Join(packageDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("creating bin dir: %w", err)
	}

	// Write package.json
	packageJSON := map[string]any{
		"name":        platform.PackageName(),
		"version":     version,
		"description": fmt.Sprintf("ghost CLI binary for %s-%s", platform.NpmOS, platform.NpmCPU),
		"license":     "Apache-2.0",
		"repository": map[string]string{
			"type": "git",
			"url":  "git+https://github.com/timescale/ghost.git",
		},
		"os":    []string{platform.NpmOS},
		"cpu":   []string{platform.NpmCPU},
		"files": []string{"bin"},
	}
	if err := writeJSON(filepath.Join(packageDir, "package.json"), packageJSON); err != nil {
		return fmt.Errorf("writing package.json: %w", err)
	}

	// Copy binary from GoReleaser dist directory
	sourceBinaryPath := filepath.Join(distDir, platform.DistDir, platform.BinaryName)
	destinationBinaryPath := filepath.Join(binDir, platform.BinaryName)
	if err := copyFile(sourceBinaryPath, destinationBinaryPath, 0o755); err != nil {
		return fmt.Errorf("copying binary: %w", err)
	}

	fmt.Printf("  %s@%s\n", platform.PackageName(), version)
	return npmPublish(packageDir, dryRun, provenance)
}

func publishMainPackage(version, workDir string, dryRun, provenance bool) error {
	packageDir := filepath.Join(workDir, "cli")
	binDir := filepath.Join(packageDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("creating bin dir: %w", err)
	}

	// Build optionalDependencies
	optionalDependencies := make(map[string]string, len(platforms))
	for _, platform := range platforms {
		optionalDependencies[platform.PackageName()] = version
	}

	// Write package.json
	packageJSON := map[string]any{
		"name":        scope + "/cli",
		"version":     version,
		"description": "Ghost CLI for managing PostgreSQL databases",
		"license":     "Apache-2.0",
		"repository": map[string]string{
			"type": "git",
			"url":  "git+https://github.com/timescale/ghost.git",
		},
		"homepage":             "https://ghost.build",
		"bin":                  map[string]string{"ghost": "bin/ghost"},
		"files":                []string{"bin"},
		"optionalDependencies": optionalDependencies,
	}
	if err := writeJSON(filepath.Join(packageDir, "package.json"), packageJSON); err != nil {
		return fmt.Errorf("writing package.json: %w", err)
	}

	// Write embedded wrapper script
	wrapperDestinationPath := filepath.Join(binDir, "ghost")
	if err := os.WriteFile(wrapperDestinationPath, wrapperScript, 0o755); err != nil {
		return fmt.Errorf("writing wrapper: %w", err)
	}

	mainPackageName := scope + "/cli"
	fmt.Printf("  %s@%s\n", mainPackageName, version)
	return npmPublish(packageDir, dryRun, provenance)
}

func npmPublish(packageDir string, dryRun, provenance bool) error {
	args := []string{"publish", "--access", "public"}
	if provenance {
		args = append(args, "--provenance")
	}
	if dryRun {
		args = append(args, "--dry-run")
	}

	prefix := ""
	if dryRun {
		prefix = "[dry-run] "
	}
	fmt.Printf("  %snpm %s (in %s)\n", prefix, strings.Join(args, " "), packageDir)

	cmd := exec.Command("npm", args...)
	cmd.Dir = packageDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// writeJSON marshals the value as indented JSON and writes it to the given path.
func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

// copyFile copies a file from source to destination with the given permissions.
func copyFile(sourcePath, destinationPath string, permissions os.FileMode) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("binary not found: %s", sourcePath)
	}
	defer sourceFile.Close()

	destinationFile, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, permissions)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	if _, err = io.Copy(destinationFile, sourceFile); err != nil {
		return err
	}

	if err = sourceFile.Close(); err != nil {
		return fmt.Errorf("closing source file: %w", err)
	}
	if err = destinationFile.Close(); err != nil {
		return fmt.Errorf("closing destination file: %w", err)
	}
	return nil
}
