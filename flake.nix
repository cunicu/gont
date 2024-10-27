# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  description = "Gont: A Go testing framework for distributed applications";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    {
      self,
      flake-utils,
      nixpkgs,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      rec {
        packages.gont = pkgs.buildGo123Module {
          name = "gont";
          src = ./.;
          vendorHash = "sha256-IXTpMzTrWRH10vB6hRsMf7ilT5tUG/EPJbYLO+8d9Ik=";
          buildInputs = with pkgs; [ libpcap ];
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

          inputsFrom = [ packages.gont ];
        };

        formatter = nixpkgs.nixfmt-rfc-style;
      }
    );
}
