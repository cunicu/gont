package gont

import (
	"runtime"
	"syscall"

	log "github.com/sirupsen/logrus"
	nl "github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"golang.org/x/sys/unix"
)

type Callback func() error

type Namespace struct {
	netns.NsHandle
	*nl.Handle

	Name string
}

func NewNamespace(name string) (*Namespace, error) {
	ns := &Namespace{
		Name: name,
	}

	log.WithField("ns", name).Info("Creating new namespace")

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

		log.WithField("ns", ns.Name).Info("Deleted namespace")
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

	log.WithField("ns", ns.Name).Debug("Entered namespace")

	return func() {
		// Restore original netns namespace
		if err := unix.Setns(curNetNs, syscall.CLONE_NEWNET); err != nil {
			panic(err)
		}

		log.WithField("ns", ns.Name).Debug("Left namespace")

		runtime.UnlockOSThread()
	}, nil
}
