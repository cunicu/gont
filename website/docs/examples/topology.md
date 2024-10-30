---
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
sidebar_position: 1
---

# Networking Topology

## Two directly connected hosts

```go
import gont "github.com/cunicu/gont/v2/pkg"
import opt "github.com/cunicu/gont/v2/pkg/options"

 ...

network, _ := gont.NewNetwork("mynet")

host1, _ := network.AddHost("host1")
host2, _ := network.AddHost("host2")

network.AddLink(
  gont.NewInterface("eth0", host1, opt.AddressIP("10.0.0.1/24")),
  gont.NewInterface("eth0", host2, opt.AddressIP("10.0.0.2/24")))

host1.Ping(host2)
```

(We `opt` throughout these examples as an import alias)


## Lets add a L2 switch

```go
switch1, _ := network.AddSwitch("switch1")

host1, _ := network.AddHost("host1",
  gont.NewInterface("eth0", switch1,
    opt.AddressIP("10.0.0.1/24")))

host2, _ := network.AddHost("host2",
  gont.NewInterface("eth0", switch1,
    opt.AddressIP("10.0.0.2/24")))

host1.Ping(host2)
```


## How about a L3 router?

```go
switch1, _ := network.AddSwitch("switch1")
switch2, _ := network.AddSwitch("switch2")

host1, _ := network.AddHost("host1",
  gont.NewInterface("eth0", switch1,
    opt.AddressIP("10.0.0.2/24")))
host2, _ := network.AddHost("host2",
  gont.NewInterface("eth0", switch2,
    opt.AddressIP("10.0.1.2/24")))

network.AddRouter("router1",
  gont.NewInterface("eth0", sw1, opt.AddressIP("10.0.0.1/24")),
  gont.NewInterface("eth1", sw2, opt.AddressIP("10.0.1.1/24")))

host1.Ping(host2)
```


## Lets do some evil NATing 😈

```go
switch1, _ := network.AddSwitch("switch1")
switch2, _ := network.AddSwitch("switch2")

host1, _ := network.AddHost("host1",
  gont.NewInterface("eth0", switch1,
    opt.AddressIP("10.0.0.2/24")))
host2, _ := network.AddHost("host2",
  gont.NewInterface("eth0", switch2,
    opt.AddressIP("10.0.1.2/24")))

network.AddNAT("n1",
  gont.NewInterface("eth0", switch1,
    opt.SouthBound,
    opt.AddressIP("10.0.0.1/24")),
  gont.NewInterface("eth1", switch2,
    opt.NorthBound,
    opt.AddressIP("10.0.1.1/24")))

host1.Ping(host2)
```


## How about a whole chain of routers?

```go
var firstSwitch *gont.Switch = network.AddSwitch("switch0")
var lastSwitch  *gont.Switch = nil

for i := 1; i < 100; i++ {
  switchName := fmt.Printf("switch%d", i)
  routerName := fmt.Printf("router%d", i)

  newSwitch, _ := network.AddSwitch(switchName)

  network.AddRouter(routerName,
    gont.NewInterface("eth0", lastSwitch,
      opt.AddressIP("10.0.0.1/24")),
    gont.NewInterface("eth1", newSwitch,
      opt.AddressIP("10.0.1.1/24"))
  )

  lastSwitch = newSwitch
}

host1, _ := network.AddHost("host1",
  gont.NewInterface("eth0", firstSwitch,
    opt.AddressIP("10.0.0.2/24")))
host2, _ := network.AddHost("host2",
  gont.NewInterface("eth0", lastSwitch,
    opt.AddressIP("10.0.1.2/24")))

host1.Ping(host2)
```

