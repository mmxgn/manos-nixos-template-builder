package nix

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/mmxgn/manos-nix-template-builder/internal/models"
)

// FlakeShowOutput represents the JSON output of `nix flake show --json`
type FlakeShowOutput struct {
	DefaultTemplate TemplateInfo            `json:"defaultTemplate"`
	Templates       map[string]TemplateInfo `json:"templates"`
}

// TemplateInfo represents template metadata
type TemplateInfo struct {
	Description string `json:"description"`
	Type        string `json:"type"`
}

// FetchNixOSTemplates fetches available templates from github:NixOS/templates
func FetchNixOSTemplates() ([]models.NixTemplate, error) {
	cmd := exec.Command("nix", "flake", "show", "--json", "github:NixOS/templates")
	// Only capture stdout, let stderr go to /dev/null
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch templates: %w", err)
	}

	var flakeOutput FlakeShowOutput
	if err := json.Unmarshal(output, &flakeOutput); err != nil {
		return nil, fmt.Errorf("failed to parse template JSON: %w", err)
	}

	templates := make([]models.NixTemplate, 0, len(flakeOutput.Templates))
	for name, info := range flakeOutput.Templates {
		templates = append(templates, models.NixTemplate{
			Name:        name,
			Description: info.Description,
			WelcomeText: "",
			Path:        fmt.Sprintf("github:NixOS/templates#%s", name),
			Source:      "nixos",
		})
	}

	return templates, nil
}

// InitializeTemplate initializes a project with the given template
func InitializeTemplate(templateName string, targetDir string) error {
	templateRef := fmt.Sprintf("github:NixOS/templates#%s", templateName)
	cmd := exec.Command("nix", "flake", "init", "-t", templateRef)
	if targetDir != "" && targetDir != "." {
		// If a specific directory is provided, we would cd there first
		// For now, we'll assume we're already in the target directory
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize template: %w (output: %s)", err, string(output))
	}

	return nil
}
