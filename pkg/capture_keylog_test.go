// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

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

	"github.com/stretchr/testify/assert"
	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
	co "github.com/stv0g/gont/pkg/options/capture"
)

// TestCaptureKeyLog tests the decryption of captured traffic
func TestCaptureKeyLog(t *testing.T) {
	tmpPCAP, err := os.CreateTemp(t.TempDir(), "gont-capture-*.pcapng")
	assert.NoError(t, err, "Failed to open temporary file")

	c1 := g.NewCapture(
		co.ToFile(tmpPCAP),
		co.LogKeys(true),
		co.Comment("This PCAPng file contains TLS decryption secrets"),
	)

	n, err := g.NewNetwork(*nname,
		o.Customize[g.NetworkOption](globalNetworkOptions, c1, // Also multiple capturers are supported
			g.NewCapture(
				co.ToFilename("all.pcapng")), // We can create a file
		)...)
	assert.NoError(t, err, "Failed to create network")

	server, err := AddWebServer(n, "server")
	assert.NoError(t, err, "Failed to create host")

	client, err := n.AddHost("client")
	assert.NoError(t, err, "Failed to create host")

	err = n.AddLink(
		g.NewInterface("veth0", client,
			o.AddressIP("fc::1:2/112")),
		g.NewInterface("veth0", server,
			o.AddressIP("fc::1:1/112")))
	assert.NoError(t, err, "Failed to add link")

	_, err = client.Run("curl", "--http2", "--silent", "--insecure", "--connect-timeout", 5, "https://server")
	assert.NoError(t, err, "cURL Request failed: %s")

	// Wait until all traffic propagates through PCAP
	time.Sleep(time.Second)

	// We must close here so all decryption secrets are written to the PCAP files
	err = n.Close()
	assert.NoError(t, err, "Failed to close network")

	t.Logf("PCAPng file: %s", tmpPCAP.Name())

	c := exec.Command("tshark", "-r", tmpPCAP.Name(), "-T", "fields", "-e", "http2.data.data", "len(http2.data.data) > 0") //nolint:gosec

	out := &bytes.Buffer{}
	c.Stdout = out

	err = c.Run()
	assert.NoError(t, err, "Failed to run tshark")

	hostPortBytes, err := hex.DecodeString(strings.TrimSpace(out.String()))
	assert.NoError(t, err, "Failed to decode HTTP response body")

	hostPort := string(hostPortBytes)
	ip, _, err := net.SplitHostPort(hostPort)
	assert.NoError(t, err, "Failed to split host:port")

	assert.Equal(t, ip, "fc::1:2", "Got wrong IP")
}
