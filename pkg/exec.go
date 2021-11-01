package gont

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/stv0g/gont/internal/execvpe"
	"github.com/stv0g/gont/internal/utils"
	"golang.org/x/sys/unix"
)

const (
	gontNetworkSuffix = ".gont"
)

func init() {
	unshare := os.Getenv("GONT_UNSHARE")
	node := os.Getenv("GONT_NODE")
	network := os.Getenv("GONT_NETWORK")

	if unshare != "" {
		SetupLogging()

		// Avoid recursion
		if err := os.Unsetenv("GONT_UNSHARE"); err != nil {
			panic(err)
		}

		if err := Exec(network, node, os.Args); err != nil {
			panic(err)
		}

		os.Exit(-1)
	}
}

func Exec(network, node string, args []string) error {
	basePath := filepath.Join(varDir, network)
	nodeDir := filepath.Join(basePath, "nodes", node)

	// Setup UTS and mount namespaces
	if err := syscall.Unshare(syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS); err != nil {
		panic(err)
	}

	hostname := fmt.Sprintf("%s.%s%s", node, network, gontNetworkSuffix)
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		panic(err)
	}

	if err := syscall.Mount("none", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		panic(err)
	}

	filesRootPath := filepath.Join(basePath, "files")

	files, err := utils.FindFiles(filesRootPath)
	if err != nil {
		panic(err)
	}

	if err := syscall.Unshare(syscall.CLONE_NEWNS); err != nil {
		panic(err)
	}

	// Bind mount our files into the unshared rootfs
	for _, path := range files {
		src := filepath.Join(filesRootPath, path)
		tgt := filepath.Join("/", path)
		if err := syscall.Mount(src, tgt, "", syscall.MS_BIND, ""); err != nil {
			return err
		}
	}

	// Switch network namespace
	netNsHandle := filepath.Join(nodeDir, "ns", "net")
	netNsFd, err := syscall.Open(netNsHandle, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	if err := unix.Setns(netNsFd, syscall.CLONE_NEWNET); err != nil {
		panic(err)
	}

	// Run program
	if err := execvpe.Execvpe(args[0], args, os.Environ()); err != nil {
		panic(err)
	}

	return nil
}
