{
  description = "aws-sso-cli flake";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "aws-sso";
          version = "main";

          src = ./.;

          vendorHash = "sha256-K4VdK4E+ShQtkr+CwN2K0qPuL0wfsMFW1pLU/JSslxs=";

          checkFlags = [ "-skip=^TestDetectShellBash$" ];
        };
      }
    ));
}
