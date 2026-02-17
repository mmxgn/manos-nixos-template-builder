{
  description = "Interactive TUI for creating Nix flake.nix files";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "manos-nix-template-builder";
          version = "0.1.0";
          src = ./.;

          # Will be updated after first build failure
          vendorHash = null;

          meta = with pkgs.lib; {
            description = "Interactive TUI for creating Nix flake.nix files";
            license = licenses.mit;
          };
        };

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/manos-nix-template-builder";
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            golangci-lint
          ];

          shellHook = ''
            echo "🔨 Nix Template Chooser development environment loaded"
            echo "Go version: $(go version)"
          '';
        };
      }
    );
}
