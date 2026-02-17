package tui

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmxgn/manos-nix-template-builder/internal/models"
	"github.com/mmxgn/manos-nix-template-builder/internal/nix"
)

// Update handles all state updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Allow quitting from any screen except during loading
			if !m.Loading && m.CurrentScreen != ScreenCompletion {
				m.Quitting = true
				return m, tea.Quit
			}
			// From completion screen, just quit
			if m.CurrentScreen == ScreenCompletion {
				m.Quitting = true
				return m, tea.Quit
			}

		case "esc":
			// If in input mode, cancel it instead of going back
			if m.AddingCustomPackage {
				m.AddingCustomPackage = false
				m.TextInput.SetValue("")
				m.TextInput.Blur()
				return m, nil
			}
			if m.AddingPyPIPackage {
				m.AddingPyPIPackage = false
				m.TextInput.SetValue("")
				m.TextInput.Blur()
				return m, nil
			}
			if m.AddingCustomTool {
				m.AddingCustomTool = false
				m.TextInput.SetValue("")
				m.TextInput.Blur()
				return m, nil
			}
			if m.AddingEnvVar {
				m.AddingEnvVar = false
				m.TextInput.SetValue("")
				m.TextInput.Blur()
				return m, nil
			}
			// Go back to previous screen (except from first screen)
			if m.CurrentScreen > ScreenModeSelection {
				return m.goBack(), nil
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Update list sizes (only if they have been initialized with items)
		// Check by seeing if the list has items, not just if width is 0
		if len(m.ModeList.Items()) > 0 && m.ModeList.Width() == 0 {
			m.ModeList.SetWidth(msg.Width - 4)
			m.ModeList.SetHeight(msg.Height - 10)
		}
		if len(m.TemplateList.Items()) > 0 && m.TemplateList.Width() == 0 {
			m.TemplateList.SetWidth(msg.Width - 4)
			m.TemplateList.SetHeight(msg.Height - 10)
		}
		if len(m.LanguageList.Items()) > 0 && m.LanguageList.Width() == 0 {
			m.LanguageList.SetWidth(msg.Width - 4)
			m.LanguageList.SetHeight(msg.Height - 10)
		}
		if len(m.TemplateOrCustomList.Items()) > 0 && m.TemplateOrCustomList.Width() == 0 {
			m.TemplateOrCustomList.SetWidth(msg.Width - 4)
			m.TemplateOrCustomList.SetHeight(msg.Height - 10)
		}
		if len(m.NixpkgsChannelList.Items()) > 0 && m.NixpkgsChannelList.Width() == 0 {
			m.NixpkgsChannelList.SetWidth(msg.Width - 4)
			m.NixpkgsChannelList.SetHeight(msg.Height - 10)
		}
		if len(m.VersionList.Items()) > 0 && m.VersionList.Width() == 0 {
			m.VersionList.SetWidth(msg.Width - 4)
			m.VersionList.SetHeight(msg.Height - 10)
		}
		if len(m.PackageList.Items()) > 0 && m.PackageList.Width() == 0 {
			m.PackageList.SetWidth(msg.Width - 4)
			m.PackageList.SetHeight(msg.Height - 12)
		}
		if len(m.ToolList.Items()) > 0 && m.ToolList.Width() == 0 {
			m.ToolList.SetWidth(msg.Width - 4)
			m.ToolList.SetHeight(msg.Height - 12)
		}
	}

	// Route to screen-specific handlers
	switch m.CurrentScreen {
	case ScreenModeSelection:
		return m.updateModeSelection(msg)
	case ScreenTemplateBrowser:
		return m.updateTemplateBrowser(msg)
	case ScreenLanguageSelector:
		return m.updateLanguageSelector(msg)
	case ScreenTemplateOrCustom:
		return m.updateTemplateOrCustom(msg)
	case ScreenNixpkgsSelector:
		return m.updateNixpkgsSelector(msg)
	case ScreenVersionSelector:
		return m.updateVersionSelector(msg)
	case ScreenPackageSelector:
		return m.updatePackageSelector(msg)
	case ScreenToolSelector:
		return m.updateToolSelector(msg)
	case ScreenFeatureSelector:
		return m.updateFeatureSelector(msg)
	case ScreenConfirmation:
		return m.updateConfirmation(msg)
	case ScreenCompletion:
		return m.updateCompletion(msg)
	}

	return m, nil
}

// goBack navigates to the previous screen
func (m Model) goBack() Model {
	switch m.CurrentScreen {
	case ScreenTemplateBrowser:
		m.CurrentScreen = ScreenModeSelection
	case ScreenLanguageSelector:
		m.CurrentScreen = ScreenModeSelection
	case ScreenTemplateOrCustom:
		m.CurrentScreen = ScreenModeSelection
	case ScreenNixpkgsSelector:
		m.CurrentScreen = ScreenTemplateOrCustom
	case ScreenVersionSelector:
		m.CurrentScreen = ScreenNixpkgsSelector
	case ScreenPackageSelector:
		m.CurrentScreen = ScreenVersionSelector
	case ScreenToolSelector:
		m.CurrentScreen = ScreenPackageSelector
	case ScreenFeatureSelector:
		m.CurrentScreen = ScreenToolSelector
	case ScreenConfirmation:
		if m.Config.Mode == "quick" {
			m.CurrentScreen = ScreenTemplateBrowser
		} else {
			if m.Config.TemplateName != "" {
				m.CurrentScreen = ScreenTemplateOrCustom
			} else if len(m.Features) > 0 {
				m.CurrentScreen = ScreenFeatureSelector
			} else {
				m.CurrentScreen = ScreenToolSelector
			}
		}
	case ScreenCompletion:
		// Can't go back from completion
	}
	return m
}

// Screen-specific update functions
func (m Model) updateModeSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected := m.ModeList.SelectedItem()
			if item, ok := selected.(ListItem); ok {
				if item.ItemTitle == "NixOS Templates" {
					m.Config.Mode = "quick"
					// Fetch templates and transition to template browser
					templates, err := nix.FetchNixOSTemplates()
					if err != nil {
						m.Err = err
						return m, nil
					}
					m.NixTemplates = templates

					// Setup template list
					items := make([]list.Item, len(templates))
					for i, tmpl := range templates {
						items[i] = ListItem{
							ItemTitle: tmpl.Name,
							ItemDesc:  tmpl.Description,
						}
					}
					delegate := list.NewDefaultDelegate()
					m.TemplateList = list.New(items, delegate, m.Width-4, m.Height-10)
					m.TemplateList.Title = "Select a NixOS Template"
					m.TemplateList.SetShowStatusBar(false)
					m.TemplateList.SetFilteringEnabled(false)

					m.CurrentScreen = ScreenTemplateBrowser
				} else {
					m.Config.Mode = "custom"
					// Skip language selection - only Python is supported
					m.Config.Language = "python"
					langDef, _ := nix.GetLanguage("python")
					m.Versions = langDef.AvailableVersions
					m.Packages = langDef.CommonPackages
					m.Features = langDef.SpecialFeatures
					m.LangTemplates = langDef.AvailableTemplates

					// Setup template/custom selection list
					items := make([]list.Item, len(m.LangTemplates))
					for i, tmpl := range m.LangTemplates {
						items[i] = ListItem{
							ItemTitle: tmpl.Name,
							ItemDesc:  tmpl.Description,
						}
					}
					delegate := list.NewDefaultDelegate()
					m.TemplateOrCustomList = list.New(items, delegate, m.Width-4, m.Height-10)
					m.TemplateOrCustomList.Title = "Python Configuration"
					m.TemplateOrCustomList.SetShowStatusBar(false)
					m.TemplateOrCustomList.SetFilteringEnabled(false)

					// Go directly to template/custom selection
					m.CurrentScreen = ScreenTemplateOrCustom
				}
			}
		}
	}

	var cmd tea.Cmd
	m.ModeList, cmd = m.ModeList.Update(msg)
	return m, cmd
}

func (m Model) updateTemplateBrowser(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected := m.TemplateList.SelectedItem()
			if item, ok := selected.(ListItem); ok {
				m.Config.SelectedTemplate = item.ItemTitle
				m.CurrentScreen = ScreenConfirmation
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.TemplateList, cmd = m.TemplateList.Update(msg)
	return m, cmd
}

func (m Model) updateTemplateOrCustom(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected := m.TemplateOrCustomList.SelectedItem()
			if item, ok := selected.(ListItem); ok {
				// Find the selected template
				for _, tmpl := range m.LangTemplates {
					if tmpl.Name == item.ItemTitle {
						if tmpl.Name == "Custom" {
							// Go to custom mode - nixpkgs channel selection first
							items := make([]list.Item, len(nix.NixpkgsChannels))
							for i, ch := range nix.NixpkgsChannels {
								desc := ch.FlakeURL
								if ch.IsDefault {
									desc += " (default)"
								}
								items[i] = ListItem{
									ItemTitle: ch.Name,
									ItemDesc:  desc,
								}
							}
							delegate := list.NewDefaultDelegate()
							m.NixpkgsChannelList = list.New(items, delegate, m.Width-4, m.Height-10)
							m.NixpkgsChannelList.Title = "Select nixpkgs Channel"
							m.NixpkgsChannelList.SetShowStatusBar(false)
							m.NixpkgsChannelList.SetFilteringEnabled(false)

							// Clear template name for custom mode
							m.Config.TemplateName = ""
							m.CurrentScreen = ScreenNixpkgsSelector
							return m, nil
						} else {
							// Use preset template
							m.Config.TemplateName = tmpl.Name
							m.Config.LanguageVersion = tmpl.Version
							m.Config.Packages = tmpl.Packages
							m.Config.Tools = tmpl.Tools

							// Enable features by converting feature names to NixAttrs
							m.Config.EnabledFeatures = make([]string, 0)
							for _, featureName := range tmpl.Features {
								// Find feature and add its NixAttrs
								for _, feature := range m.Features {
									if feature.Name == featureName {
										m.Config.EnabledFeatures = append(m.Config.EnabledFeatures, feature.NixAttrs...)
										break
									}
								}
							}

							// Go directly to confirmation
							m.CurrentScreen = ScreenConfirmation
							return m, nil
						}
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.TemplateOrCustomList, cmd = m.TemplateOrCustomList.Update(msg)
	return m, cmd
}

// updateNixpkgsSelector handles nixpkgs channel selection
func (m Model) updateNixpkgsSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected := m.NixpkgsChannelList.SelectedItem()
			if item, ok := selected.(ListItem); ok {
				for _, ch := range nix.NixpkgsChannels {
					if ch.Name == item.ItemTitle {
						m.SelectedNixpkgs = ch
						m.Config.NixpkgsURL = ch.FlakeURL
						// Reset cursor for version selection
						m.Cursor = 0
						m.CurrentScreen = ScreenVersionSelector
						return m, nil
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.NixpkgsChannelList, cmd = m.NixpkgsChannelList.Update(msg)
	return m, cmd
}

func (m Model) updateLanguageSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected := m.LanguageList.SelectedItem()
			if item, ok := selected.(ListItem); ok {
				// Find the language key and get its data
				for key, langDef := range nix.LanguageDefinitions {
					if langDef.Name == item.ItemTitle {
						m.Config.Language = key
						m.Versions = langDef.AvailableVersions
						m.Packages = langDef.CommonPackages
						m.Features = langDef.SpecialFeatures
						m.LangTemplates = langDef.AvailableTemplates

						// Setup template/custom selection list
						items := make([]list.Item, len(m.LangTemplates))
						for i, tmpl := range m.LangTemplates {
							items[i] = ListItem{
								ItemTitle: tmpl.Name,
								ItemDesc:  tmpl.Description,
							}
						}
						delegate := list.NewDefaultDelegate()
						m.TemplateOrCustomList = list.New(items, delegate, m.Width-4, m.Height-10)
						m.TemplateOrCustomList.Title = "📋 Choose Configuration"
						m.TemplateOrCustomList.SetShowStatusBar(false)
						m.TemplateOrCustomList.SetFilteringEnabled(false)

						// Transition to template/custom selection
						m.CurrentScreen = ScreenTemplateOrCustom
						return m, nil
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.LanguageList, cmd = m.LanguageList.Update(msg)
	return m, cmd
}

// updateVersionSelector handles version selection (cursor-based, respects nixpkgs channel support)
func (m Model) updateVersionSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Versions)-1 {
				m.Cursor++
			}
		case "enter":
			if m.Cursor >= 0 && m.Cursor < len(m.Versions) {
				version := m.Versions[m.Cursor]
				// Do not allow selecting unsupported versions
				if !m.isVersionSupported(version.NixAttr) {
					return m, nil
				}
				m.Config.LanguageVersion = version.NixAttr
				m.Cursor = 0
				m.Tools = nix.CommonTools
				m.CurrentScreen = ScreenPackageSelector
				return m, nil
			}
		}
	}
	return m, nil
}

// isVersionSupported reports whether the given Python NixAttr is available in the selected channel.
func (m Model) isVersionSupported(nixAttr string) bool {
	for _, supported := range m.SelectedNixpkgs.SupportedPythonVersions {
		if supported == nixAttr {
			return true
		}
	}
	return false
}

// updatePackageSelector handles multi-select package selection
func (m Model) updatePackageSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle text input when adding custom package
	if m.AddingCustomPackage {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				// Add custom package if input is not empty
				value := m.TextInput.Value()
				if value != "" {
					// Add to packages list
					newPkg := models.Package{
						Name:     value,
						NixAttr:  value,
						Description: "Custom package",
					}
					m.Packages = append(m.Packages, newPkg)
					// Mark as selected
					m.SelectedPackages[value] = true
					m.TextInput.SetValue("")
				}
				return m, nil
			case "esc":
				// Cancel adding custom package
				m.AddingCustomPackage = false
				m.TextInput.SetValue("")
				m.TextInput.Blur()
				return m, nil
			}
		}
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, cmd
	}

	// Handle PyPI overlay: cursor navigation + text input to add more
	if m.AddingPyPIPackage {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.TextInput.Value() == "" && m.PyPICursor > 0 {
					m.PyPICursor--
				}
				return m, nil
			case "down", "j":
				if m.TextInput.Value() == "" && m.PyPICursor < len(m.PyPIPackages)-1 {
					m.PyPICursor++
				}
				return m, nil
			case " ":
				// Toggle selection of package under cursor (only when input empty)
				if m.TextInput.Value() == "" && len(m.PyPIPackages) > 0 {
					pkg := m.PyPIPackages[m.PyPICursor]
					m.SelectedPyPI[pkg] = !m.SelectedPyPI[pkg]
				}
				return m, nil
			case "enter":
				value := m.TextInput.Value()
				if value != "" {
					m.PyPIPackages = append(m.PyPIPackages, value)
					m.SelectedPyPI[value] = true
					m.PyPICursor = len(m.PyPIPackages) - 1
					m.TextInput.SetValue("")
				} else {
					// Empty enter closes the overlay
					m.AddingPyPIPackage = false
					m.TextInput.Blur()
				}
				return m, nil
			case "esc":
				m.AddingPyPIPackage = false
				m.TextInput.SetValue("")
				m.TextInput.Blur()
				return m, nil
			}
		}
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, cmd
	}

	// Handle env var overlay: cursor navigation + text input to add more
	if m.AddingEnvVar {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.TextInput.Value() == "" && m.EnvVarCursor > 0 {
					m.EnvVarCursor--
				}
				return m, nil
			case "down", "j":
				if m.TextInput.Value() == "" && m.EnvVarCursor < len(m.EnvVars)-1 {
					m.EnvVarCursor++
				}
				return m, nil
			case " ":
				if m.TextInput.Value() == "" && len(m.EnvVars) > 0 {
					v := m.EnvVars[m.EnvVarCursor]
					m.SelectedEnvVars[v] = !m.SelectedEnvVars[v]
				}
				return m, nil
			case "enter":
				value := m.TextInput.Value()
				if value != "" {
					m.EnvVars = append(m.EnvVars, value)
					m.SelectedEnvVars[value] = true
					m.EnvVarCursor = len(m.EnvVars) - 1
					m.TextInput.SetValue("")
				} else {
					m.AddingEnvVar = false
					m.TextInput.Blur()
				}
				return m, nil
			case "esc":
				m.AddingEnvVar = false
				m.TextInput.SetValue("")
				m.TextInput.Blur()
				return m, nil
			}
		}
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, cmd
	}

	// Normal package selection
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Packages)-1 {
				m.Cursor++
			}
		case "left", "h":
			if step := m.columnStep(m.Packages); step > 0 && m.Cursor >= step {
				m.Cursor -= step
			}
		case "right", "l":
			if step := m.columnStep(m.Packages); step > 0 && m.Cursor+step < len(m.Packages) {
				m.Cursor += step
			}
		case " ": // Spacebar toggles selection
			if m.Cursor >= 0 && m.Cursor < len(m.Packages) {
				pkg := m.Packages[m.Cursor]
				m.SelectedPackages[pkg.NixAttr] = !m.SelectedPackages[pkg.NixAttr]
			}
		case "a": // Select all
			for _, pkg := range m.Packages {
				m.SelectedPackages[pkg.NixAttr] = true
			}
		case "n": // Select none
			m.SelectedPackages = make(map[string]bool)
		case "c": // Add custom nixpkg package
			m.AddingCustomPackage = true
			m.TextInput.SetValue("")
			m.TextInput.Focus()
			return m, nil
		case "p": // Open PyPI overlay
			m.AddingPyPIPackage = true
			m.PyPICursor = len(m.PyPIPackages) - 1
			if m.PyPICursor < 0 {
				m.PyPICursor = 0
			}
			m.TextInput.SetValue("")
			m.TextInput.Focus()
			return m, nil
		case "e": // Open env var overlay
			m.AddingEnvVar = true
			m.EnvVarCursor = len(m.EnvVars) - 1
			if m.EnvVarCursor < 0 {
				m.EnvVarCursor = 0
			}
			m.TextInput.SetValue("")
			m.TextInput.Focus()
			return m, nil
		case "enter":
			// Collect selected packages
			m.Config.Packages = make([]string, 0)
			for _, pkg := range m.Packages {
				if m.SelectedPackages[pkg.NixAttr] {
					m.Config.Packages = append(m.Config.Packages, pkg.NixAttr)
				}
			}
			// Collect only selected PyPI packages
			m.Config.PyPIPackages = make([]string, 0)
			for _, pkg := range m.PyPIPackages {
				if m.SelectedPyPI[pkg] {
					m.Config.PyPIPackages = append(m.Config.PyPIPackages, pkg)
				}
			}
			// Collect only selected env vars
			m.Config.EnvVars = make([]string, 0)
			for _, v := range m.EnvVars {
				if m.SelectedEnvVars[v] {
					m.Config.EnvVars = append(m.Config.EnvVars, v)
				}
			}
			// Reset cursor for tool selection
			m.Cursor = 0
			m.Tools = nix.CommonTools
			m.CurrentScreen = ScreenToolSelector
			return m, nil
		}
	}
	return m, nil
}

// updateToolSelector handles multi-select tool selection
func (m Model) updateToolSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle text input when adding custom tool
	if m.AddingCustomTool {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				value := m.TextInput.Value()
				if value != "" {
					newTool := models.Package{
						Name:        value,
						NixAttr:     value,
						Description: "Custom tool",
					}
					m.Tools = append(m.Tools, newTool)
					m.SelectedTools[value] = true
					m.TextInput.SetValue("")
				}
				return m, nil
			case "esc":
				m.AddingCustomTool = false
				m.TextInput.SetValue("")
				m.TextInput.Blur()
				return m, nil
			}
		}
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Tools)-1 {
				m.Cursor++
			}
		case "left", "h":
			if step := m.columnStep(m.Tools); step > 0 && m.Cursor >= step {
				m.Cursor -= step
			}
		case "right", "l":
			if step := m.columnStep(m.Tools); step > 0 && m.Cursor+step < len(m.Tools) {
				m.Cursor += step
			}
		case " ": // Spacebar toggles selection
			if m.Cursor >= 0 && m.Cursor < len(m.Tools) {
				tool := m.Tools[m.Cursor]
				m.SelectedTools[tool.NixAttr] = !m.SelectedTools[tool.NixAttr]
			}
		case "a": // Select all
			for _, tool := range m.Tools {
				m.SelectedTools[tool.NixAttr] = true
			}
		case "n": // Select none
			m.SelectedTools = make(map[string]bool)
		case "c": // Add custom tool from nixpkgs
			m.AddingCustomTool = true
			m.TextInput.SetValue("")
			m.TextInput.Focus()
			return m, nil
		case "enter":
			// Collect selected tools
			m.Config.Tools = make([]string, 0)
			for _, tool := range m.Tools {
				if m.SelectedTools[tool.NixAttr] {
					m.Config.Tools = append(m.Config.Tools, tool.NixAttr)
				}
			}
			// Reset cursor for feature selection
			m.Cursor = 0
			// Transition to feature selector or skip if no features
			if len(m.Features) > 0 {
				m.CurrentScreen = ScreenFeatureSelector
			} else {
				m.CurrentScreen = ScreenConfirmation
			}
			return m, nil
		}
	}
	return m, nil
}

// updateFeatureSelector handles multi-select feature selection
func (m Model) updateFeatureSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Features)-1 {
				m.Cursor++
			}
		case " ": // Spacebar toggles selection
			if m.Cursor >= 0 && m.Cursor < len(m.Features) {
				feature := m.Features[m.Cursor]
				// Toggle all NixAttrs for this feature
				key := feature.Name
				m.SelectedFeatures[key] = !m.SelectedFeatures[key]
			}
		case "enter":
			// Collect selected features' NixAttrs (skip the FHS sentinel)
			m.Config.EnabledFeatures = make([]string, 0)
			m.Config.UseFHS = m.SelectedFeatures["FHS Environment"]
			for _, feature := range m.Features {
				if m.SelectedFeatures[feature.Name] && feature.Name != "FHS Environment" {
					m.Config.EnabledFeatures = append(m.Config.EnabledFeatures, feature.NixAttrs...)
				}
			}
			m.CurrentScreen = ScreenConfirmation
			return m, nil
		}
	}
	return m, nil
}

func (m Model) updateConfirmation(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.AskingOverwrite {
			switch msg.String() {
			case "y", "enter":
				m.AskingOverwrite = false
				return m.writeAndComplete(), nil
			case "n", "q", "esc":
				m.AskingOverwrite = false
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "enter", "y":
			// Check if output file already exists
			if m.Config.Mode != "quick" {
				if _, err := os.Stat(m.Config.OutputPath); err == nil {
					m.AskingOverwrite = true
					return m, nil
				}
			}
			return m.writeAndComplete(), nil
		}
	}
	return m, nil
}

// writeAndComplete generates/initialises the flake and transitions to the completion screen.
func (m Model) writeAndComplete() Model {
	if m.Config.Mode == "quick" {
		m.Err = nix.InitializeTemplate(m.Config.SelectedTemplate, ".")
	} else {
		content, err := nix.GenerateFlake(m.Config)
		if err != nil {
			m.Err = err
			m.CurrentScreen = ScreenCompletion
			return m
		}
		m.Err = nix.WriteFlake(content, m.Config.OutputPath)
	}
	m.CurrentScreen = ScreenCompletion
	return m
}

// flakeDir returns the absolute directory containing the output flake.nix.
func (m Model) flakeDir() string {
	abs, err := filepath.Abs(m.Config.OutputPath)
	if err != nil {
		return "."
	}
	return filepath.Dir(abs)
}

// isFlakeTrackedByGit returns true if the output file is tracked in its git repo.
func (m Model) isFlakeTrackedByGit() bool {
	abs, err := filepath.Abs(m.Config.OutputPath)
	if err != nil {
		return true // assume fine if we can't resolve
	}
	check := exec.Command("git", "-C", m.flakeDir(), "ls-files", "--error-unmatch", filepath.Base(abs))
	check.Stdout = nil
	check.Stderr = nil
	return check.Run() == nil
}

func (m Model) updateCompletion(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Only handle keys if there was no error
	if m.Err == nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// Sub-prompt: flake.nix is untracked, confirm running without git-add
			if m.AskingGitAdd {
				switch msg.String() {
				case "y":
					m.AskingGitAdd = false
					dir := m.flakeDir()
					cmd := exec.Command("nix", "develop", "path:"+dir)
					cmd.Dir = dir
					return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
						return tea.Quit()
					})
				case "n", "q", "esc":
					m.AskingGitAdd = false
					return m, nil
				}
				return m, nil
			}

			switch msg.String() {
			case "y":
				if !m.isFlakeTrackedByGit() {
					m.AskingGitAdd = true
					return m, nil
				}
				dir := m.flakeDir()
				cmd := exec.Command("nix", "develop", "path:"+dir)
				cmd.Dir = dir
				return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
					return tea.Quit()
				})
			case "e":
				editor := os.Getenv("EDITOR")
				if editor == "" {
					editor = "vim"
				}
				cmd := exec.Command(editor, m.Config.OutputPath)
				return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
					return nil
				})
			case "n", "q":
				m.Quitting = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}
