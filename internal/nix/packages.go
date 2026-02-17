package nix

import (
	"fmt"
	"strings"

	"github.com/mmxgn/nix-template-chooser/internal/models"
)

// NixpkgsChannels lists the available nixpkgs channels with their Python version support.
var NixpkgsChannels = []models.NixpkgsChannel{
	{
		Name:      "nixos-unstable (recommended)",
		FlakeURL:  "github:NixOS/nixpkgs/nixos-unstable",
		IsDefault: true,
		SupportedPythonVersions: []string{
			"python3", "python39", "python310", "python311", "python312",
		},
	},
	{
		Name:     "nixos-24.11 (stable)",
		FlakeURL: "github:NixOS/nixpkgs/nixos-24.11",
		SupportedPythonVersions: []string{
			"python3", "python39", "python310", "python311", "python312",
		},
	},
	{
		Name:     "nixos-24.05",
		FlakeURL: "github:NixOS/nixpkgs/nixos-24.05",
		SupportedPythonVersions: []string{
			"python3", "python310", "python311", "python312",
		},
	},
	{
		Name:     "nixos-23.11",
		FlakeURL: "github:NixOS/nixpkgs/nixos-23.11",
		SupportedPythonVersions: []string{
			"python3", "python310", "python311",
		},
	},
	{
		Name:     "nixos-23.05",
		FlakeURL: "github:NixOS/nixpkgs/nixos-23.05",
		SupportedPythonVersions: []string{
			"python3", "python310", "python311",
		},
	},
}

// LanguageDefinitions maps language names to their configurations
// Currently focused on Python only
var LanguageDefinitions = map[string]models.Language{
	"python": {
		Name: "Python",
		AvailableVersions: []models.LanguageVersion{
			{Name: "Python (latest)", NixAttr: "python3", IsDefault: true},
			{Name: "Python 3.9", NixAttr: "python39", IsDefault: false},
			{Name: "Python 3.10", NixAttr: "python310", IsDefault: false},
			{Name: "Python 3.11", NixAttr: "python311", IsDefault: false},
			{Name: "Python 3.12", NixAttr: "python312", IsDefault: false},
		},
		AvailableTemplates: []models.LanguageTemplate{
			{
				Name:        "Data Science",
				Description: "NumPy, Pandas, Jupyter, Matplotlib for data analysis",
				Version:     "python311",
				Packages:    []string{"numpy", "pandas", "matplotlib", "ipython", "jupyter"},
				Tools:       []string{"git"},
				Features:    []string{},
			},
			{
				Name:        "Web Development",
				Description: "Flask, Requests, pytest for web apps",
				Version:     "python311",
				Packages:    []string{"flask", "requests", "pytest", "black"},
				Tools:       []string{"git", "curl", "jq"},
				Features:    []string{},
			},
			{
				Name:        "Machine Learning (CUDA)",
				Description: "PyTorch, NumPy, Pandas with GPU acceleration",
				Version:     "python311",
				Packages:    []string{"torch", "numpy", "pandas", "matplotlib", "transformers", "diffusers", "sentencepiece", "triton"},
				Tools:       []string{"git"},
				Features:    []string{"CUDA Support"},
			},
			{
				Name:        "Minimal Python",
				Description: "Just Python with basic tools",
				Version:     "python311",
				Packages:    []string{"pytest", "black"},
				Tools:       []string{"git"},
				Features:    []string{},
			},
			{
				Name:        "Custom",
				Description: "Choose your own version, packages and features",
				Version:     "",
				Packages:    []string{},
				Tools:       []string{},
				Features:    []string{},
			},
		},
		CommonPackages: []models.Package{
			// Package Managers
			{Name: "uv", NixAttr: "uv", Description: "Fast Python package manager", Category: "Package Managers"},
			{Name: "pip", NixAttr: "pip", Description: "Python package installer", Category: "Package Managers"},
			// Code Quality
			{Name: "pytest", NixAttr: "pytest", Description: "Testing framework", Category: "Code Quality"},
			{Name: "black", NixAttr: "black", Description: "Code formatter", Category: "Code Quality"},
			{Name: "ruff", NixAttr: "ruff", Description: "Fast Python linter", Category: "Code Quality"},
			{Name: "mypy", NixAttr: "mypy", Description: "Static type checker", Category: "Code Quality"},
			// Interactive
			{Name: "ipython", NixAttr: "ipython", Description: "Enhanced interactive shell", Category: "Interactive"},
			{Name: "jupyter", NixAttr: "jupyter", Description: "Jupyter notebooks", Category: "Interactive"},
			// Science
			{Name: "numpy", NixAttr: "numpy", Description: "Numerical computing", Category: "Science"},
			{Name: "pandas", NixAttr: "pandas", Description: "Data analysis", Category: "Science"},
			{Name: "matplotlib", NixAttr: "matplotlib", Description: "Plotting library", Category: "Science"},
			// Web
			{Name: "requests", NixAttr: "requests", Description: "HTTP library", Category: "Web"},
			{Name: "flask", NixAttr: "flask", Description: "Web framework", Category: "Web"},
			{Name: "django", NixAttr: "django", Description: "Full-stack web framework", Category: "Web"},
			// Machine Learning
			{Name: "torch", NixAttr: "torch", Description: "PyTorch ML framework", Category: "Machine Learning"},
			{Name: "torchaudio", NixAttr: "torchaudio", Description: "PyTorch audio processing", Category: "Machine Learning"},
			{Name: "tensorflow", NixAttr: "tensorflow", Description: "TensorFlow ML framework", Category: "Machine Learning"},
			{Name: "transformers", NixAttr: "transformers", Description: "Hugging Face Transformers", Category: "Machine Learning"},
			{Name: "diffusers", NixAttr: "diffusers", Description: "Hugging Face Diffusers", Category: "Machine Learning"},
			{Name: "sentencepiece", NixAttr: "sentencepiece", Description: "Text tokenization library", Category: "Machine Learning"},
			{Name: "triton", NixAttr: "triton", Description: "GPU kernel programming (OpenAI)", Category: "Machine Learning"},
			{Name: "accelerate", NixAttr: "accelerate", Description: "Hugging Face Accelerate for distributed training", Category: "Machine Learning"},
			{Name: "xformers", NixAttr: "xformers", Description: "Memory-efficient transformers with custom CUDA kernels", Category: "Machine Learning"},
			{Name: "flash-attn", NixAttr: "flash-attn", Description: "Fast and memory-efficient exact attention", Category: "Machine Learning"},
			// Audio
			{Name: "soundfile", NixAttr: "soundfile", Description: "Read/write audio files", Category: "Audio"},
			{Name: "soxr", NixAttr: "soxr", Description: "High-quality audio resampling", Category: "Audio"},
			// Utilities
			{Name: "platformdirs", NixAttr: "platformdirs", Description: "Platform-specific directories", Category: "Utilities"},
		},
		SpecialFeatures: []models.Feature{
			{
				Name:        "CUDA Support",
				NixAttrs:    []string{"cudaPackages.cudatoolkit", "cudaPackages.cudnn"},
				Description: "Enable NVIDIA CUDA for GPU acceleration",
				Languages:   []string{"python"},
			},
			{
				Name:        "FHS Environment",
				NixAttrs:    []string{},
				Description: "Wrap shell in buildFHSEnv (standard Linux paths — useful for CUDA or foreign binaries)",
				Languages:   []string{"python"},
			},
		},
		BuildSystem: "buildPythonPackage",
	},
}

// CommonTools available across all languages
var CommonTools = []models.Package{
	// Version Control
	{Name: "git", NixAttr: "git", Description: "Version control", Category: "Version Control"},
	{Name: "git-lfs", NixAttr: "git-lfs", Description: "Git extension for large file storage", Category: "Version Control"},
	// Network
	{Name: "curl", NixAttr: "curl", Description: "HTTP client", Category: "Network"},
	// Data Processing
	{Name: "jq", NixAttr: "jq", Description: "JSON processor", Category: "Data Processing"},
	{Name: "ffmpeg", NixAttr: "ffmpeg", Description: "Audio/video conversion and processing", Category: "Data Processing"},
	{Name: "soxr", NixAttr: "soxr", Description: "High-quality audio resampling library", Category: "Data Processing"},
	// Task & Environment
	{Name: "pre-commit", NixAttr: "pre-commit", Description: "Git pre-commit hook manager", Category: "Task & Environment"},
	{Name: "direnv", NixAttr: "direnv", Description: "Environment switcher", Category: "Task & Environment"},
	{Name: "just", NixAttr: "just", Description: "Command runner", Category: "Task & Environment"},
	// Search
	{Name: "ripgrep", NixAttr: "ripgrep", Description: "Fast grep alternative", Category: "Search"},
	{Name: "fd", NixAttr: "fd", Description: "Fast find alternative", Category: "Search"},
	// Editors
	{Name: "neovim", NixAttr: "neovim", Description: "Hyperextensible Vim-based text editor", Category: "Editors"},
	{Name: "emacs", NixAttr: "emacs", Description: "Extensible, self-documenting text editor", Category: "Editors"},
	// Compilers
	{Name: "gcc", NixAttr: "gcc", Description: "GNU Compiler Collection", Category: "Compilers"},
	// Viewers & System
	{Name: "bat", NixAttr: "bat", Description: "Cat with syntax highlighting", Category: "Viewers & System"},
	{Name: "exa", NixAttr: "exa", Description: "Modern ls replacement", Category: "Viewers & System"},
	{Name: "htop", NixAttr: "htop", Description: "Interactive process viewer", Category: "Viewers & System"},
}

// GetLanguageNames returns all available language names
func GetLanguageNames() []string {
	names := make([]string, 0, len(LanguageDefinitions))
	for name := range LanguageDefinitions {
		names = append(names, name)
	}
	return names
}

// GetLanguage returns the language definition for a given name
func GetLanguage(name string) (models.Language, bool) {
	lang, ok := LanguageDefinitions[name]
	return lang, ok
}

// GetVersionedPackageAttr returns the version-specific package attribute
func GetVersionedPackageAttr(version string, pkgAttr string) string {
	// For Python: python311Packages.pytest
	// For others: just the attr
	if strings.HasPrefix(version, "python") {
		// Skip if it's already a full package path (like nodePackages.npm)
		if strings.Contains(pkgAttr, "Packages.") {
			return pkgAttr
		}
		return fmt.Sprintf("%sPackages.%s", version, pkgAttr)
	}
	return pkgAttr
}

// GetDefaultVersion returns the default version for a language
func GetDefaultVersion(lang models.Language) models.LanguageVersion {
	for _, version := range lang.AvailableVersions {
		if version.IsDefault {
			return version
		}
	}
	// Fallback to first version if no default
	if len(lang.AvailableVersions) > 0 {
		return lang.AvailableVersions[0]
	}
	return models.LanguageVersion{}
}
