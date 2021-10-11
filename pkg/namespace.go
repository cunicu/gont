package gont

import (
	"runtime"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"

	"golang.org/x/sys/unix"
)

type Callback func() error

type Namespace struct {
	netns.NsHandle

	Name string
}

func NewNamespace(name string) (*Namespace, error) {
	ns := &Namespace{
		NsHandle: -1,
		Name:     name,
	}

	log.WithField("ns", name).Info("Creating new namespace")

	if err := ns.Ensure(); err != nil {
		return nil, err
	}

	return ns, nil
}

// Ensure ensures that the namespace exists in the kernel
func (ns *Namespace) Ensure() error {
	if ns.NsHandle < 0 {
		return ns.RunFunc(func() error { return nil })
	} else {
		return nil
	}
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
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save fd to current network namespace
	curNetNs, err := syscall.Open("/proc/self/ns/net", syscall.O_RDONLY, 0777)
	if err != nil {
		panic(err)
	}

	if n.NsHandle == -1 {
		// Lazy create new namespace
		if n.NsHandle, err = netns.NewNamed(n.Name); err != nil {
			panic(err)
		}
	} else {
		// Switch to existing network namespace
		if err := unix.Setns(int(n.NsHandle), syscall.CLONE_NEWNET); err != nil {
			panic(err)
		}
	}

	errCb := cb()

	// Restore original netns namespace
	if err := unix.Setns(curNetNs, syscall.CLONE_NEWNET); err != nil {
		panic(err)
	}

	return errCb
}
