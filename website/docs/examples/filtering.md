---
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Firewall

Powered by [Netfilter's nftables](https://nftables.org/).

## Add some firewall rules for a host

```go
import fopt "github.com/cunicu/gont/v2/options/filter"

_, src, _ := net.ParseCIDR("10.0.0.1/32")

host1, _ := network.AddHost("host1",
  opt.Filter(
    gont.FilterInput,
      fopt.Source(src),
      fopt.Protocol(unix.AF_INET),
      fopt.TransportProtocol(unix.IPPROTO_TCP),
      fopt.SourcePortRange(0, 1024)))
```

