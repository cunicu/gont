# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  description = "Gont: A Go testing framework for distributed applications";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = {
    self,
    flake-utils,
    nixpkgs,
  }:
    flake-utils.lib.eachDefaultSystem
    (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in rec {
        packages.gont = pkgs.buildGoModule {
          name = "gont";
          src = ./.;
          vendorHash = "sha256-QOh1jBR7FL/fKFmJv7wGxuCghRLR3DV/0TzXd+bUFP0=";
          buildInputs = with pkgs; [
            libpcap
          ];
          doCheck = false;
        };

        devShell = pkgs.mkShell {
          packages = with pkgs; [
            golangci-lint
            reuse
            traceroute
            gnumake
            tshark
            packages.gont
          ];

          inputsFrom = [
            packages.gont
          ];
        };

        formatter = nixpkgs.alejandra;
      }
    );
}
