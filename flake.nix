{
  description = "aws-sso-cli flake";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    (flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; };

      in {
        packages.default = pkgs.buildGoModule {
          pname = "aws-sso";
          version = "main";

          src = ./.;

          vendorHash = "sha256-yhcYEzrDd2TujxW9ip0sSCT5vgSZ650ci4009nllNt0=";

          checkFlags = [ "-skip=^TestDetectShellBash$" ];
        };
      }));
}
