package gont_test

import (
	"bytes"
	"encoding/hex"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
	copt "github.com/stv0g/gont/pkg/options/capture"
)

// TestCaptureKeyLog tests the decryption of captured trafic
func TestCaptureKeyLog(t *testing.T) {
	var (
		err    error
		n      *g.Network
		client *g.Host
		server *HTTPServer
	)

	tmpPCAP, err := os.CreateTemp("", "gont-capture-*.pcapng")
	if err != nil {
		t.Fatalf("Failed to open temporary file: %s", err)
	}

	c1 := gopt.CaptureAll(
		copt.ToFile(tmpPCAP),
		copt.LogKeys(true),
		copt.Comment("This PCAPng file contains TLS decryption secrets"),
	)

	if n, err = g.NewNetwork(*nname,
		gopt.Customize(globalNetworkOptions, c1, // Also multiple capturers are supported
			gopt.CaptureAll(
				copt.ToFilename("all.pcapng"), // We can create a file
			),
		)...,
	); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	if server, err = AddWebServer(n, "server"); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if client, err = n.AddHost("client"); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if err := n.AddLink(
		gopt.Interface("veth0", client,
			gopt.AddressIP("fc::1:2/112")),
		gopt.Interface("veth0", server,
			gopt.AddressIP("fc::1:1/112")),
	); err != nil {
		t.Fatalf("Failed to add link: %s", err)
	}

	if _, _, err = client.Run("curl", "--http2", "--silent", "--insecure", "--connect-timeout", 5, "https://server"); err != nil {
		t.Fatalf("cURL Request failed: %s", err)
	}

	// Wait until all trafic propagates through PCAP
	time.Sleep(time.Second)

	// We must close here so all decryption secrets are written to the PCAP files
	if err := n.Close(); err != nil {
		t.Fatalf("Failed to close network: %s", err)
	}

	t.Logf("PCAPng file: %s", tmpPCAP.Name())

	c := exec.Command("tshark", "-r", tmpPCAP.Name(), "-T", "fields", "-e", "http2.data.data", "len(http2.data.data) > 0")

	out := &bytes.Buffer{}
	c.Stdout = out

	if err := c.Run(); err != nil {
		t.Fatalf("Failed to run tshark: %s", err)
	}

	hostPortBytes, err := hex.DecodeString(strings.TrimSpace(out.String()))
	if err != nil {
		t.Fatalf("Failed to decode HTTP response body: %s", err)
	}

	hostPort := string(hostPortBytes)
	ip, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		t.Fatalf("Failed to split host:port: %s", err)
	}

	if ip != "fc::1:2" {
		t.Fatalf("Got wrong IP: %s", ip)
	}
}
