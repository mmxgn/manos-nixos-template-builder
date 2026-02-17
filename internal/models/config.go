package models

// UserConfig holds the user's configuration choices
type UserConfig struct {
	Mode             string   // "quick" or "custom"
	SelectedTemplate string   // For quick mode (NixOS templates)
	TemplateName     string   // For custom mode with preset template
	Language         string   // For custom mode
	LanguageVersion  string   // Selected version NixAttr (e.g., "python311")
	Packages         []string // Language-specific packages (NixAttrs)
	PyPIPackages     []string // PyPI packages to install with pip
	Tools            []string // Dev tools (git, jq, etc.)
	EnabledFeatures  []string // Selected feature NixAttrs
	EnvVars          []string // Extra environment variable names (set to empty string)
	UseFHS           bool     // Wrap devShell in buildFHSEnv (useful for CUDA)
	NixpkgsURL       string   // Nixpkgs flake input URL (e.g., "github:NixOS/nixpkgs/nixos-unstable")
	OutputPath       string   // Where to write flake.nix
}
