# Gont - A testing framework for distributed Go applications

[![GitHub Workflow Status (main)](https://img.shields.io/github/actions/workflow/status/cunicu/gont/build.yaml)](https://github.com/cunicu/gont/actions)
[![Codacy grade](https://img.shields.io/codacy/grade/d6da26516eae43b7b9ef23c5f24c35a6)](https://app.codacy.com/gh/cunicu/gont/)
[![Codecov branch](https://img.shields.io/codecov/c/github/cunicu/gont/main?style=flat-square&token=2QHPZ691UD)](https://app.codecov.io/gh/cunicu/gont/tree/main)
[![libraries.io](https://img.shields.io/librariesio/github/cunicu/gont)](https://libraries.io/github/cunicu/gont)
[![DOI](https://zenodo.org/badge/413409974.svg)](https://zenodo.org/badge/latestdoi/413409974)
[![License](https://img.shields.io/github/license/cunicu/gont)](https://github.com/cunicu/gont/blob/main/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/cunicu/gont)
[![Go Reference](https://pkg.go.dev/badge/github.com/cunicu/gont.svg)](https://pkg.go.dev/github.com/cunicu/gont/v2)

Gont is a Go package to support the development networked and distributed applications.

It can construct a virtual network using Linux network spaces, to simulate switches, routers, NAT and endpoints, on a single machine (VM, cloud or native).
In addition, it supports developers with tooling for tracing and debugger tooling for investigating distributed applications.

Gont is heavily inspired by [Mininet](https://mininet.org).
It allows the user to build virtual network topologies defined Go code.
Under the hood the network is then constructed using Linux virtual bridges and network namespaces.

Gont runs on all moderatly recent Linux versions and requires `NET_ADMIN` capabilities (or root access).

Using Gont, developers can test complex distributed peer-to-peer and federated applications like routing daemons or social networks and messaging.
Integration tests can be automated and executed in CI environments like GitHub actions (which are in fact used to test Gont itself).

## Getting started

Have a look at our **[slide set](https://cunicu.github.io/gont/)** to get you started.

## Features

-   Various common network nodes:
    -   Standard host
    -   Layer-3 Routers
    -   Layer-2 Switches
    -   Layer-3 NAT Routers
    -   Layer-3 NAT to host networks

-   Hostname resolution for test nodes (/etc/hosts overlay)
-   Execution of sub-processes, Go code & functions in the network namespace of test nodes
-   Simultaneous setup of multiple isolated networks
-   Ideal for Golang unit tests
-   Can run in workflows powered by GitHub's runners
-   Lean code thanks to [functional options](https://sagikazarmark.hu/blog/functional-options-on-steroids/)
-   Full IPv6 support
-   Per link network emulation and bandwidth limiting via for [Netem](https://man7.org/linux/man-pages/man8/tc-netem.8.html) and [TBF](https://man7.org/linux/man-pages/man8/tc-tbf.8.html) queuing disciplines
-   Use of existing network namespaces as nodes
-   Configuration of per-host nftables firewall rules
-   Built-in Ping & Traceroute diagnostic tools
-   Built-in packet tracing with [PCAPng](https://wiki.wireshark.org/Development/PcapNg) output
    - Real-time streaming of PCAPng data to WireShark via [TCP sockets or named-pipes](https://wiki.wireshark.org/CaptureSetup/Pipes.md)
    - Automatic decryption of captured trafic using Wireshark/thark by including session secrets in PCAPng file
    - Automatic instrumentation of sub-processes using [`SSLKEYLOGFILE` environment variable](https://everything.curl.dev/usingcurl/tls/sslkeylogfile)
- Distributed tracing of events
  - A `slog.Handler` to emit [structured log](https://pkg.go.dev/log/slog) records as trace events
  - A `zapcore.Core` to emit [zap](https://github.com/uber-go/zap) log messages as trace events
  - Dedicated [gont/trace](https://pkg.go.dev/github.com/cunicu/gont/v2/pkg/trace) package for emitting trace events
  - Capturing of trace events in PCAPng files
  - WireShark Lua dissector for decoding events
- Built-in [Delve](https://github.com/go-delve/delve) debugger
  - Simultaneous attachment to multiple processes
  - Tracing via HW watch- & breakpoints to emit tracer events (see above)
    - Capture and investigate tracepoints in WireShark
  - Remote debugging via [DAP](https://microsoft.github.io/debug-adapter-protocol/)
  - Generation of VS Code [compound launch configurations](https://code.visualstudio.com/docs/editor/debugging#_compound-launch-configurations)
    - Start Gont test and attach to all processes at once

## Examples

Have a look at the unit tests for usage examples:

-   [Ping](pkg/ping_test.go)
-   [Run](pkg/run_test.go)
-   [NAT](pkg/nat_test.go)
-   [Switch](pkg/switch_test.go)
-   [Links](pkg/link_test.go)
-   [Firewall Rules](pkg/filter_test.go)
-   [Packet tracing](pkg/capture_test.go)
    - [With TLS decryption](pkg/capture_keylog_test.go)
-   [Event tracing](pkg/trace_test.go)
-   [Debugging](pkg/debug_test.go)

## Prerequisites

-   Go version 1.19 or later
-   `traceroute` (for testing)
-   `libpcap` (for compiling BPF filter expressions of packet tracing feature)

## Architecture

[![](https://mermaid.ink/img/eyJjb2RlIjoiY2xhc3NEaWFncmFtXG4gICAgZGlyZWN0aW9uIEJUXG5cbiAgICBjbGFzcyBOZXR3b3JrIHtcbiAgICAgICAgTm9kZXMgW11Ob2RlXG4gICAgICAgIExpbmtzIFtdTGlua1xuICAgIH1cblxuICAgIGNsYXNzIExpbmsge1xuICAgICAgICBMZWZ0IEVuZHBvaW50XG4gICAgICAgIFJpZ2h0IEVuZHBvaW50XG4gICAgfVxuXG4gICAgY2xhc3MgSW50ZXJmYWNlIHtcbiAgICAgICAgTmFtZSBzdHJpbmdcbiAgICAgICAgTm9kZSBOb2RlXG5cbiAgICAgICAgQWRkcmVzc2VzIFtdbmV0LklQTmV0XG4gICAgfVxuXG4gICAgY2xhc3MgTmFtZXNwYWNlIHtcbiAgICAgICAgTnNGZCBpbnRcbiAgICAgICAgUnVuKClcbiAgICB9XG5cbiAgICBjbGFzcyBOb2RlIHtcbiAgICAgICAgTmFtZSBzdHJpbmdcbiAgICB9XG5cbiAgICBjbGFzcyBIb3N0IHtcbiAgICAgICAgSW50ZXJmYWNlcyBbXUludGVyZmFjZVxuICAgICAgICBBZGRJbnRlcmZhY2UoKVxuICAgIH1cblxuICAgIGNsYXNzIFN3aXRjaCB7XG4gICAgICAgIFBvcnRzIFtdUG9ydFxuICAgICAgICBBZGRQb3J0KClcbiAgICB9XG5cbiAgICBjbGFzcyBSb3V0ZXIge1xuICAgICAgICBBZGRSb3V0ZSgpXG4gICAgfVxuXG4gICAgY2xhc3MgTkFUIHtcblxuICAgIH1cbiAgICAgICAgICAgIFxuICAgIE5vZGUgKi0tIE5hbWVzcGFjZVxuICAgIEhvc3QgKi0tIE5vZGVcbiAgICBSb3V0ZXIgKi0tIEhvc3RcbiAgICBOQVQgKi0tIFJvdXRlclxuICAgIFN3aXRjaCAqLS0gTm9kZVxuXG4gICAgSW50ZXJmYWNlIFwiMVwiIG8tLSBcIjFcIiBOb2RlXG5cblxuICAgIExpbmsgXCIxXCIgby0tIFwiMlwiIEludGVyZmFjZVxuXG4gICAgTmV0d29yayBcIjFcIiBvLS0gXCIqXCIgTGlua1xuICAgIE5ldHdvcmsgXCIxXCIgby0tIFwiKlwiIE5vZGUiLCJtZXJtYWlkIjp7InRoZW1lIjoiZGVmYXVsdCJ9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlLCJhdXRvU3luYyI6dHJ1ZSwidXBkYXRlRGlhZ3JhbSI6ZmFsc2V9)](https://mermaid.live/edit/#eyJjb2RlIjoiY2xhc3NEaWFncmFtXG4gICAgZGlyZWN0aW9uIEJUXG5cbiAgICBjbGFzcyBOZXR3b3JrIHtcbiAgICAgICAgTm9kZXMgW11Ob2RlXG4gICAgICAgIExpbmtzIFtdTGlua1xuICAgIH1cblxuICAgIGNsYXNzIExpbmsge1xuICAgICAgICBMZWZ0IEVuZHBvaW50XG4gICAgICAgIFJpZ2h0IEVuZHBvaW50XG4gICAgfVxuXG4gICAgY2xhc3MgSW50ZXJmYWNlIHtcbiAgICAgICAgTmFtZSBzdHJpbmdcbiAgICAgICAgTm9kZSBOb2RlXG5cbiAgICAgICAgQWRkcmVzc2VzIFtdbmV0LklQTmV0XG4gICAgfVxuXG4gICAgY2xhc3MgTmFtZXNwYWNlIHtcbiAgICAgICAgTnNGZCBpbnRcbiAgICAgICAgUnVuKClcbiAgICB9XG5cbiAgICBjbGFzcyBOb2RlIHtcbiAgICAgICAgTmFtZSBzdHJpbmdcbiAgICB9XG5cbiAgICBjbGFzcyBIb3N0IHtcbiAgICAgICAgSW50ZXJmYWNlcyBbXUludGVyZmFjZVxuICAgICAgICBBZGRJbnRlcmZhY2UoKVxuICAgIH1cblxuICAgIGNsYXNzIFN3aXRjaCB7XG4gICAgICAgIFBvcnRzIFtdUG9ydFxuICAgICAgICBBZGRQb3J0KClcbiAgICB9XG5cbiAgICBjbGFzcyBSb3V0ZXIge1xuICAgICAgICBBZGRSb3V0ZSgpXG4gICAgfVxuXG4gICAgY2xhc3MgTkFUIHtcblxuICAgIH1cbiAgICAgICAgICAgIFxuICAgIE5vZGUgKi0tIE5hbWVzcGFjZVxuICAgIEhvc3QgKi0tIE5vZGVcbiAgICBSb3V0ZXIgKi0tIEhvc3RcbiAgICBOQVQgKi0tIFJvdXRlclxuICAgIFN3aXRjaCAqLS0gTm9kZVxuXG4gICAgSW50ZXJmYWNlIFwiMVwiIG8tLSBcIjFcIiBOb2RlXG5cblxuICAgIExpbmsgXCIxXCIgby0tIFwiMlwiIEludGVyZmFjZVxuXG4gICAgTmV0d29yayBcIjFcIiBvLS0gXCIqXCIgTGlua1xuICAgIE5ldHdvcmsgXCIxXCIgby0tIFwiKlwiIE5vZGUiLCJtZXJtYWlkIjoie1xuICBcInRoZW1lXCI6IFwiZGVmYXVsdFwiXG59IiwidXBkYXRlRWRpdG9yIjpmYWxzZSwiYXV0b1N5bmMiOnRydWUsInVwZGF0ZURpYWdyYW0iOmZhbHNlfQ)

<!-- 
```mermaid
classDiagram
    direction BT

    class Network {
        Nodes []Node
        Links []Link
    }

    class Link {
        Left Endpoint
        Right Endpoint
    }

    class Interface {
        Name string
        Node Node

        Addresses []net.IPNet
    }

    class Namespace {
        NsFd int
        Run()
    }

    class Node {
        Name string
    }

    class Host {
        Interfaces []Interface
        AddInterface()
    }

    class Switch {
        Ports []Port
        AddPort()
    }

    class Router {
        AddRoute()
    }

    class NAT {

    }
            
    Node *-- Namespace
    Host *-- Node
    Router *-- Host
    NAT *-- Router
    Switch *-- Node

    Interface "1" o-- "1" Node


    Link "1" o-- "2" Interface

    Network "1" o-- "*" Link
    Network "1" o-- "*" Node
``` -->

## Credits

-   [Steffen Vogel](https://github.com/stv0g) [📧](mailto:post@steffenvogel.de)

### Funding acknowledment

<img alt="European Flag" src="https://erigrid2.eu/wp-content/uploads/2020/03/europa_flag_low.jpg" align="left" style="margin-right: 10px"/> The development of [Gont][gont] has been supported by the [ERIGrid 2.0][erigrid] project \
of the H2020 Programme under [Grant Agreement No. 870620](https://cordis.europa.eu/project/id/870620)

## License

Gont is [REUSE compliant](https://reuse.software/) and mainly licensed under the [Apache 2.0 license](LICENSES/Apache-2.0.txt)

- SPDX-FileCopyrightText: 2023 Steffen Vogel \<post@steffenvogel.de\>
- SPDX-License-Identifier: Apache-2.0

[erigrid]: https://erigrid2.eu
[gont]: https://github.com/cunicu/gont
