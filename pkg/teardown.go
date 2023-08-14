// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"sort"

	"cunicu.li/gont/v2/internal/utils"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

func NetworkNames() []string {
	names := []string{}

	nets, err := os.ReadDir(baseVarDir)
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

	for i := 0; i < 32; i++ {
		random := GetRandomName()

		index := sort.SearchStrings(existing, random)
		if index >= len(existing) || existing[index] != random {
			return random
		}
	}

	// TODO: This can generate non-unique network names!
	index := rand.Intn(len(Names)) //nolint:gosec
	random := Names[index]

	return fmt.Sprintf("%s%d", random, rand.Intn(128)+1) //nolint:gosec
}

func TeardownAllNetworks() error {
	for _, name := range NetworkNames() {
		if err := TeardownNetwork(name); err != nil {
			return fmt.Errorf("failed to teardown network '%s': %w", name, err)
		}
	}

	return nil
}

func TeardownNetwork(network string) error {
	networkVarPath := filepath.Join(baseVarDir, network)
	networkTmpPath := filepath.Join(baseTmpDir, network)
	nodesVarPath := filepath.Join(networkVarPath, "nodes")

	fis, err := os.ReadDir(nodesVarPath)
	if err != nil {
		return fmt.Errorf("failed to read nodes dir: %w", err)
	}

	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}

		node := fi.Name()
		if err := TeardownNode(network, node); err != nil {
			return fmt.Errorf("failed to teardown node '%s': %w", node, err)
		}
	}

	if err := os.RemoveAll(networkVarPath); err != nil {
		return fmt.Errorf("failed to delete network dir: %w", err)
	}

	if err := os.RemoveAll(networkTmpPath); err != nil {
		return fmt.Errorf("failed to delete network dir: %w", err)
	}

	return nil
}

func TeardownNode(network, node string) error {
	nodePath := filepath.Join(baseVarDir, network, "nodes", node)
	nsMount := filepath.Join(nodePath, "ns", "net")

	netNsName := fmt.Sprintf("gont-%s-%s", network, node)

	if mounted, err := utils.IsMountPoint(nsMount); err == nil && mounted {
		if err := unix.Unmount(nsMount, 0); err != nil {
			return fmt.Errorf("failed to unmount netns of node '%s': %w", node, err)
		}
	} else if err != nil && !errors.Is(err, unix.ENOENT) {
		return fmt.Errorf("failed to check if mounted: %w", err)
	}

	if err := netns.DeleteNamed(netNsName); err != nil && !errors.Is(err, unix.ENOENT) {
		return fmt.Errorf("failed to delete named network namespace: %w", err)
	}

	if err := os.RemoveAll(nodePath); err != nil {
		return fmt.Errorf("failed to delete node dir: %w", err)
	}

	return nil
}
