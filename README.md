# Gont - A Go network testing toolkit

[![Go Reference](https://pkg.go.dev/badge/github.com/stv0g/gont.svg)](https://pkg.go.dev/github.com/stv0g/gont)
![](https://img.shields.io/snyk/vulnerabilities/github/stv0g/gont)
[![](https://img.shields.io/github/checks-status/stv0g/gont/master)](https://github.com/stv0g/gont/actions)
[![](https://img.shields.io/librariesio/release/stv0g/gont)](https://libraries.io/github/stv0g/gont)
[![GitHub](https://img.shields.io/github/license/stv0g/gont)](https://github.com/stv0g/gont/blob/master/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/stv0g/gont)

Gont is a package to creat realistic virtual network, running real kernel, switch and application code, on a single machine (VM, cloud or native).

Gont is heavily inspired by [Mininet](https://mininet.org).
It allows the user to build virtual network topologies using Go code.
Under to hood the network is then constructed using Linux virtual bridges and network namespaces.

## Features

- L3 Routers
- L2 Switches
- L3 NAT Routers
- L3 Host NAT (to external networks)
- Host name resolution (using /etc/hosts)
- Support for multiple simultaneous networks
- Ideal for golang unit tests
- Can run in GitHub powered runners / workflows

## Prerequisites

- `iproute2`
- `iptables` (for NAT)
- `ping` (for testing)
- `traceroute` (for testing)

## Roadmap

- Use netfilter-nft kernel API instead of iptables
- Use functional options pattern
- Integrate go versions of ping and traceroute
- More tests
- Fix host NAT
- Add support for netem and tbf qdiscs on Links
- Add separate examples directory

## Example

```go
// TestPing performs and end-to-end ping test
// between two hosts on a switched topology
//
//  h1 <-> sw <-> h2
func TestPing(t *testing.T) {
	n := gont.NewNetwork("ping")
	defer n.Close()

	sw, err := n.AddSwitch("sw")
	if err != nil {
		t.Fail()
	}

	h1, err := n.AddHost("h1", nil, &gont.Interface{"eth0", net.IPv4(10, 0, 0, 1), mask(), sw})
	if err != nil {
		t.Fail()
	}

	h2, err := n.AddHost("h2", nil, &gont.Interface{"eth0", net.IPv4(10, 0, 0, 2), mask(), sw})
	if err != nil {
		t.Fail()
	}

	err = h1.Ping(h2, "-c", "1")
	if err != nil {
		t.Fail()
	}
}
```

## Credits

- Steffen Vogel (@stv0g)

### Funding acknowledment

![](https://erigrid2.eu/wp-content/uploads/2020/03/europa_flag_low.jpg) The development of [Gont] has been supported by the [ERIGrid 2.0] project of the H2020 Programme under [Grant Agreement No. 870620](https://cordis.europa.eu/project/id/870620)

[Gont]: https://github.com/stv0g/gont
