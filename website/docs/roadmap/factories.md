---
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Topology Factories

Generate complex network topologies from code

```go
createHost := func(pos int) (*gont.Host. error) {
  return network.AddHost(fmt.Sprintf("h%d", pos))
}

linkHosts := func(a, b *gont.Node) error {
  _, err := network.AddRouter(fmt.Sprintf("r%d", pos),
    gont.NewInterface("eth0", a, opt.AddressIPv4(10, 0, 0, a.Position, 24),
    gont.NewInterface("eth1", b, opt.AddressIPv4(10, 0, 0, b.Position, 24)
  )
  return err
}

topo.Linear(n, 100, createHost, linkHosts)

network.Nodes["h0"].Traceroute(network.Nodes["h99"])
```
