package nix

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mmxgn/nix-template-chooser/internal/models"
)

//go:embed templates/flake_template.nix.tmpl
var flakeTemplate string

// FlakeTemplateData holds the data for generating a flake.nix
type FlakeTemplateData struct {
	Description  string
	NixpkgsURL   string
	Packages     []string
	PyPIPackages []string
	HasPyPI      bool
	EnvVars      []string
	UseFHS       bool
}

// GenerateFlake generates a flake.nix file based on user configuration
func GenerateFlake(config models.UserConfig) (string, error) {
	lang, ok := GetLanguage(config.Language)
	if !ok {
		return "", fmt.Errorf("unknown language: %s", config.Language)
	}

	// Get version-specific package attributes
	versionedPackages := make([]string, 0, len(config.Packages))
	for _, pkgAttr := range config.Packages {
		versionedPkg := GetVersionedPackageAttr(config.LanguageVersion, pkgAttr)
		versionedPackages = append(versionedPackages, versionedPkg)
	}

	// Combine: version, packages, tools, feature attrs
	allPackages := []string{config.LanguageVersion}
	allPackages = append(allPackages, versionedPackages...)
	allPackages = append(allPackages, config.Tools...)
	allPackages = append(allPackages, config.EnabledFeatures...)

	// Add pip if there are PyPI packages
	if len(config.PyPIPackages) > 0 {
		allPackages = append(allPackages, "pip")
	}

	// Create template with custom functions (using embedded template)
	tmpl, err := template.New("flake").Funcs(template.FuncMap{
		"join": strings.Join,
	}).Parse(flakeTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Determine nixpkgs URL (fall back to unstable if not set)
	nixpkgsURL := config.NixpkgsURL
	if nixpkgsURL == "" {
		nixpkgsURL = "github:NixOS/nixpkgs/nixos-unstable"
	}

	// Prepare template data
	data := FlakeTemplateData{
		Description:  fmt.Sprintf("%s development environment", lang.Name),
		NixpkgsURL:   nixpkgsURL,
		Packages:     allPackages,
		PyPIPackages: config.PyPIPackages,
		HasPyPI:      len(config.PyPIPackages) > 0,
		EnvVars:      config.EnvVars,
		UseFHS:       config.UseFHS,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// WriteFlake writes the generated flake content to a file
func WriteFlake(content string, outputPath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write flake.nix: %w", err)
	}

	return nil
}

