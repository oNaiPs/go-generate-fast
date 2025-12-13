{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; };
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            # Core Go tooling
            go
            gopls

            # Linting and testing
            golangci-lint
            gotestsum

            # Code generation tools (used by tests and e2e)
            protobuf
            protoc-gen-go

            # Build tools
            gnumake

            # Nix formatting
            nixfmt-classic
          ];

          shellHook = ''
            # Add Go bin directory to PATH for installed tools
            export PATH="$HOME/go/bin:$PATH"

            echo "go-generate-fast development environment"
            echo "Go version: $(go version)"
            echo ""
            echo "Run 'make install-deps' to install Go-based code generation tools"
          '';
        };
      });
}
