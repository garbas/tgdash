{
  description = "tgdash - A terragrunt dashboard";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    pre-commit-hooks.url = "github:cachix/pre-commit-hooks.nix";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      pre-commit-hooks,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        pre-commit-check = pre-commit-hooks.lib.${system}.run {
          src = ./.;
          hooks = {
            nixfmt-rfc-style.enable = true;
            markdownlint = {
              enable = true;
              excludes = [
                "^blog/"
              ];
            };
            govet.enable = true;
            gofmt.enable = true;
            actionlint.enable = true;
            yamllint.enable = true;
            shellcheck = {
              enable = true;
              excludes = [ "^\\.envrc$" ];
            };
            convco-check = {
              enable = true;
              name = "convco-check";
              description = "Validate commit message follows Conventional Commits";
              entry = "${pkgs.bash}/bin/bash -c '${pkgs.coreutils}/bin/cat \"$1\" | ${pkgs.convco}/bin/convco check --from-stdin --strip' --";
              language = "system";
              pass_filenames = true;
              stages = [ "commit-msg" ];
            };
          };
        };
        tgdash = pkgs.buildGoModule {
          pname = "tgdash";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-TZa7C8KbBRZaD921T4sNVaqgHiRwCHDGdeaG7O9qeLg=";

          meta = {
            description = "Terminal dashboard for Terragrunt";
            homepage = "https://github.com/rok/tgdash";
            license = pkgs.lib.licenses.mit;
            mainProgram = "tgdash";
          };
        };
      in
      {
        checks = {
          pre-commit = pre-commit-check;
        };

        packages = {
          default = tgdash;
          tgdash = tgdash;
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            goreleaser
          ];

          shellHook = pre-commit-check.shellHook + "";
        };
      }
    );
}
