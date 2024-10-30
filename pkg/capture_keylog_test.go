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

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	co "cunicu.li/gont/v2/pkg/options/capture"
	"github.com/stretchr/testify/require"
)

// TestCaptureKeyLog tests the decryption of captured traffic
func TestCaptureKeyLog(t *testing.T) {
	tmpPCAP, err := os.CreateTemp(t.TempDir(), "gont-capture-*.pcapng")
	require.NoError(t, err, "Failed to open temporary file")

	c1 := g.NewCapture(
		co.ToFile(tmpPCAP),
		co.LogKeys(true),
		co.Comment("This PCAPng file contains TLS decryption secrets"),
	)

	n, err := g.NewNetwork(*nname,
		g.Customize(globalNetworkOptions, c1, // Also multiple capturers are supported
			g.NewCapture(
				co.ToFilename("all.pcapng")), // We can create a file
		)...)
	require.NoError(t, err, "Failed to create network")

	server, err := AddWebServer(n, "server")
	require.NoError(t, err, "Failed to create host")

	client, err := n.AddHost("client")
	require.NoError(t, err, "Failed to create host")

	err = n.AddLink(
		g.NewInterface("veth0", client,
			o.AddressIP("fc::1:2/112")),
		g.NewInterface("veth0", server,
			o.AddressIP("fc::1:1/112")))
	require.NoError(t, err, "Failed to add link")

	_, err = client.Run("curl", "--http2", "--silent", "--insecure", "--connect-timeout", 5, "https://server")
	require.NoError(t, err, "cURL Request failed: %s", err)

	// Wait until all traffic propagates through PCAP
	time.Sleep(time.Second)

	// We must close here so all decryption secrets are written to the PCAP files
	err = n.Close()
	require.NoError(t, err, "Failed to close network")

	t.Logf("PCAPng file: %s", tmpPCAP.Name())

	c := exec.Command("tshark", "-r", tmpPCAP.Name(), "-T", "fields", "-e", "http2.data.data", "len(http2.data.data) > 0") //nolint:gosec

	out := &bytes.Buffer{}
	c.Stdout = out

	err = c.Run()
	require.NoError(t, err, "Failed to run tshark")

	hostPortBytes, err := hex.DecodeString(strings.TrimSpace(out.String()))
	require.NoError(t, err, "Failed to decode HTTP response body")

	hostPort := string(hostPortBytes)
	ip, _, err := net.SplitHostPort(hostPort)
	require.NoError(t, err, "Failed to split host:port")

	require.Equal(t, ip, "fc::1:2", "Got wrong IP")
}
