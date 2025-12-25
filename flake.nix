{
  description = "Flake for Caddy plugin development";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

    git-hooks-nix.url = "github:cachix/git-hooks.nix?ref=master";
    git-hooks-nix.inputs.nixpkgs.follows = "nixpkgs";

    flake-parts.url = "github:hercules-ci/flake-parts?ref=main";
  };

  outputs =
    { flake-parts, ... }@inputs:

    flake-parts.lib.mkFlake { inherit inputs; } (
      top@{ config, ... }:

      {
        imports = [
          inputs.git-hooks-nix.flakeModule
        ];

        systems = [ "x86_64-linux" ];

        perSystem =
          {
            config,
            pkgs,
            system,
            ...
          }:
          let
            goPackage = pkgs.go;
            buildGoModule = pkgs.buildGoModule.override { go = goPackage; };
            buildWithSpecificGo = pkg: pkg.override { inherit buildGoModule; };
          in
          {
            pre-commit.settings.hooks = {
              treefmt = {
                enable = true;
                settings.formatters = [
                  (buildWithSpecificGo pkgs.gofumpt)
                  pkgs.nixfmt-rfc-style
                  pkgs.toml-sort
                ];
              };
              typos.enable = true;
              reuse.enable = true;
            };

            devShells.default = pkgs.mkShell {
              name = "caddy-plugin-devshell";
              shellHook = config.pre-commit.installationScript;
              buildInputs = config.pre-commit.settings.enabledPackages ++ [
                config.pre-commit.settings.package

                goPackage
                (buildWithSpecificGo pkgs.gotools)
                (buildWithSpecificGo pkgs.xcaddy)
                (buildWithSpecificGo pkgs.pkgsite)
              ];
            };
          };
      }
    );
}
