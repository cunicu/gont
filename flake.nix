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
        packages.gont = pkgs.buildGoModule {
          name = "gont";
          src = ./.;
          vendorHash = "sha256-+6NZh6mBf1M7TU9WivWLesmd4CkUEkhXKEioMuHfo9Y=";
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
