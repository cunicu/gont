package gont_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

// TestGetMyIP performs and end-to-end ping test
// through a NAT topology using IPv6 addressing
// and checks the proper NATing by using a HTTP
// "getmyip" service
//
//	h1 <-> nat1 <-> h2
func TestGetMyIP(t *testing.T) {
	var (
		err    error
		n      *g.Network
		server *HTTPServer
		client *g.Host
		nat    *g.NAT
	)

	if n, err = g.NewNetwork(*nname, globalNetworkOptions...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}
	defer n.Close()

	if server, err = AddWebServer(n, "server"); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if client, err = n.AddHost("client"); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if nat, err = n.AddNAT("n1"); err != nil {
		t.Fatalf("Failed to create nat: %s", err)
	}

	if err := n.AddLink(
		o.Interface("veth0", client,
			o.AddressIP("fc::1:2/112")),
		o.Interface("veth0", nat, o.SouthBound,
			o.AddressIP("fc::1:1/112")),
	); err != nil {
		t.Fatalf("Failed to add link: %s", err)
	}

	if err := n.AddLink(
		o.Interface("veth0", server,
			o.AddressIP("fc::2:2/112")),
		o.Interface("veth1", nat, o.NorthBound,
			o.AddressIP("fc::2:1/112")),
	); err != nil {
		t.Fatalf("Failed to add link: %s", err)
	}

	if err := client.AddDefaultRoute(net.ParseIP("fc::1:1")); err != nil {
		t.Fatalf("Failed to setup default route: %s", err)
	}

	out, _, err := client.Run("curl", "-sk", "--connect-timeout", 1000, "https://server")
	if err != nil {
		t.Fatalf("Request failed: %s", err)
	}

	hostPort := string(out)

	ip, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		t.Fatalf("Failed to split host:port: %s", err)
	}

	if ip != "fc::2:1" {
		t.Fatalf("Got wrong IP: %s", ip)
	}
}

type HTTPServer struct {
	*g.Host
	http.Server
}

func AddWebServer(n *g.Network, name string) (*HTTPServer, error) {
	h, err := n.AddHost(name)
	if err != nil {
		return nil, err
	}

	pub, priv, err := GenerateKeys(name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	cert, err := tls.X509KeyPair(pub, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create x509 keypair: %w", err)
	}

	s := &HTTPServer{
		Host: h,
	}
	s.Server = http.Server{
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256}, // force PFS
		},
		Handler: s,
	}

	listener, err := s.ListenTCP(443)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	go s.ServeTLS(listener, "", "")

	return s, nil
}

func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// We reply to all requests with the IP of the requester
	w.Write([]byte(req.RemoteAddr))
}

func (h *HTTPServer) ListenTCP(port int) (net.Listener, error) {
	laddr := &net.TCPAddr{
		Port: port,
	}

	var err error
	var listener net.Listener
	return listener, h.RunFunc(func() error {
		listener, err = net.ListenTCP("tcp", laddr)
		return err
	})
}

func GenerateKeys(host string) ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Create public key
	pubBuf := new(bytes.Buffer)
	err = pem.Encode(pubBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write data to cert.pem: %w", err)
	}

	// Create private key
	privBuf := new(bytes.Buffer)
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal private key: %w", err)
	}

	err = pem.Encode(privBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write data to key.pem: %w", err)
	}

	return pubBuf.Bytes(), privBuf.Bytes(), nil
}
