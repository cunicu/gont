# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{ buildGoModule, libpcap }:
buildGoModule {
  name = "gont";
  src = ./.;
  vendorHash = "sha256-9/TPK8MD1lA9s3jhKmHweY7quw383kHgrcL2XLyuQ54=";
  buildInputs = [ libpcap ];
  doCheck = false;
}
