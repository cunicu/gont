// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"testing"

	g "cunicu.li/gont/v2/pkg"
	vmo "cunicu.li/gont/v2/pkg/options/vm"
	"github.com/stretchr/testify/require"
)

const imageURL = "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-generic-arm64.qcow2" //nolint:unused

// TestPing performs and end-to-end ping test
// between two hosts on a switched topology
//
//	h1 <-> sw1 <-> h2
func TestQEmuVM(t *testing.T) {
	t.Skip()

	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.Close()

	sw1, err := n.AddSwitch("sw1")
	require.NoError(t, err, "Failed to add switch")

	vmOpts := []g.Option{
		vmo.Architecture("x86_64"),
		vmo.Machine("q35"),
		// vmo.CPU("host"),
		vmo.Memory(2 * 1024),

		vmo.NoGraphic,
		vmo.VNC(0),

		vmo.Device("virtio-blk-pci", map[string]any{"drive": "disk"}),

		vmo.Device("virtio-net-pci", map[string]any{"netdev": "net0"}),
		vmo.NetDev("tap", map[string]any{"id": "net0", "script": "no", "downscript": "no"}),
		vmo.Drive(map[string]any{"if": "none", "id": "disk", "format": "qcow2", "file": "/home/stv0g/workspace/cunicu/gont/image.qcow2"}),

		g.NewInterface("veth0", sw1),

		vmo.CloudInitMetaData{
			"instance-id":    "h1",
			"local-hostname": "h1",
		},

		vmo.CloudInitUserData{
			"ssh_pwauth": true,
			"users": []any{
				"default",
				map[string]any{
					"name": "stv0g",
					"ssh_authorized_keys": []string{
						"ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBERI62l9pAbMxi6QYd3xnEMJhOY9NxcUOvgzNrJsDqSSRs5UgRjHCTDbw+7+yqr+ibcwDAcQgnzJEdRqsdhdTdc=",
					},
					"ssh_import_id": []string{
						"gh:stv0g",
					},
					"groups":            []string{"users", "admin", "wheel"},
					"plain_text_passwd": "testtest",
				},
			},
		},
	}

	h1, err := n.AddQEmuVM("h1", vmOpts...)
	require.NoError(t, err, "Failed to add VM")

	// h2, err := n.AddQEmuVM("h2", vmOpts...)
	// require.NoError(t, err, "Failed to add VM")

	cmd, err := h1.Start()
	require.NoError(t, err)

	err = cmd.Wait()
	require.NoError(t, err)
}
