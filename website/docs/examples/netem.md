---
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
sidebar_position: 10
---

# Network Emulation

Powered by Linux's Traffic Control: [Netem Qdisc](https://man7.org/linux/man-pages/man8/tc-netem.8.html).

## Attach a netem Qdisc to an interface

```go
import tcopt "github.com/cunicu/gont/v2/options/tc"

network.AddLink(
  gont.NewInterface("eth0", host1,
    opt.WithNetem(
      tcopt.Latency(50 * time.Millisecond),
      tcopt.Jitter(5 * time.Millisecond),
      tcopt.Loss(0.1),
    ),
    opt.AddressIP("10.0.0.1/24")),
  gont.NewInterface("eth0", host2,
    opt.AddressIP("10.0.0.2/24")),
)

host1.Ping(host2)
```
