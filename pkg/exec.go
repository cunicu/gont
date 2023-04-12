// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/stv0g/gont/v2/internal/execvpe"
	"golang.org/x/sys/unix"
)

const (
	gontNetworkSuffix  = ".gont"
	persGetPersonality = 0xffffffff // Argument to pass to personality syscall to get the current personality
	persNoRandomize    = 0x0040000  // ADDR_NO_RANDOMIZE
)

func init() {
	unshare := os.Getenv("GONT_UNSHARE")
	node := os.Getenv("GONT_NODE")
	network := os.Getenv("GONT_NETWORK")
	disableASLR := os.Getenv("GONT_DISABLE_ASLR")

	if disableASLR != "" {
		oldPers, _, err := syscall.Syscall(syscall.SYS_PERSONALITY, persGetPersonality, 0, 0)
		if err != 0 {
			panic(err)
		}

		newPers := oldPers | persNoRandomize
		if _, _, err := syscall.Syscall(syscall.SYS_PERSONALITY, newPers, 0, 0); err != syscall.Errno(0) {
			panic(err)
		}
	}

	if unshare != "" {
		// Avoid recursion
		if err := os.Unsetenv("GONT_UNSHARE"); err != nil {
			panic(err)
		}

		// Enter new namespaces
		if err := Unshare(network, node); err != nil {
			panic(err)
		}

		// Run program
		if err := execvpe.Execvpe(os.Args[0], os.Args, os.Environ()); err != nil {
			panic(err)
		}

		os.Exit(-1)
	}
}

func Exec(network, node string, args []string) error {
	if err := Unshare(network, node); err != nil {
		return err
	}

	return execvpe.Execvpe(args[0], args, os.Environ())
}

func Unshare(network, node string) error {
	networkDir := filepath.Join(baseVarDir, network)
	nodeDir := filepath.Join(networkDir, "nodes", node)

	// Setup UTS and mount namespaces
	if err := syscall.Unshare(syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS); err != nil {
		return fmt.Errorf("failed to unshare namespaces: %w", err)
	}

	// Setup node hostname
	hostname := fmt.Sprintf("%s.%s%s", node, network, gontNetworkSuffix)
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return fmt.Errorf("failed to set hostname: %w", err)
	}

	// Setup bind mounts
	if err := syscall.Mount("none", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("failed to make root mount point private: %w", err)
	}

	if err := setupBindMounts(networkDir); err != nil {
		return fmt.Errorf("failed setup network bind mounts: %w", err)
	}

	if err := setupBindMounts(nodeDir); err != nil {
		return fmt.Errorf("failed setup node bind mounts: %w", err)
	}

	// Switch network namespace
	netNsHandle := filepath.Join(nodeDir, "ns", "net")
	netNsFd, err := syscall.Open(netNsHandle, os.O_RDONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open netns: %w", err)
	}

	if err := unix.Setns(netNsFd, syscall.CLONE_NEWNET); err != nil {
		return fmt.Errorf("failed to switch to netns: %w", err)
	}

	return nil
}

func setupBindMounts(basePath string) error {
	filesRootPath := filepath.Join(basePath, "files")
	files, err := findBindMounts(filesRootPath)
	if err != nil {
		return fmt.Errorf("failed to find bindable mount points: %w", err)
	}

	// Bind mount our files and dirs into the unshared root filesystem
	for _, path := range files {
		src := filepath.Join(filesRootPath, path)
		tgt := filepath.Join("/", path)

		if err := syscall.Mount(src, tgt, "", syscall.MS_BIND, ""); err != nil {
			return fmt.Errorf("failed to mount: %w", err)
		}
	}

	return nil
}

// findBindMounts returns a slice of all files/directories which should be bind mounted.
func findBindMounts(basePath string) ([]string, error) {
	files := []string{}

	if err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		path = strings.TrimPrefix(path, basePath)

		if info.IsDir() {
			// Directories containing a hidden .mount file will be mounted
			// as a whole instead of the individual files contained in it.
			// Note: This can shadow parts of the underlying mount point.
			hfn := filepath.Join(basePath, path, ".mount")
			if fi, err := os.Stat(hfn); err == nil && !fi.IsDir() {
				files = append(files, path)
				return filepath.SkipDir
			}
		} else {
			files = append(files, path)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}
