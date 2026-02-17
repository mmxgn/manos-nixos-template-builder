package nix

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mmxgn/manos-nix-template-builder/internal/models"
)

//go:embed templates/flake_template.nix.tmpl
var flakeTemplate string

// FlakeTemplateData holds the data for generating a flake.nix
type FlakeTemplateData struct {
	Description    string
	NixpkgsURL     string
	PythonVersion  string          // e.g. "python3", "python311"
	PythonPackages []string        // package attrs for withPackages (no prefix: numpy, pandas, …)
	SystemPackages []string        // pkgs.* items: zlib, git, cudaPackages.cudatoolkit, …
	PyPIPackages   []PyPIPackageInfo
	EnvVars        []string
	UseFHS         bool
}

// GenerateFlake generates a flake.nix file based on user configuration
func GenerateFlake(config models.UserConfig) (string, error) {
	lang, ok := GetLanguage(config.Language)
	if !ok {
		return "", fmt.Errorf("unknown language: %s", config.Language)
	}

	// Python packages go into python.withPackages — use raw attr names (no prefix)
	pythonPackages := make([]string, len(config.Packages))
	copy(pythonPackages, config.Packages)

	// System packages: zlib (always, for C-extension compatibility) + tools + CUDA/FHS features
	systemPackages := []string{"zlib"}
	systemPackages = append(systemPackages, config.Tools...)
	systemPackages = append(systemPackages, config.EnabledFeatures...)

	// Resolve PyPI packages: fetch version + SHA-256 from PyPI JSON API.
	// Falls back to pkgs.lib.fakeHash on network failure.
	pypiPackages := make([]PyPIPackageInfo, len(config.PyPIPackages))
	for i, name := range config.PyPIPackages {
		pypiPackages[i] = ResolvePyPIPackage(name)
	}

	tmpl, err := template.New("flake").Funcs(template.FuncMap{
		"join": strings.Join,
	}).Parse(flakeTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	nixpkgsURL := config.NixpkgsURL
	if nixpkgsURL == "" {
		nixpkgsURL = "github:NixOS/nixpkgs/nixos-unstable"
	}

	data := FlakeTemplateData{
		Description:    fmt.Sprintf("%s development environment", lang.Name),
		NixpkgsURL:     nixpkgsURL,
		PythonVersion:  config.LanguageVersion,
		PythonPackages: pythonPackages,
		SystemPackages: systemPackages,
		PyPIPackages:   pypiPackages,
		EnvVars:        config.EnvVars,
		UseFHS:         config.UseFHS,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// WriteFlake writes the generated flake content to a file
func WriteFlake(content string, outputPath string) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write flake.nix: %w", err)
	}
	return nil
}
