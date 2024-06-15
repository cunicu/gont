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
        pkgs = import nixpkgs {
          inherit system;
          inherit overlays;
        };

        overlay = final: prev: { gont = final.callPackage ./default.nix { }; };
        overlays = [ overlay ];

        luaPkgs = {
          lua-struct = pkgs.lua.pkgs.buildLuarocksPackage {
            pname = "lua-struct";
            version = "0.9.2-1";

            src = pkgs.fetchFromGitHub {
              owner = "iryont";
              repo = "lua-struct";
              rev = "0.9.2-1";
              hash = "sha256-tyZU+Cm/f3urG3A5nBFC2NZ9nwtWh1yD4Oj0MHRtDlI=";
            };
          };
        };
      in
      {
        inherit overlays;

        packages.default = pkgs.gont;

        devShell =
          let
            tshark = pkgs.tshark.overrideAttrs (
              final: prev: {
                # Make sure we can load our own Lua dissector plugin
                # by Wiresharks Lua interpreter when tests are executed by root
                postFixup = ''
                  echo "run_user_scripts_when_superuser = true" >> $out/lib/wireshark/plugins/init.lua
                '';
              }
            );
          in
          pkgs.mkShell {
            packages = with pkgs // luaPkgs; [
              golangci-lint
              reuse
              traceroute
              gnumake
              tshark
              gont
              lua-struct
            ];

            inputsFrom = with pkgs; [ gont ];

            hardeningDisable = [ "fortify" ];
          };

        formatter = pkgs.nixfmt-rfc-style;
      }
    );
}
