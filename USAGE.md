# Nix Template Chooser - Usage Guide

## What is this?

An interactive TUI tool to help you create Nix flake.nix files easily, without having to remember the syntax.

## Features

- ✅ **Quick Mode**: Browse and use 22+ existing NixOS templates
- ✅ **Custom Mode**: Generate custom flake.nix with language selection
- ✅ **5 Languages Supported**: Go, Python, Rust, JavaScript/TypeScript, Haskell
- ✅ **Curated Packages**: Pre-selected common packages and tools for each language
- ✅ **Ready for `nix develop`**: Generated flakes work immediately

## How to Use

### Run Locally

```bash
# Build and run
nix build
./result/bin/manos-nix-template-builder

# Or run directly
nix run .#default
```

### Run from GitHub (after pushing)

```bash
nix run github:mmxgn/manos-nixos-template-builder
```

## Workflow

1. **Choose Mode**
   - Quick Mode: Use existing NixOS templates
   - Custom Mode: Build your own flake

2. **Quick Mode Flow**
   - Browse 22+ templates (go-hello, python, rust, etc.)
   - Select one
   - Confirm
   - Template initialized in current directory!

3. **Custom Mode Flow**
   - Select your language (Go, Python, Rust, JS/TS, Haskell)
   - Review configuration
   - Confirm
   - flake.nix created!

4. **Next Steps**
   ```bash
   nix develop           # Enter the development environment
   nix flake show        # See available outputs
   # Customize your flake.nix as needed
   ```

## Example Generated Flake (Custom Mode - Go)

```nix
{
  description = "Go development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; };
      in {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            golangci-lint
            air
            git
            jq
          ];

          shellHook = ''
            echo "🚀 Go development environment loaded!"
            echo "Go version: $(go version)"
          '';
        };
      }
    );
}
```

## Navigation

- **Arrow keys**: Navigate lists
- **Enter**: Select/confirm
- **Esc**: Go back to previous screen
- **q / Ctrl+C**: Quit

## Future Enhancements (Phase 2)

- Multi-select for packages and tools
- Dynamic package search
- Preview flake before writing
- More language support
- Configuration profiles

## Troubleshooting

**TUI doesn't start**: You need an interactive terminal. The TUI won't work in non-interactive environments.

**Template fetch fails**: Check your internet connection - templates are fetched from GitHub.

**Permission denied**: Make sure you have write permissions in the current directory.

## Development

```bash
# Enter dev environment
nix develop

# Run in development mode
go run main.go

# Build
go build -o manos-nix-template-builder  # binary name in local dev

# Rebuild Nix package
nix build
```

## Credits

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
