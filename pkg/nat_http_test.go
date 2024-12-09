// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

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

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	co "cunicu.li/gont/v2/pkg/options/cmd"
	"github.com/stretchr/testify/require"
)

// TestGetMyIP performs and end-to-end ping test
// through a NAT topology using IPv6 addressing
// and checks the proper NATing by using a HTTP
// "getmyip" service
//
//	h1 <-> nat1 <-> h2
func TestGetMyIP(t *testing.T) {
	n, err := g.NewNetwork(*nname)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	server, err := AddWebServer(n, "server")
	require.NoError(t, err, "Failed to create host")

	client, err := n.AddHost("client")
	require.NoError(t, err, "Failed to create host")

	nat, err := n.AddNAT("n1")
	require.NoError(t, err, "Failed to create NAT")

	err = n.AddLink(
		g.NewInterface("veth0", client,
			o.AddressIP("fc::1:2/112")),
		g.NewInterface("veth0", nat, o.SouthBound,
			o.AddressIP("fc::1:1/112")))
	require.NoError(t, err, "Failed to add link")

	err = n.AddLink(
		g.NewInterface("veth0", server,
			o.AddressIP("fc::2:2/112")),
		g.NewInterface("veth1", nat, o.NorthBound,
			o.AddressIP("fc::2:1/112")))
	require.NoError(t, err, "Failed to add link")

	err = client.AddDefaultRoute(net.ParseIP("fc::1:1"))
	require.NoError(t, err, "Failed to setup default route")

	outp := &bytes.Buffer{}
	_, err = client.Run("curl", "--silent", "--insecure", "--connect-timeout", 1000, "https://server",
		co.Stdout(outp))
	require.NoError(t, err, "Request failed")

	ip, _, err := net.SplitHostPort(outp.String())
	require.NoError(t, err, "Failed to split host:port")

	require.Equal(t, ip, "fc::2:1", "Got wrong IP")
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
		return nil, fmt.Errorf("failed to create x509 key pair: %w", err)
	}

	s := &HTTPServer{
		Host: h,
	}
	s.Server = http.Server{
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256}, // force PFS
			MinVersion:   tls.VersionTLS13,
		},
		ReadHeaderTimeout: time.Second,
		Handler:           s,
	}

	listener, err := s.ListenTCP(443)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	go s.ServeTLS(listener, "", "") //nolint:errcheck

	return s, nil
}

func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// We reply to all requests with the IP of the requester
	w.Write([]byte(req.RemoteAddr)) //nolint:errcheck
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
