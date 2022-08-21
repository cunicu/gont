# Gont - A Go network testing toolkit

[![GitHub Workflow Status (master)](https://img.shields.io/github/workflow/status/stv0g/gont/Build%20and%20Test/master)](https://github.com/stv0g/gont/actions)
[![Codacy grade](https://img.shields.io/codacy/grade/d6da26516eae43b7b9ef23c5f24c35a6)](https://app.codacy.com/gh/stv0g/gont/)
[![Codacy coverage](https://img.shields.io/codacy/coverage/d6da26516eae43b7b9ef23c5f24c35a6)](https://app.codacy.com/gh/stv0g/gont/)
[![libraries.io](https://img.shields.io/librariesio/github/stv0g/gont)](https://libraries.io/github/stv0g/gont)
[![DOI](https://zenodo.org/badge/413409974.svg)](https://zenodo.org/badge/latestdoi/413409974)
[![License](https://img.shields.io/github/license/stv0g/gont)](https://github.com/stv0g/gont/blob/master/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/stv0g/gont)
[![Go Reference](https://pkg.go.dev/badge/github.com/stv0g/gont.svg)](https://pkg.go.dev/github.com/stv0g/gont)

Gont is a package to create realistic virtual network, running real kernel, switch and application code, on a single machine (VM, cloud or native).

Gont is heavily inspired by [Mininet](https://mininet.org).
It allows the user to build virtual network topologies using Go code.
Under to hood the network is then constructed using Linux virtual bridges and network namespaces.

## Getting started

Have a look at our **[slide set](https://stv0g.github.io/gont/)** to get you started.

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
    - Automatic decryption of captured trafic using Wireshark/thark by including session secrets in PCAPng file
    - Automatic instrumentation of sub-processes using [`SSLKEYLOGFILE` environment variable](https://everything.curl.dev/usingcurl/tls/sslkeylogfile)

## Examples

Have a look at the unit tests for usage examples:

-   [Simple](pkg/simple_test.go)
-   [Run](pkg/node_test.go)
-   [NAT](pkg/nat_test.go)
-   [Switch](pkg/switch_test.go)
-   [Links](pkg/link_test.go)
-   [Firewall Rules](pkg/filter_test.go)
-   [Packet tracing](pkg/capture_test.go)

## Prerequisites

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

[erigrid]: https://erigrid2.eu
[gont]: https://github.com/stv0g/gont
