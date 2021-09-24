package gont

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"

	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type Network struct {
	Prefix string

	Nodes   []*Node
	TempDir string
}

func NewNetwork(name string) *Network {
	if name == "" {
		name = fmt.Sprintf("gont-%d", rand.Intn(1<<16))
	}

	n := &Network{
		Prefix:  name + "-",
		TempDir: filepath.Join(os.TempDir(), name),
	}

	n.Reset()

	return n
}

func (n *Network) Reset() error {
	out, _, err := n.Run("ip", "netns", "list")
	if err != nil {
		return err
	}

	for _, ns := range strings.Split(string(out), "\n") {
		if len(ns) == 0 || !strings.HasPrefix(ns, n.Prefix) {
			continue
		}

		log.WithField("ns", ns).Warn("Removing stale namespace")

		if _, _, err := n.Run("ip", "netns", "delete", ns); err != nil {
			return err
		}
	}

	_, _, err = n.Run("sed", "-i", "/# gont:"+n.Prefix+"ns$/d", hostsFile)

	return err
}

func (n *Network) Close() error {
	for _, node := range n.Nodes {
		node.NS.Close()
	}

	if n.TempDir != "" {
		os.RemoveAll(n.TempDir)
	}

	return n.Reset()
}

func (n *Network) Run(cmd ...string) ([]byte, *exec.Cmd, error) {
	var err error
	var out []byte

	c := exec.Command(cmd[0], cmd[1:]...)

	if out, err = c.CombinedOutput(); err != nil {
		log.WithField("cmd", cmd).WithError(err).Error("Failed to execute")
		if len(out) > 0 && string(out) != "\n" {
			os.Stdout.Write(out)
		}
		return out, c, err
	}

	log.WithField("cmd", cmd).Info("Run command")

	if len(out) > 0 {
		os.Stdout.Write(out)
	}

	return out, c, nil
}

func (n *Network) RunAsync(cmd ...string) (*exec.Cmd, *io.ReadCloser, *io.ReadCloser, error) {
	c := exec.Command(cmd[0], cmd[1:]...)

	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	log.WithField("cmd", cmd).Info("Run async command")

	err = c.Start()
	if err != nil {
		return nil, nil, nil, err
	}

	return c, &stdout, &stderr, nil
}

func (n *Network) RunNS(ns Namespace, cmd ...string) ([]byte, *exec.Cmd, error) {
	cmd = append([]string{"ip", "netns", "exec", ns.String()}, cmd...)
	return n.Run(cmd...)
}

func (n *Network) RunNSAsync(ns Namespace, cmd ...string) (*exec.Cmd, *io.ReadCloser, *io.ReadCloser, error) {
	cmd = append([]string{"ip", "netns", "exec", ns.String()}, cmd...)
	return n.RunAsync(cmd...)
}

func (n *Network) GoRunNSAsync(ns Namespace, cmd ...string) (*exec.Cmd, *io.ReadCloser, *io.ReadCloser, error) {
	tmp := filepath.Join(n.TempDir, fmt.Sprintf("go-build-%d", rand.Intn(1<<16)))
	_, _, err := n.Run("go", "build", "-o", tmp, cmd[0])
	if err != nil {
		return nil, nil, nil, err
	}

	cmd = append([]string{"ip", "netns", "exec", ns.String(), tmp}, cmd[1:]...)
	return n.RunAsync(cmd...)
}

func (n *Network) GoRunAsync(cmd ...string) (*exec.Cmd, *io.ReadCloser, *io.ReadCloser, error) {
	tmp := filepath.Join(n.TempDir, fmt.Sprintf("go-build-%d", rand.Intn(1<<16)))
	_, _, err := n.Run("go", "build", "-o", tmp, cmd[0])
	if err != nil {
		return nil, nil, nil, err
	}

	cmd[0] = tmp
	return n.RunAsync(cmd...)
}

func (n *Network) CreateNamespace(name string) (Namespace, error) {
	if _, _, err := n.Run("ip", "netns", "add", n.Prefix+name); err != nil {
		return Namespace{}, err
	}

	return Namespace{
		Name:    n.Prefix + name,
		Network: n,
	}, nil
}

func (n *Network) AddToHostsFile(addr net.IP, name string) error {
	_, _, err := n.Run("sed", "-i", "$ a "+addr.String()+" "+name+" # gont:"+n.Prefix+"ns", hostsFile)
	return err
}

func (n *Network) RemoveFromHostsFile(addr net.IP) error {
	_, _, err := n.Run("sed", "-i", "/^"+addr.String()+"/d", hostsFile)
	return err
}

func (n *Network) RemoveAllFromHostsFile() error {
	_, _, err := n.Run("sed", "-i", "/ # gont:"+n.Prefix+"ns$/d", hostsFile)
	return err
}
