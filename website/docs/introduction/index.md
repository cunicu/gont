---
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
sidebar_position: 1
---

# Introduction


## What does Gont do?

-   Software-defined virtual networking for testing
-   Define hosts, switches, routers in a single host
-   Reentrancy
-   Reproducibility
-   Inspired by [Mininet](http://mininet.org/)


## Mininet

> "Mininet creates a realistic virtual network, running real kernel, switch and application code, on a single machine (VM, cloud or native)"

-- [mininet.org](http://mininet.org/)

-   Written in Python 2
-   Lacking active maintainer
-   Focus on SDN: OpenFlow controllers
-   No SSL cert on homepage?!

**→** We need something better


## Why?

-   Describe network topologies quickly in code
-   Automate construction of complex topologies
-   Unit / CI testing
-   Parallel test execution
-   Example use-cases
    -   VPN / network tools development
    -   SDN Openflow controller development
    -   cunīcu: zeroconf • p2p • mesh • vpn agent
        ([cunīcu](https://github.com/cunicu/cunicu))

## Gont ...

-   can be used in Go unit & integration-tests
    -   on Github-powered CI runners
-   is licensed under Apache-2.0
-   is available at
    [github.com/cunicu/gont](https://github.com/cunicu/gont)
-   is documented at
    [pkg.go.dev/cunicu.li/gont/v2](https://pkg.go.dev/cunicu.li/gont/v2)
-   has slides at [cunicu.github.io/gont](https://cunicu.github.io/gont)
