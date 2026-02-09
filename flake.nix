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
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
          ];

          shellHook = pre-commit-check.shellHook + '''';
        };
      }
    );
}
