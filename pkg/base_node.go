package gont

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	nl "github.com/vishvananda/netlink"
)

type BaseNode struct {
	Node

	*Namespace
	Network *Network

	name string

	ExistingNamespace string
	DockerContainer   string

	BasePath string
}

func (n *Network) AddNode(name string, opts ...Option) (*BaseNode, error) {
	var err error

	basePath := filepath.Join(n.BasePath, "nodes", name)
	for _, path := range []string{"ns"} {
		path = filepath.Join(basePath, path)
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, err
		}
	}

	node := &BaseNode{
		name:     name,
		Network:  n,
		BasePath: basePath,
	}

	log.WithField("name", name).Info("Adding new node")

	for _, opt := range opts {
		if nopt, ok := opt.(NodeOption); ok {
			nopt.Apply(node)
		}
	}

	nsName := fmt.Sprintf("gont-%s-%s", n.Name, name)
	node.Namespace, err = NewNamespace(nsName)
	if err != nil {
		return nil, err
	}

	nsMount := filepath.Join(netnsDir, nsName)
	nsSymlink := filepath.Join(basePath, "ns", "net")
	if err := os.Symlink(nsMount, nsSymlink); err != nil {
		return nil, err
	}

	n.Nodes[name] = node

	return node, nil
}

func (n *BaseNode) Base() *BaseNode {
	return n
}

func (n *BaseNode) Name() string {
	return n.name
}

func (n *BaseNode) ConfigurePort(p Port) error {
	log.WithField("intf", n.name+"/"+p.Name).Info("Activating interface")

	return n.Namespace.RunFunc(func() error {
		link, err := nl.LinkByName(p.Name)
		if err != nil {
			return err
		}

		if p.Group != Default {
			log.WithFields(log.Fields{
				"intf":  p,
				"group": p.Group,
			}).Info("Setting device group")
			if err := nl.LinkSetGroup(link, int(p.Group)); err != nil {
				return err
			}
		}

		return nl.LinkSetUp(link)
	})
}

func (n *BaseNode) Teardown() error {
	if err := n.Namespace.Close(); err != nil {
		return err
	}

	if err := os.RemoveAll(n.BasePath); err != nil {
		return err
	}

	return nil
}

func (n *BaseNode) GoRun(script string, arg ...string) ([]byte, *exec.Cmd, error) {
	tmp := filepath.Join(n.Network.BasePath, fmt.Sprintf("go-build-%d", rand.Intn(1<<16)))
	_, _, err := n.Network.HostNode.Run("go", "build", "-o", tmp, script)
	if err != nil {
		return nil, nil, err
	}

	return n.Run(tmp, arg...)
}

func (n *BaseNode) GoStart(script string, arg ...string) (io.Reader, io.Reader, *exec.Cmd, error) {
	tmp := filepath.Join(n.Network.BasePath, fmt.Sprintf("go-build-%d", rand.Intn(1<<16)))
	_, _, err := n.Network.HostNode.Run("go", "build", "-o", tmp, script)
	if err != nil {
		return nil, nil, nil, err
	}

	return n.Start(tmp, arg...)
}

func (n *BaseNode) WriteProcFS(path, value string) error {
	_, _, err := n.Run("sh", "-c", "echo "+value+" > "+path)
	return err
}

func (n *BaseNode) EnableForwarding() error {
	if err := n.WriteProcFS("/proc/sys/net/ipv4/conf/all/forwarding", "1"); err != nil {
		return err
	}

	if err := n.WriteProcFS("/proc/sys/net/ipv6/conf/all/forwarding", "1"); err != nil {
		return err
	}

	return nil
}

func (n *BaseNode) LinkAddAddr(name string, addr net.IPNet) error {
	return n.Namespace.RunFunc(func() error {
		link, err := nl.LinkByName(name)
		if err != nil {
			return err
		}

		nlAddr := &nl.Addr{
			IPNet: &addr,
		}

		log.WithField("intf", n.name+"/"+name).WithField("addr", addr.String()).Info("Adding new address to interface")
		if err := nl.AddrAdd(link, nlAddr); err != nil {
			return err
		}

		return err
	})
}

func (n *BaseNode) AddRoute(dst net.IPNet, gw net.IP) error {
	log.WithField("node", n.name).WithField("dst", dst.String()).WithField("gw", gw.String()).Info("Add route")

	return n.Namespace.RunFunc(func() error {
		return nl.RouteAdd(&nl.Route{
			Dst: &dst,
			Gw:  gw,
		})
	})
}

func (n *BaseNode) Command(name string, arg ...string) *exec.Cmd {
	// Actual namespace switching is done similar to Docker's reexec
	// in a forked version of ourself by passing all required details
	// in environment variables.

	c := exec.Command(name, arg...)

	if !n.NsHandle.Equal(n.Network.HostNode.NsHandle) {
		c.Path = "/proc/self/exe"
		c.Env = append(os.Environ(),
			"GONT_UNSHARE=exec",
			"GONT_NODE="+n.name,
			"GONT_NETWORK="+n.Network.Name)
	}

	return c
}

func (n *BaseNode) Run(cmd string, arg ...string) ([]byte, *exec.Cmd, error) {
	stdout, stderr, c, err := n.Start(cmd, arg...)
	if err != nil {
		return nil, nil, err
	}

	combined := io.MultiReader(stdout, stderr)
	buf, err := io.ReadAll(combined)
	if err != nil {
		return nil, nil, err
	}

	if err = c.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, nil, err
		}
	}

	rlogger := log.WithFields(log.Fields{
		"ns":       n.name,
		"cmd":      cmd,
		"cmd_args": arg,
		"pid":      c.Process.Pid,
		"rc":       c.ProcessState.ExitCode(),
		"sys_time": c.ProcessState.SystemTime(),
	})

	f := rlogger.Info
	if !c.ProcessState.Success() {
		f = rlogger.Error
	}
	f("Process terminated")

	return buf, c, err
}

func (n *BaseNode) Start(cmd string, arg ...string) (io.Reader, io.Reader, *exec.Cmd, error) {
	var err error
	var stdout, stderr io.Reader

	c := n.Command(cmd, arg...)

	if stdout, err = c.StdoutPipe(); err != nil {
		return nil, nil, nil, err
	}

	if stderr, err = c.StderrPipe(); err != nil {
		return nil, nil, nil, err
	}

	logger := log.WithFields(log.Fields{
		"ns":       n.name,
		"cmd":      cmd,
		"cmd_args": arg,
	})

	if err = c.Start(); err != nil {
		logger.WithError(err).Error("Failed to start")

		return nil, nil, c, err
	}

	logger = logger.WithField("pid", c.Process.Pid)

	logger.Info("Process started")

	if log.GetLevel() >= log.DebugLevel {
		slogger := log.WithFields(log.Fields{
			"pid": c.Process.Pid,
		})

		outReader, outWriter := io.Pipe()
		errReader, errWriter := io.Pipe()

		stdout = io.TeeReader(stdout, outWriter)
		stderr = io.TeeReader(stderr, errWriter)

		go io.Copy(slogger.WriterLevel(log.InfoLevel), outReader)
		go io.Copy(slogger.WriterLevel(log.WarnLevel), errReader)
	}

	return stdout, stderr, c, nil
}
