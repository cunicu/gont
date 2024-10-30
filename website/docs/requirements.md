---
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
sidebar_position: 2
---

# Requirements

-   Go >= 1.19
-   A moderate recent Linux kernel (&gt;= 4.9)
    -   `mnt` and `net` namespace support
-   Root access / `NET_ADMIN` caps
-   Traceroute userspace tool
-   [libpcap](https://www.tcpdump.org/) for packet captures
-   [Systemd](https://systemd.io/) for CGroups
