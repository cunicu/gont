# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{ buildGoModule, libpcap }:
buildGoModule {
  name = "gont";
  src = ./.;
  vendorHash = "sha256-IXTpMzTrWRH10vB6hRsMf7ilT5tUG/EPJbYLO+8d9Ik=";
  buildInputs = [ libpcap ];
  doCheck = false;
}
