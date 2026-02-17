package models

// NixTemplate represents a Nix flake template
type NixTemplate struct {
	Name        string
	Description string
	WelcomeText string
	Path        string
	Source      string // "nixos" or "custom"
}

// TemplateList holds a collection of templates
type TemplateList struct {
	Templates []NixTemplate
}
