package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mmxgn/nix-template-chooser/internal/models"
	"github.com/mmxgn/nix-template-chooser/internal/nix"
)

// View renders the current screen
func (m Model) View() string {
	if m.Quitting {
		return ""
	}

	// Route to screen-specific views
	switch m.CurrentScreen {
	case ScreenModeSelection:
		return m.viewModeSelection()
	case ScreenTemplateBrowser:
		return m.viewTemplateBrowser()
	case ScreenLanguageSelector:
		return m.viewLanguageSelector()
	case ScreenTemplateOrCustom:
		return m.viewTemplateOrCustom()
	case ScreenNixpkgsSelector:
		return m.viewNixpkgsSelector()
	case ScreenVersionSelector:
		return m.viewVersionSelector()
	case ScreenPackageSelector:
		return m.viewPackageSelector()
	case ScreenToolSelector:
		return m.viewToolSelector()
	case ScreenFeatureSelector:
		return m.viewFeatureSelector()
	case ScreenConfirmation:
		return m.viewConfirmation()
	case ScreenCompletion:
		return m.viewCompletion()
	}

	return "Unknown screen"
}

// Screen-specific view functions
func (m Model) viewModeSelection() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(TitleStyle.Render("Nix Flake Template Chooser"))
	s.WriteString("\n")
	s.WriteString(SubtitleStyle.Render("Create your perfect Nix development environment"))
	s.WriteString("\n\n")

	s.WriteString(m.ModeList.View())
	s.WriteString("\n\n")
	s.WriteString(HelpStyle.Render("Press enter to select, q to quit"))

	return s.String()
}

func (m Model) viewTemplateBrowser() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(TitleStyle.Render("Browse NixOS Templates"))
	s.WriteString("\n\n")

	if m.Err != nil {
		s.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v", m.Err)))
		s.WriteString("\n\n")
		s.WriteString(HelpStyle.Render("Press esc to go back, q to quit"))
		return s.String()
	}

	s.WriteString(m.TemplateList.View())
	s.WriteString("\n\n")
	s.WriteString(HelpStyle.Render("Press enter to select, esc to go back, q to quit"))
	return s.String()
}

func (m Model) viewLanguageSelector() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(TitleStyle.Render("Select Programming Language"))
	s.WriteString("\n\n")
	s.WriteString(m.LanguageList.View())
	s.WriteString("\n\n")
	s.WriteString(HelpStyle.Render("Press enter to select, esc to go back, q to quit"))
	return s.String()
}

func (m Model) viewTemplateOrCustom() string {
	var s strings.Builder
	s.WriteString("\n")

	// Get the language display name
	langName := m.Config.Language
	if langDef, ok := nix.LanguageDefinitions[m.Config.Language]; ok {
		langName = langDef.Name
	}

	s.WriteString(TitleStyle.Render(fmt.Sprintf("%s Configuration", langName)))
	s.WriteString("\n")
	s.WriteString(SubtitleStyle.Render("Choose a preset template or customize your own"))
	s.WriteString("\n\n")
	s.WriteString(m.TemplateOrCustomList.View())
	s.WriteString("\n\n")
	s.WriteString(HelpStyle.Render("Press enter to select, esc to go back, q to quit"))
	return s.String()
}

func (m Model) viewNixpkgsSelector() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(TitleStyle.Render("Select nixpkgs Channel"))
	s.WriteString("\n")
	s.WriteString(SubtitleStyle.Render("The nixpkgs version determines which packages are available"))
	s.WriteString("\n\n")
	s.WriteString(m.NixpkgsChannelList.View())
	s.WriteString("\n\n")
	s.WriteString(HelpStyle.Render("Press enter to select, esc to go back, q to quit"))
	return s.String()
}

func (m Model) viewVersionSelector() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(TitleStyle.Render("Select Python Version"))
	s.WriteString("\n")
	s.WriteString(SubtitleStyle.Render(fmt.Sprintf("nixpkgs: %s", m.SelectedNixpkgs.Name)))
	s.WriteString("\n\n")

	for i, version := range m.Versions {
		cursorStr := "  "
		if i == m.Cursor {
			cursorStr = "> "
		}

		supported := m.isVersionSupported(version.NixAttr)

		var line string
		if supported {
			label := version.Name
			if version.IsDefault {
				label += " (default)"
			}
			if i == m.Cursor {
				label = SelectedItemStyle.Render(label)
			}
			line = fmt.Sprintf("%s%s", cursorStr, label)
		} else {
			line = DisabledStyle.Render(fmt.Sprintf("%s%s  [Not supported for this nixpkgs version]", cursorStr, version.Name))
		}
		s.WriteString(line)
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(InfoStyle.Render("Note: selecting a non-default Python version may lead to significantly longer compile times."))
	s.WriteString("\n\n")
	s.WriteString(HelpStyle.Render("up/down: navigate | Enter: select | Esc: back"))
	return s.String()
}

func (m Model) viewPackageSelector() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(TitleStyle.Render("Select Packages"))
	s.WriteString("\n\n")

	// Show text input overlay if adding custom package
	if m.AddingCustomPackage {
		s.WriteString(SubtitleStyle.Render("Add Custom Nixpkg Package"))
		s.WriteString("\n\n")
		s.WriteString("Package name (e.g., 'scipy', 'pillow'): ")
		s.WriteString(m.TextInput.View())
		s.WriteString("\n\n")
		s.WriteString(HelpStyle.Render("Enter: add | Esc: cancel"))
		return s.String()
	}

	// Show PyPI overlay: selectable list + text input to add more
	if m.AddingPyPIPackage {
		s.WriteString(SubtitleStyle.Render("PyPI Packages"))
		s.WriteString("\n\n")
		if len(m.PyPIPackages) > 0 {
			for i, pkg := range m.PyPIPackages {
				cursor := "  "
				if i == m.PyPICursor {
					cursor = "> "
				}
				checkbox := "[ ]"
				if m.SelectedPyPI[pkg] {
					checkbox = "[x]"
				}
				line := fmt.Sprintf("%s%s %s", cursor, checkbox, pkg)
				if i == m.PyPICursor {
					s.WriteString(SelectedItemStyle.Render(line))
				} else {
					s.WriteString(line)
				}
				s.WriteString("\n")
			}
			s.WriteString("\n")
		}
		s.WriteString("Add package (e.g., 'beautifulsoup4', 'openai'): ")
		s.WriteString(m.TextInput.View())
		s.WriteString("\n\n")
		s.WriteString(HelpStyle.Render("Enter: add (or done if empty) | Space: toggle | up/down: navigate | Esc: cancel"))
		return s.String()
	}

	// Show env var overlay
	if m.AddingEnvVar {
		s.WriteString(SubtitleStyle.Render("Environment Variables"))
		s.WriteString("\n\n")
		if len(m.EnvVars) > 0 {
			for i, v := range m.EnvVars {
				cursor := "  "
				if i == m.EnvVarCursor {
					cursor = "> "
				}
				checkbox := "[ ]"
				if m.SelectedEnvVars[v] {
					checkbox = "[x]"
				}
				line := fmt.Sprintf("%s%s %s", cursor, checkbox, v)
				if i == m.EnvVarCursor {
					s.WriteString(SelectedItemStyle.Render(line))
				} else {
					s.WriteString(line)
				}
				s.WriteString("\n")
			}
			s.WriteString("\n")
		}
		s.WriteString("Variable name (e.g., 'API_KEY', 'DATABASE_URL'): ")
		s.WriteString(m.TextInput.View())
		s.WriteString("\n\n")
		s.WriteString(HelpStyle.Render("Enter: add (or done if empty) | Space: toggle | up/down: navigate | Esc: cancel"))
		return s.String()
	}

	// Render multi-select list
	s.WriteString(m.renderMultiSelectList(m.Packages, m.SelectedPackages))

	// Show count
	count := 0
	for _, selected := range m.SelectedPackages {
		if selected {
			count++
		}
	}
	s.WriteString("\n")
	s.WriteString(fmt.Sprintf("Selected: %d/%d packages\n", count, len(m.Packages)))
	s.WriteString("\n")

	// Show PyPI packages summary
	if len(m.PyPIPackages) > 0 {
		sel := 0
		for _, pkg := range m.PyPIPackages {
			if m.SelectedPyPI[pkg] {
				sel++
			}
		}
		s.WriteString(SelectedItemStyle.Render(fmt.Sprintf("PyPI packages: %d/%d selected", sel, len(m.PyPIPackages))))
		s.WriteString(" (press p to manage)\n")
	}

	// Show env vars summary
	if len(m.EnvVars) > 0 {
		sel := 0
		for _, v := range m.EnvVars {
			if m.SelectedEnvVars[v] {
				sel++
			}
		}
		s.WriteString(SelectedItemStyle.Render(fmt.Sprintf("Env vars: %d/%d selected", sel, len(m.EnvVars))))
		s.WriteString(" (press e to manage)\n")
	}

	s.WriteString("\n")
	s.WriteString(HelpStyle.Render("Space: toggle | up/down/left/right: navigate | a: all | n: none | c: custom nixpkg | p: PyPI | e: env vars | Enter: continue | Esc: back"))
	return s.String()
}

func (m Model) viewToolSelector() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(TitleStyle.Render("Select Development Tools"))
	s.WriteString("\n\n")

	// Show text input overlay if adding custom tool
	if m.AddingCustomTool {
		s.WriteString(SubtitleStyle.Render("Add Custom Nixpkg Tool"))
		s.WriteString("\n\n")
		s.WriteString("Tool name (e.g., 'ripgrep', 'jq', 'htop'): ")
		s.WriteString(m.TextInput.View())
		s.WriteString("\n\n")
		s.WriteString(HelpStyle.Render("Enter: add | Esc: cancel"))
		return s.String()
	}

	// Render multi-select list
	s.WriteString(m.renderMultiSelectList(m.Tools, m.SelectedTools))

	// Show count
	count := 0
	for _, selected := range m.SelectedTools {
		if selected {
			count++
		}
	}
	s.WriteString("\n")
	s.WriteString(fmt.Sprintf("Selected: %d/%d tools\n", count, len(m.Tools)))
	s.WriteString("\n")
	s.WriteString(HelpStyle.Render("Space: toggle | up/down/left/right: navigate | a: all | n: none | c: add custom tool | Enter: continue | Esc: back"))
	return s.String()
}

func (m Model) viewFeatureSelector() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(TitleStyle.Render("Select Features (Optional)"))
	s.WriteString("\n\n")

	// Render features as multi-select
	for i, feature := range m.Features {
		cursorStr := "  "
		if i == m.Cursor {
			cursorStr = "> "
		}

		checkbox := UncheckedStyle.Render("[ ]")
		if m.SelectedFeatures[feature.Name] {
			checkbox = CheckboxStyle.Render("[x]")
		}

		itemText := feature.Name
		if i == m.Cursor {
			itemText = SelectedItemStyle.Render(itemText)
		}

		s.WriteString(fmt.Sprintf("%s%s %s - %s\n", cursorStr, checkbox, itemText, feature.Description))
	}

	s.WriteString("\n")
	s.WriteString(HelpStyle.Render("Space: toggle, up/down: navigate, Enter: continue, Esc: back"))
	return s.String()
}

// columnStep returns the number of items per column for the current terminal height,
// or 0 when a single column is used. Used by update.go for left/right navigation.
func (m Model) columnStep(items []models.Package) int {
	if len(items) == 0 {
		return 0
	}
	availableRows := m.Height - 12
	if availableRows < 4 {
		availableRows = 4
	}
	if estimateListRows(items) <= availableRows {
		return 0 // single column
	}
	numCols := 2
	itemsPerCol := (len(items) + numCols - 1) / numCols
	if estimateListRows(items[:itemsPerCol]) > availableRows {
		numCols = 3
		itemsPerCol = (len(items) + numCols - 1) / numCols
	}
	return itemsPerCol
}

// estimateListRows estimates the number of terminal rows needed to render items in a single column.
func estimateListRows(items []models.Package) int {
	rows := 0
	lastCategory := ""
	for _, item := range items {
		if item.Category != "" && item.Category != lastCategory {
			if lastCategory != "" {
				rows++ // blank line before new category
			}
			rows++ // category header line
			lastCategory = item.Category
		}
		rows++ // item line
	}
	return rows
}

// renderColItems renders a subset of items as a single column, using globalStart to
// correctly identify the cursor position within the full items slice.
func (m Model) renderColItems(items []models.Package, selections map[string]bool, globalStart int) string {
	var s strings.Builder
	lastCategory := ""
	for i, item := range items {
		globalIdx := globalStart + i

		if item.Category != "" && item.Category != lastCategory {
			if lastCategory != "" {
				s.WriteString("\n")
			}
			s.WriteString(SubtitleStyle.Render(item.Category))
			s.WriteString("\n")
			lastCategory = item.Category
		}

		cursorStr := "  "
		if globalIdx == m.Cursor {
			cursorStr = "> "
		}

		checkbox := UncheckedStyle.Render("[ ]")
		if selections[item.NixAttr] {
			checkbox = CheckboxStyle.Render("[x]")
		}

		itemText := item.Name
		if globalIdx == m.Cursor {
			itemText = SelectedItemStyle.Render(itemText)
		}

		s.WriteString(fmt.Sprintf("%s%s %s - %s\n", cursorStr, checkbox, itemText, item.Description))
	}
	return s.String()
}

// renderMultiSelectList renders a multi-select list with checkboxes and category headers.
// When the list is too tall for the terminal, it splits into multiple columns.
func (m Model) renderMultiSelectList(items []models.Package, selections map[string]bool) string {
	if len(items) == 0 {
		return ""
	}

	// Reserve space for title, subtitle, count line, help line, etc.
	availableRows := m.Height - 12
	if availableRows < 4 {
		availableRows = 4
	}

	totalRows := estimateListRows(items)

	// Determine number of columns needed
	numCols := 1
	if totalRows > availableRows {
		numCols = 2
	}
	itemsPerCol := (len(items) + numCols - 1) / numCols
	// If 2 columns still overflow, use 3
	if numCols == 2 && estimateListRows(items[:itemsPerCol]) > availableRows {
		numCols = 3
		itemsPerCol = (len(items) + numCols - 1) / numCols
	}

	if numCols == 1 {
		return m.renderColItems(items, selections, 0)
	}

	// Calculate column width (leave 2 chars padding between columns)
	colWidth := (m.Width - (numCols - 1)*2) / numCols
	if colWidth < 20 {
		colWidth = 20
	}

	cols := make([]string, 0, numCols)
	for c := 0; c < numCols; c++ {
		start := c * itemsPerCol
		end := start + itemsPerCol
		if end > len(items) {
			end = len(items)
		}
		if start >= len(items) {
			break
		}
		colContent := m.renderColItems(items[start:end], selections, start)
		cols = append(cols, lipgloss.NewStyle().Width(colWidth).Render(colContent))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

func (m Model) viewConfirmation() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(TitleStyle.Render("Review Your Configuration"))
	s.WriteString("\n\n")

	if m.Config.Mode == "quick" {
		s.WriteString(fmt.Sprintf("Mode: %s\n", SelectedItemStyle.Render("NixOS Templates")))
		s.WriteString(fmt.Sprintf("Template: %s\n", SelectedItemStyle.Render(m.Config.SelectedTemplate)))
		s.WriteString(fmt.Sprintf("\nThis will initialize the '%s' template in the current directory.\n", m.Config.SelectedTemplate))
	} else {
		s.WriteString(fmt.Sprintf("Mode: %s\n", SelectedItemStyle.Render("Python")))

		// Show template name if preset was used
		if m.Config.TemplateName != "" {
			s.WriteString(fmt.Sprintf("Template: %s\n", SelectedItemStyle.Render(m.Config.TemplateName)))
		}

		s.WriteString(fmt.Sprintf("Version: %s\n", SelectedItemStyle.Render(m.Config.LanguageVersion)))

		// Show packages
		if len(m.Config.Packages) > 0 {
			s.WriteString(fmt.Sprintf("Packages (%d): ", len(m.Config.Packages)))
			s.WriteString(strings.Join(m.Config.Packages, ", "))
			s.WriteString("\n")
		} else {
			s.WriteString("Packages: (none)\n")
		}

		// Show PyPI packages
		if len(m.Config.PyPIPackages) > 0 {
			s.WriteString(fmt.Sprintf("PyPI packages (%d): ", len(m.Config.PyPIPackages)))
			s.WriteString(strings.Join(m.Config.PyPIPackages, ", "))
			s.WriteString("\n")
		}

		// Show env vars
		if len(m.Config.EnvVars) > 0 {
			s.WriteString(fmt.Sprintf("Env vars (%d): ", len(m.Config.EnvVars)))
			s.WriteString(strings.Join(m.Config.EnvVars, ", "))
			s.WriteString("\n")
		}

		// Show tools
		if len(m.Config.Tools) > 0 {
			s.WriteString(fmt.Sprintf("Tools (%d): ", len(m.Config.Tools)))
			s.WriteString(strings.Join(m.Config.Tools, ", "))
			s.WriteString("\n")
		} else {
			s.WriteString("Tools: (none)\n")
		}

		// Show features
		if len(m.Config.EnabledFeatures) > 0 {
			s.WriteString(fmt.Sprintf("Features: %s\n", strings.Join(m.Config.EnabledFeatures, ", ")))
		}
		if m.Config.UseFHS {
			s.WriteString("Shell type: FHS Environment (buildFHSEnv)\n")
		}

		s.WriteString(fmt.Sprintf("\nOutput: %s\n", SelectedItemStyle.Render(m.Config.OutputPath)))
	}

	s.WriteString("\n")
	if m.AskingOverwrite {
		s.WriteString(InfoStyle.Render(fmt.Sprintf("%s already exists.", m.Config.OutputPath)))
		s.WriteString("\n")
		s.WriteString(HelpStyle.Render("y/enter: overwrite | n/esc: cancel"))
	} else {
		s.WriteString(HelpStyle.Render("Press enter to confirm, esc to go back, q to quit"))
	}
	return s.String()
}

func (m Model) viewCompletion() string {
	var s strings.Builder
	s.WriteString("\n\n")

	if m.Err != nil {
		s.WriteString(ErrorStyle.Render("Error"))
		s.WriteString("\n\n")
		s.WriteString(fmt.Sprintf("Failed to create flake: %v\n", m.Err))
		s.WriteString("\n")
		s.WriteString(HelpStyle.Render("Press q to quit"))
	} else {
		s.WriteString(SuccessStyle.Render("Done"))
		s.WriteString("\n\n")

		if m.Config.Mode == "quick" {
			s.WriteString(fmt.Sprintf("Template '%s' has been initialized.\n\n", m.Config.SelectedTemplate))
		} else {
			s.WriteString(fmt.Sprintf("Your flake.nix has been created at %s\n\n", m.Config.OutputPath))
		}

		if m.AskingGitAdd {
			s.WriteString(InfoStyle.Render("flake.nix is not tracked by Git."))
			s.WriteString("\n")
			s.WriteString("Nix requires files to be tracked before evaluating a flake.\n\n")
			s.WriteString(HelpStyle.Render("y: run 'nix develop path:.' anyway | n/esc: cancel"))
		} else {
			s.WriteString("Would you like to enter the development environment?\n\n")
			s.WriteString(HelpStyle.Render("y: run 'nix develop path:.' | e: edit flake.nix | n/q: quit"))
		}
	}

	return s.String()
}
