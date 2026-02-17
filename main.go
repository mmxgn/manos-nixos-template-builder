package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmxgn/nix-template-chooser/internal/tui"
)

func main() {
	outputFlag := flag.String("o", "", "output path for flake.nix")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: nix-template-chooser [OPTIONS] [PATH]\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  PATH    Where to write flake.nix (default: ./flake.nix)\n")
		fmt.Fprintf(os.Stderr, "          If the parent directory does not exist you will be asked to create it.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -o PATH    Same as the positional PATH argument\n")
		fmt.Fprintf(os.Stderr, "  -h         Show this help message\n")
	}
	flag.Parse()

	// Resolve output path: -o flag takes priority, then positional arg, then default
	outputPath := "./flake.nix"
	if *outputFlag != "" {
		outputPath = *outputFlag
	} else if flag.NArg() > 0 {
		outputPath = flag.Arg(0)
	}

	// If a non-default path was given, ensure the parent directory exists
	if outputPath != "./flake.nix" {
		dir := filepath.Dir(outputPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Printf("Directory %q does not exist. Create it? [y/N] ", dir)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			if !isYes(answer) {
				fmt.Fprintln(os.Stderr, "Aborted.")
				os.Exit(1)
			}
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "error: failed to create directory: %v\n", err)
				os.Exit(1)
			}
		}
	}

	m := tui.InitialModel()
	m.Config.OutputPath = outputPath

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func isYes(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "y" || s == "yes"
}
