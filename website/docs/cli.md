---
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# CLI utility

## Example

Make network persistent:

```go
network, _ := gont.NewNetwork("mynet", opt.Persistent(true))
```

Introspect network after creation with `gontc`

```shell
$ gontc list
mynet

$ gontc list mynet
mynet/host1
mynet/host2

$ gontc exec mynet/host1 hostname
host1.mynet.gont

$ gontc shell mynet/host1
$ mynet/host1: ip address show
```


## Usage

```text
Usage: gontc [flags] <command>

    Supported <commands> are:

      identify                               return the network and node name if gontc is executed within a network namespace
      shell [<net>]/<node>                   get an interactive shell inside <node>
      exec  [<net>]/<node> <command> [args]  executes a <command> in the namespace of <node> with optional [args]
      list  [<net>]                          list all active Gont networks or nodes of a given network
      clean [<net>]                          removes the all or just the specified Gont network
      help                                   show this usage information
      version                                shows the version of Gont

   Example:

      gontc exec zorn/host1 ping host2

   Gont - The Go network tester
      Author Steffen Vogel <post@steffenvogel>
```

