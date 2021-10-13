# Gont - A Go network testing toolkit

[![Go Reference](https://pkg.go.dev/badge/github.com/stv0g/gont.svg)](https://pkg.go.dev/github.com/stv0g/gont)
![Snyk.io](https://img.shields.io/snyk/vulnerabilities/github/stv0g/gont)
[![Build](https://img.shields.io/github/checks-status/stv0g/gont/master)](https://github.com/stv0g/gont/actions)
[![libraries.io](https://img.shields.io/librariesio/release/stv0g/gont)](https://libraries.io/github/stv0g/gont)
[![GitHub](https://img.shields.io/github/license/stv0g/gont)](https://github.com/stv0g/gont/blob/master/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/stv0g/gont)

Gont is a package to creat realistic virtual network, running real kernel, switch and application code, on a single machine (VM, cloud or native).

Gont is heavily inspired by [Mininet](https://mininet.org).
It allows the user to build virtual network topologies using Go code.
Under to hood the network is then constructed using Linux virtual bridges and network namespaces.

## Getting started

Have a look at our **[slide set](docs/index.html)** to get you started.

## Features

- L3 Routers
- L2 Switches
- L3 NAT Routers
- L3 Host NAT (to host network)
- Hostname resolution (using /etc/hosts)
- Support for multiple simultaneous and isolated networks
- Ideal for golang unit tests
- Can run in workflows powered by GitHub's runners
- Lean code thanks to [functional options](https://sagikazarmark.hu/blog/functional-options-on-steroids/)
- Full IPv6 support
- Use of existing network namespaces as nodes

## Examples

Have a look at the unit tests for usage examples:

- [Simple](pkg/simple_test.go)
- [Run](pkg/node_test.go)
- [NAT](pkg/nat_test.go)
- [Switch](pkg/switch_test.go)
- [Links](pkg/link_test.go)

## Prerequisites

- `iptables` (for NAT)
- `ping` (for testing)
- `traceroute` (for testing)

## Roadmap

- NAT
  - Use netlink socket instead of `iptables` tool for configuring NAT
- Integrate go imlementations of `ping` and `traceroute` tools
- More tests
    - Especially for host NAT and multiple networks
- Add support for netem and tbf qdiscs on Links
- Add separate examples directory
- Topology factories
- Add support for using existing Docker containers as nodes

## Architecture

[![Architecture](https://mermaid.ink/img/eyJjb2RlIjoiY2xhc3NEaWFncmFtXG4gICAgZGlyZWN0aW9uIFJMXG5cbiAgICBjbGFzcyBOZXR3b3JrIHtcbiAgICAgICAgTm9kZXMgW11Ob2RlXG4gICAgICAgIExpbmtzIFtdTGlua1xuICAgIH1cblxuICAgIGNsYXNzIExpbmsge1xuICAgICAgICBMZWZ0IEVuZHBvaW50XG4gICAgICAgIFJpZ2h0IEVuZHBvaW50XG4gICAgfVxuXG4gICAgY2xhc3MgRW5kcG9pbnQge1xuICAgIH1cblxuICAgIGNsYXNzIFBvcnQge1xuICAgICAgICBOYW1lIHN0cmluZ1xuICAgICAgICBOb2RlIE5vZGVcbiAgICB9XG5cbiAgICBjbGFzcyBJbnRlcmZhY2Uge1xuICAgICAgICBBZGRyZXNzZXMgW11uZXQuSVBOZXRcbiAgICB9XG5cbiAgICBjbGFzcyBOYW1lc3BhY2Uge1xuICAgICAgICBOc0ZkIGludFxuICAgICAgICBSdW4oKVxuICAgIH1cblxuICAgIGNsYXNzIE5vZGUge1xuICAgICAgICBOYW1lIHN0cmluZ1xuICAgIH1cblxuICAgIGNsYXNzIEhvc3Qge1xuICAgICAgICBJbnRlcmZhY2VzIFtdSW50ZXJmYWNlXG4gICAgICAgIEFkZEludGVyZmFjZSgpXG4gICAgfVxuXG4gICAgY2xhc3MgU3dpdGNoIHtcbiAgICAgICAgUG9ydHMgW11Qb3J0XG4gICAgICAgIEFkZFBvcnQoKVxuICAgIH1cblxuICAgIGNsYXNzIFJvdXRlciB7XG4gICAgICAgIEFkZFJvdXRlKClcbiAgICB9XG5cbiAgICBjbGFzcyBOQVQge1xuXG4gICAgfVxuICAgICAgICAgICAgXG4gICAgTm9kZSAqLS0gTmFtZXNwYWNlXG4gICAgSG9zdCAqLS0gTm9kZVxuICAgIFJvdXRlciAqLS0gSG9zdFxuICAgIE5BVCAqLS0gUm91dGVyXG4gICAgU3dpdGNoICotLSBOb2RlXG5cbiAgICBQb3J0ICotLSBFbmRwb2ludFxuICAgIEludGVyZmFjZSAqLS0gUG9ydFxuXG4gICAgUG9ydCBcIjFcIiBvLS0gXCIxXCIgTm9kZVxuXG4gICAgSG9zdCBcIjFcIiBvLS0gXCIqXCIgSW50ZXJmYWNlXG4gICAgU3dpdGNoIFwiMVwiIG8tLSBcIipcIiBQb3J0XG5cbiAgICBMaW5rIFwiMVwiIG8tLSBcIjJcIiBFbmRwb2ludFxuXG4gICAgTmV0d29yayBcIjFcIiBvLS0gXCIqXCIgTGlua1xuICAgIE5ldHdvcmsgXCIxXCIgby0tIFwiKlwiIE5vZGUiLCJtZXJtYWlkIjp7InRoZW1lIjoiZGVmYXVsdCJ9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlLCJhdXRvU3luYyI6dHJ1ZSwidXBkYXRlRGlhZ3JhbSI6ZmFsc2V9)](https://mermaid.live/edit#eyJjb2RlIjoiY2xhc3NEaWFncmFtXG4gICAgZGlyZWN0aW9uIFJMXG5cbiAgICBjbGFzcyBOZXR3b3JrIHtcbiAgICAgICAgTm9kZXMgW11Ob2RlXG4gICAgICAgIExpbmtzIFtdTGlua1xuICAgIH1cblxuICAgIGNsYXNzIExpbmsge1xuICAgICAgICBMZWZ0IEVuZHBvaW50XG4gICAgICAgIFJpZ2h0IEVuZHBvaW50XG4gICAgfVxuXG4gICAgY2xhc3MgRW5kcG9pbnQge1xuICAgIH1cblxuICAgIGNsYXNzIFBvcnQge1xuICAgICAgICBOYW1lIHN0cmluZ1xuICAgICAgICBOb2RlIE5vZGVcbiAgICB9XG5cbiAgICBjbGFzcyBJbnRlcmZhY2Uge1xuICAgICAgICBBZGRyZXNzZXMgW11uZXQuSVBOZXRcbiAgICB9XG5cbiAgICBjbGFzcyBOYW1lc3BhY2Uge1xuICAgICAgICBOc0ZkIGludFxuICAgICAgICBSdW4oKVxuICAgIH1cblxuICAgIGNsYXNzIE5vZGUge1xuICAgICAgICBOYW1lIHN0cmluZ1xuICAgIH1cblxuICAgIGNsYXNzIEhvc3Qge1xuICAgICAgICBJbnRlcmZhY2VzIFtdSW50ZXJmYWNlXG4gICAgICAgIEFkZEludGVyZmFjZSgpXG4gICAgfVxuXG4gICAgY2xhc3MgU3dpdGNoIHtcbiAgICAgICAgUG9ydHMgW11Qb3J0XG4gICAgICAgIEFkZFBvcnQoKVxuICAgIH1cblxuICAgIGNsYXNzIFJvdXRlciB7XG4gICAgICAgIEFkZFJvdXRlKClcbiAgICB9XG5cbiAgICBjbGFzcyBOQVQge1xuXG4gICAgfVxuICAgICAgICAgICAgXG4gICAgTm9kZSAqLS0gTmFtZXNwYWNlXG4gICAgSG9zdCAqLS0gTm9kZVxuICAgIFJvdXRlciAqLS0gSG9zdFxuICAgIE5BVCAqLS0gUm91dGVyXG4gICAgU3dpdGNoICotLSBOb2RlXG5cbiAgICBQb3J0ICotLSBFbmRwb2ludFxuICAgIEludGVyZmFjZSAqLS0gUG9ydFxuXG4gICAgUG9ydCBcIjFcIiBvLS0gXCIxXCIgTm9kZVxuXG4gICAgSG9zdCBcIjFcIiBvLS0gXCIqXCIgSW50ZXJmYWNlXG4gICAgU3dpdGNoIFwiMVwiIG8tLSBcIipcIiBQb3J0XG5cbiAgICBMaW5rIFwiMVwiIG8tLSBcIjJcIiBFbmRwb2ludFxuXG4gICAgTmV0d29yayBcIjFcIiBvLS0gXCIqXCIgTGlua1xuICAgIE5ldHdvcmsgXCIxXCIgby0tIFwiKlwiIE5vZGUiLCJtZXJtYWlkIjoie1xuICBcInRoZW1lXCI6IFwiZGVmYXVsdFwiXG59IiwidXBkYXRlRWRpdG9yIjpmYWxzZSwiYXV0b1N5bmMiOnRydWUsInVwZGF0ZURpYWdyYW0iOmZhbHNlfQ)
<!-- 
```mermaid
classDiagram
    direction RL

    class Network {
        Nodes []Node
        Links []Link
    }

    class Link {
        Left Endpoint
        Right Endpoint
    }

    class Endpoint {
    }

    class Port {
        Name string
        Node Node
    }

    class Interface {
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

    Port *-- Endpoint
    Interface *-- Port

    Port "1" o-- "1" Node

    Host "1" o-- "*" Interface
    Switch "1" o-- "*" Port

    Link "1" o-- "2" Endpoint

    Network "1" o-- "*" Link
    Network "1" o-- "*" Node
``` -->

## Credits

- [Steffen Vogel](https://github.com/stv0g) [📧](mailto:post@steffenvogel.de)

### Funding acknowledment

<img alt="European Flag" src="https://erigrid2.eu/wp-content/uploads/2020/03/europa_flag_low.jpg" align="left" style="margin-right: 10px"/> The development of [Gont] has been supported by the [ERIGrid 2.0] project \
of the H2020 Programme under [Grant Agreement No. 870620](https://cordis.europa.eu/project/id/870620)

[ERIGrid 2.0]: https://erigrid2.eu
[Gont]: https://github.com/stv0g/gont
