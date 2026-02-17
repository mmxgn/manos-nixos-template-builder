package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmxgn/manos-nix-template-builder/internal/models"
	"github.com/mmxgn/manos-nix-template-builder/internal/nix"
)

// Screen represents the current screen in the TUI
type Screen int

const (
	ScreenModeSelection    Screen = iota
	ScreenTemplateBrowser         // NixOS templates (Quick Mode)
	ScreenLanguageSelector        // Language picker (reserved for future multi-language)
	ScreenTemplateOrCustom        // Language-specific preset templates or custom
	ScreenNixpkgsSelector         // Nixpkgs channel selection (before Python version)
	ScreenVersionSelector         // Python version selection (Custom mode only)
	ScreenPackageSelector         // Package multi-select (Custom mode only)
	ScreenToolSelector            // Tool multi-select (Custom mode only)
	ScreenFeatureSelector         // Feature multi-select (Custom mode only)
	ScreenConfirmation
	ScreenCompletion
)

// Model is the main Bubble Tea model
type Model struct {
	// State
	CurrentScreen Screen
	Config        models.UserConfig
	Err           error
	Quitting      bool

	// Data
	NixTemplates     []models.NixTemplate        // For quick mode (NixOS templates)
	Languages        []string
	LangTemplates    []models.LanguageTemplate   // For custom mode (language-specific templates)

	// Multi-select data
	Versions []models.LanguageVersion
	Packages []models.Package
	Tools    []models.Package
	Features []models.Feature

	// UI Components
	ModeList             list.Model
	TemplateList         list.Model  // For quick mode (NixOS templates)
	LanguageList         list.Model
	TemplateOrCustomList list.Model  // Language-specific preset templates
	NixpkgsChannelList   list.Model  // Nixpkgs channel selection
	VersionList          list.Model  // Kept for potential future use
	PackageList          list.Model
	ToolList             list.Model
	FeatureList          list.Model
	TextInput            textinput.Model

	// Selection state
	Cursor              int                    // For multi-select navigation
	SelectedNixpkgs     models.NixpkgsChannel  // Currently chosen nixpkgs channel
	SelectedPackages    map[string]bool
	SelectedTools       map[string]bool
	SelectedFeatures    map[string]bool
	CustomPackages      []string        // User-entered custom nixpkgs packages
	PyPIPackages        []string        // User-entered PyPI packages
	SelectedPyPI        map[string]bool // Which PyPI packages are selected
	PyPICursor          int             // Cursor within PyPI list in overlay
	EnvVars             []string        // User-entered environment variable names
	SelectedEnvVars     map[string]bool // Which env vars are selected
	EnvVarCursor        int             // Cursor within env var list in overlay

	// Input modes
	AddingCustomPackage bool // True when adding custom package inline
	AddingPyPIPackage   bool // True when adding PyPI package inline
	AddingCustomTool    bool // True when adding custom tool inline
	AddingEnvVar        bool // True when adding env var inline
	AskingGitAdd        bool // True when flake.nix is untracked and we ask whether to git-add
	AskingOverwrite     bool // True when output file already exists and we ask whether to overwrite

	// Dimensions
	Width  int
	Height int

	// Loading state
	Loading bool
}

// InitialModel creates a new Model with default values
func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "./flake.nix"
	ti.CharLimit = 256
	ti.Width = 50

	// Setup mode selection list
	items := []list.Item{
		ListItem{
			ItemTitle: "Python",
			ItemDesc:  "Build a custom Python flake with packages and tools",
		},
		ListItem{
			ItemTitle: "NixOS Templates",
			ItemDesc:  "Browse and use existing NixOS templates",
		},
	}

	delegate := list.NewDefaultDelegate()
	modeList := list.New(items, delegate, 0, 0)
	modeList.Title = "Nix Flake Template Chooser"
	modeList.SetShowStatusBar(false)
	modeList.SetFilteringEnabled(false)

	// Find default nixpkgs channel
	defaultChannel := nix.NixpkgsChannels[0]
	for _, ch := range nix.NixpkgsChannels {
		if ch.IsDefault {
			defaultChannel = ch
			break
		}
	}

	return Model{
		CurrentScreen:    ScreenModeSelection,
		Config:           models.UserConfig{OutputPath: "./flake.nix", NixpkgsURL: defaultChannel.FlakeURL},
		SelectedNixpkgs:  defaultChannel,
		SelectedPackages: make(map[string]bool),
		SelectedTools:    make(map[string]bool),
		SelectedFeatures: make(map[string]bool),
		SelectedPyPI:     make(map[string]bool),
		SelectedEnvVars:  make(map[string]bool),
		TextInput:        ti,
		ModeList:         modeList,
		Cursor:           0,
		Loading:          false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// ListItem is a simple item for lists
type ListItem struct {
	ItemTitle string
	ItemDesc  string
}

func (i ListItem) Title() string       { return i.ItemTitle }
func (i ListItem) Description() string { return i.ItemDesc }
func (i ListItem) FilterValue() string { return i.ItemTitle }
