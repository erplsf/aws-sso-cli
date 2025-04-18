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

          vendorHash = "sha256-Nsw75DIyX60PkgpjGheSNckYpXfEahHIGD3LuLjKBHs=";

          checkFlags = [ "-skip=^TestDetectShellBash$" ];
        };
      }));
}
