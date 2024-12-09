// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"cunicu.li/gont/v2/internal/utils"
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

func NetworkCGroups() []string {
	names := []string{}

	dirs, err := os.ReadDir(filepath.Join(cgroupDir, "gont.slice"))
	if err != nil {
		return names
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		name := dir.Name()

		if strings.HasPrefix(name, "gont-") {
			name = strings.TrimPrefix(name, "gont-")
		} else {
			continue
		}

		if strings.HasSuffix(name, ".slice") {
			name = strings.TrimSuffix(name, ".slice")
		} else {
			continue
		}

		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func NetworkNames() []string {
	names := []string{}

	dirs, err := os.ReadDir(baseVarDir)
	if err != nil {
		return names
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		names = append(names, dir.Name())
	}

	sort.Strings(names)

	return names
}

func NodeNames(network string) []string {
	names := []string{}

	nodesDir := path.Join(baseVarDir, network, "nodes")

	nets, err := os.ReadDir(nodesDir)
	if err != nil {
		return names
	}

	for _, net := range nets {
		if net.IsDir() {
			names = append(names, net.Name())
		}
	}

	sort.Strings(names)

	return names
}

func GenerateNetworkName() string {
	existing := NetworkNames()

	for name := range RandomNames {
		if !slices.Contains(existing, name) {
			return name
		}
	}

	// TODO: This can generate non-unique network names!
	index := rand.Intn(len(Names)) //nolint:gosec
	random := Names[index]

	return fmt.Sprintf("%s%d", random, rand.Intn(128)+1) //nolint:gosec
}

func TeardownNetwork(ctx context.Context, c *dbus.Conn, network string) error {
	networkVarPath := filepath.Join(baseVarDir, network)
	networkTmpPath := filepath.Join(baseTmpDir, network)
	nodesVarPath := filepath.Join(networkVarPath, "nodes")

	dirs, err := os.ReadDir(nodesVarPath)
	if err != nil {
		return fmt.Errorf("failed to read nodes dir: %w", err)
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		node := dir.Name()
		if err := TeardownNode(ctx, c, network, node); err != nil {
			return fmt.Errorf("failed to teardown node '%s': %w", node, err)
		}
	}

	// Delete files
	if err := os.RemoveAll(networkVarPath); err != nil {
		return fmt.Errorf("failed to delete network dir: %w", err)
	}

	// Delete temporary files
	if err := os.RemoveAll(networkTmpPath); err != nil {
		return fmt.Errorf("failed to delete network dir: %w", err)
	}

	// Stop CGroup slice
	sliceName := fmt.Sprintf("gont-%s.slice", network)
	if _, err := c.StopUnitContext(ctx, sliceName, "fail", nil); err != nil {
		return fmt.Errorf("failed to stop cgroup: %w", err)
	}

	return nil
}

func TeardownNode(ctx context.Context, c *dbus.Conn, network, node string) error {
	nodePath := filepath.Join(baseVarDir, network, "nodes", node)
	nsMount := filepath.Join(nodePath, "ns", "net")

	// Delete network namespace mount
	if mounted, err := utils.IsMountPoint(nsMount); err == nil && mounted {
		if err := unix.Unmount(nsMount, 0); err != nil {
			return fmt.Errorf("failed to unmount netns of node '%s': %w", node, err)
		}
	} else if err != nil && !errors.Is(err, unix.ENOENT) {
		return fmt.Errorf("failed to check if mounted: %w", err)
	}

	// Delete named network namespace
	netNsName := fmt.Sprintf("gont-%s-%s", network, node)
	if err := netns.DeleteNamed(netNsName); err != nil && !errors.Is(err, unix.ENOENT) {
		return fmt.Errorf("failed to delete named network namespace: %w", err)
	}

	// Delete files
	if err := os.RemoveAll(nodePath); err != nil {
		return fmt.Errorf("failed to delete node dir: %w", err)
	}

	// Stop CGroup slice
	sliceName := fmt.Sprintf("gont-%s-%s.slice", network, node)
	if _, err := c.StopUnitContext(ctx, sliceName, "fail", nil); err != nil {
		return fmt.Errorf("failed to stop cgroup: %w", err)
	}

	return nil
}

// TeardownStaleCgroups deletes all stale CGroup slices for which no corresponding Gont network exists.
func TeardownStaleCgroups(ctx context.Context, c *dbus.Conn) ([]string, error) {
	networks := map[string]any{}
	for _, name := range NetworkNames() {
		networks[name] = nil
	}

	deleted := []string{}

	for _, name := range NetworkCGroups() {
		if _, ok := networks[name]; ok {
			continue
		}

		sliceName := fmt.Sprintf("gont-%s.slice", name)
		if _, err := c.StopUnitContext(ctx, sliceName, "fail", nil); err != nil {
			return nil, fmt.Errorf("failed to stop cgroup: %w", err)
		}

		deleted = append(deleted, name)
	}

	return deleted, nil
}
