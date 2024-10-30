# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{ buildGoModule, libpcap }:
buildGoModule {
  name = "gont";
  src = ./.;
  vendorHash = "sha256-EAwP8nNyS6lnLi/OBxxdZzePIiy30l6uFr1Z8SPAllA=";
  buildInputs = [ libpcap ];
  doCheck = false;
}
