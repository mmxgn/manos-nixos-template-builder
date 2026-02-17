package models

// NixpkgsChannel represents a versioned nixpkgs input
type NixpkgsChannel struct {
	Name                    string   // Display name (e.g., "nixos-unstable (recommended)")
	FlakeURL                string   // Flake input URL
	IsDefault               bool     // Whether this is the default choice
	SupportedPythonVersions []string // Python NixAttrs available in this channel
}

// Package represents a Nix package
type Package struct {
	Name        string
	NixAttr     string // Nix attribute path (e.g., "pytest", "numpy")
	Description string
	Category    string // Display group (e.g., "Machine Learning", "Audio")
}

// LanguageVersion represents a specific version of a language
type LanguageVersion struct {
	Name      string // Display name (e.g., "Python 3.11")
	NixAttr   string // Nix attribute (e.g., "python311")
	IsDefault bool   // Default version for this language
}

// Feature represents an optional feature or capability
type Feature struct {
	Name        string   // Display name (e.g., "CUDA Support")
	NixAttrs    []string // Nix attributes to add (e.g., ["cudatoolkit", "cudnn"])
	Description string   // What this feature provides
	Languages   []string // Which languages support this (empty = all)
}

// LanguageTemplate represents a preset configuration for a language
type LanguageTemplate struct {
	Name        string   // Display name (e.g., "Data Science")
	Description string   // What it includes
	Version     string   // Version NixAttr (e.g., "python311")
	Packages    []string // Package NixAttrs to include
	Tools       []string // Tool NixAttrs to include
	Features    []string // Feature names to enable (e.g., "CUDA Support")
}

// Language represents a programming language configuration
type Language struct {
	Name               string
	AvailableVersions  []LanguageVersion  // Available versions
	AvailableTemplates []LanguageTemplate // Preset templates
	CommonPackages     []Package
	SpecialFeatures    []Feature // Optional features
	BuildSystem        string    // e.g., "buildGoModule", "buildPythonPackage"
}
