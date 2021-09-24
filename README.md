# Gont - A Go network testing toolkit

Gont is a package to creat realistic virtual network, running real kernel, switch and application code, on a single machine (VM, cloud or native).

Gont is heavily inspired by [Mininet](https://mininet.org).
It allows the user to build virtual network topologies using Go code.
Under to hood the network is then constructed using Linux virtual bridges and network namespaces.

## Prerequisites

- `iproute2`
- `iptables` (for NAT)
- `ping` (for testing)
- `traceroute` (for testing)

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
