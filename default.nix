# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{ buildGoModule, libpcap }:
buildGoModule {
  name = "gont";
  src = ./.;
  vendorHash = "sha256-xZuaTS0NeXxoiz9qsrICAlGKgSFIfR/KI9y7oXvQaGg=";
  buildInputs = [ libpcap ];
  doCheck = false;
}
