{
  description = "aws-sso-cli flake";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.gomod2nix.url = "github:JonathanLorimer/gomod2nix/jonathan/update-go";

  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    (flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };

      in {
        packages.default = pkgs.callPackage ./. { };
        devShells.default = import ./shell.nix { inherit pkgs; };
      }));
}
