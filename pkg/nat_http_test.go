package gont_test

import (
	"net"
	"net/http"
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

// TestGetMyIP performs and end-to-end ping test
// through a NAT topology using IPv6 addressing
// and checks the proper NATing by using a HTTP
// "getmyip" service
//
//  h1 <-> nat1 <-> h2
func TestGetMyIP(t *testing.T) {
	var (
		err    error
		n      *g.Network
		server *HTTPServer
		client *g.Host
		nat    *g.NAT
	)

	if n, err = g.NewNetwork(nname, opts...); err != nil {
		t.Errorf("Failed to create network: %s", err)
		t.FailNow()
	}
	defer n.Close()

	if server, err = AddWebServer(n, "server"); err != nil {
		t.Errorf("Failed to create host: %s", err)
		t.FailNow()
	}

	if client, err = n.AddHost("client"); err != nil {
		t.Errorf("Failed to create host: %s", err)
		t.FailNow()
	}

	if nat, err = n.AddNAT("n1"); err != nil {
		t.Errorf("Failed to create nat: %s", err)
		t.FailNow()
	}

	if err := n.AddLink(
		o.Interface("veth0", client,
			o.AddressIP("fc::1:2/112")),
		o.Interface("veth0", nat, o.SouthBound,
			o.AddressIP("fc::1:1/112")),
	); err != nil {
		t.Fail()
	}

	if err := n.AddLink(
		o.Interface("veth0", server,
			o.AddressIP("fc::2:2/112")),
		o.Interface("veth1", nat, o.NorthBound,
			o.AddressIP("fc::2:1/112")),
	); err != nil {
		t.Fail()
	}

	if err := client.AddDefaultRoute(net.ParseIP("fc::1:1")); err != nil {
		t.Fail()
	}

	out, _, err := client.Run("curl", "-s", "--connect-timeout", 1000, "http://server:8080")
	if err != nil {
		t.Errorf("Request failed: %s", err)
	}

	hostPort := string(out)

	ip, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		t.Errorf("Failed to split host:port: %s", err)
	}

	if ip != "fc::2:1" {
		t.Fail()
	}
}

type HTTPServer struct {
	g.Host

	listener net.Listener
}

func AddWebServer(n *g.Network, name string) (*HTTPServer, error) {
	h, err := n.AddHost(name)
	if err != nil {
		return nil, err
	}

	s := &HTTPServer{
		Host: *h,
	}

	if err := s.ListenTCP(8080); err != nil {
		return nil, err
	}

	go http.Serve(s.listener, s)

	return s, nil
}

func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// We reply to all requests with the IP of the requester
	w.Write([]byte(req.RemoteAddr))
}

func (h *HTTPServer) ListenTCP(port int) error {
	laddr := &net.TCPAddr{
		Port: port,
	}

	return h.RunFunc(func() error {
		var err error
		h.listener, err = net.ListenTCP("tcp", laddr)
		return err
	})
}

// type HTTPClient struct {
// 	g.Host

// 	Client *http.Client
// }

// func AddHttpClient(n *g.Network, name string) (*HTTPClient, error) {
// 	h, err := n.AddHost(name)
// 	if err != nil {
// 		return nil, err
// 	}

// 	s := &HTTPClient{
// 		Host: *h,
// 	}

// 	s.Client = &http.Client{
// 		Transport: &http.Transport{
// 			Dial: s.Dial,
// 		},
// 	}

// 	return s, nil
// }

// func (h *HTTPClient) Dial(network, addr string) (net.Conn, error) {
// 	return net.Dial(network)
// }
