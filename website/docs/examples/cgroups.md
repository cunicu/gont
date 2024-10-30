---
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
sidebar_position: 5
---

# CGroups

Powered by [systemd](https://systemd.io/) System and Service Manager.

## Hierarchy

_Networks_, _nodes_ and _commands_ are forming a Unified Cgroup Hierarchy (v2):

```shell
$ systemd-cgls
...
└─gont.slice
  └─gont-barlow.slice
    └─gont-barlow-h1.slice
      └─gont-run-2797869.scope
        └─2797869 some-process
```

We rely on systemd's service manager to manage the hierarchy. 

## Freeze, Thaw, Kill

All processes of a Cgroup can be controlled together:

```go
cmd.Freeze()   // Suspends all processes, included forked sub-processes
cmd.Thaw()     // Resumes all processes
cmd.Teardown() // Sends a SIGKILL to all processes
```

This also works on the _host_ and _network_ levels.

By default, `Close()` will invoke `Teardown()`.

Hence, guaranteeing that no lingering processes will stick around.


## Resource Control

We can control resource quota's in each level of the hierarchy using systemd resource control properties:

```go
import sdopt "github.com/cunicu/gont/v2/options/systemd"

network, _ := gont.NewNetwork("mynet", sdopt.AllowedCPUs(0b1100))
host1, _ := network.AddHost("host1", sdopt.TasksMax(10))

cmd := host1.Command("long-running-command", sdopt.RuntimeMax(10 * time.Second))
cmd := host1.Command("memory-hungry-command", sdopt.MemoryMax(1 << 20))
```

See: [systemd.resource-control](https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html)
