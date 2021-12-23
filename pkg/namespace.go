package gont

import (
	"runtime"
	"syscall"

	nl "github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"go.uber.org/zap"

	"golang.org/x/sys/unix"
)

type Callback func() error

type Namespace struct {
	netns.NsHandle
	*nl.Handle

	Name string

	logger *zap.Logger
}

func NewNamespace(name string) (*Namespace, error) {
	ns := &Namespace{
		Name:   name,
		logger: zap.L().Named("namespace").With(zap.String("ns", name)),
	}

	ns.logger.Info("Creating new namespace")

	return ns, ns.createNamespaceAndNetlinkHandles()
}

func (ns *Namespace) createNamespaceAndNetlinkHandles() error {
	var err error

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save fd to current network namespace
	curNetNs, err := syscall.Open("/proc/self/ns/net", syscall.O_RDONLY, 0777)
	if err != nil {
		return err
	}

	// Create new named namespace
	if ns.NsHandle, err = netns.NewNamed(ns.Name); err != nil {
		return err
	}

	// Create a netlink socket handle while we are in the namespace
	if ns.Handle, err = nl.NewHandle(); err != nil {
		return err
	}

	// Restore original netns namespace
	if err := unix.Setns(curNetNs, syscall.CLONE_NEWNET); err != nil {
		return err
	}

	return nil
}

func (ns *Namespace) Close() error {
	if ns.NsHandle >= 0 {
		if err := netns.DeleteNamed(ns.Name); err != nil {
			return err
		}

		ns.logger.Info("Deleted namespace")
	}

	return nil
}

func (n *Namespace) RunFunc(cb Callback) error {
	exit, _ := n.Enter()
	defer exit()

	errCb := cb()

	return errCb
}

func (ns *Namespace) Enter() (func(), error) {
	runtime.LockOSThread()

	// Save fd to current network namespace
	curNetNs, err := syscall.Open("/proc/self/ns/net", syscall.O_RDONLY, 0777)
	if err != nil {
		return nil, err
	}

	// Switch to network namespace
	if err := unix.Setns(int(ns.NsHandle), syscall.CLONE_NEWNET); err != nil {
		return nil, err
	}

	ns.logger.Debug("Entered namespace")

	return func() {
		// Restore original netns namespace
		if err := unix.Setns(curNetNs, syscall.CLONE_NEWNET); err != nil {
			panic(err)
		}

		ns.logger.Debug("Left namespace")

		runtime.UnlockOSThread()
	}, nil
}
