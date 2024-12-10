// SPDX-FileCopyrightText: 2023 The Go Authors.
// SPDX-License-Identifier: BSD-3-Clause

package unix

import (
	"errors"
	"os"
	"sync"
	"syscall"
)

// From Go: src/os/pidfd_linux.go

var checkPidfdOnce = sync.OnceValue(checkPidfd) //nolint:gochecknoglobals

const (
	_P_PIDFD = 3 //nolint:stylecheck,revive

	pidfdSendSignalTrap uintptr = 424
	pidfdOpenTrap       uintptr = 434
)

func PidFDWorks() bool {
	return checkPidfdOnce() == nil
}

func PidFDSendSignal(pidfd uintptr, s syscall.Signal) error {
	_, _, errno := syscall.Syscall(pidfdSendSignalTrap, pidfd, uintptr(s), 0)
	if errno != 0 {
		return errno
	}
	return nil
}

func PidFDOpen(pid, flags int) (uintptr, error) {
	pidfd, _, errno := syscall.Syscall(pidfdOpenTrap, uintptr(pid), uintptr(flags), 0)
	if errno != 0 {
		return ^uintptr(0), errno
	}
	return pidfd, nil
}

// checkPidfd checks whether all required pidfd-related syscalls work.
// This consists of pidfd_open and pidfd_send_signal syscalls, and waitid
// syscall with idtype of P_PIDFD.
//
// Reasons for non-working pidfd syscalls include an older kernel and an
// execution environment in which the above system calls are restricted by
// seccomp or a similar technology.
func checkPidfd() error {
	// Get a pidfd of the current process (opening of "/proc/self" won't
	// work for waitid).
	fd, err := PidFDOpen(syscall.Getpid(), 0)
	if err != nil {
		return os.NewSyscallError("pidfd_open", err)
	}
	defer syscall.Close(int(fd))

	// Check waitid(P_PIDFD) works.
	for {
		_, _, err = syscall.Syscall6(syscall.SYS_WAITID, _P_PIDFD, fd, 0, syscall.WEXITED, 0, 0)
		if !errors.Is(err, syscall.EINTR) {
			break
		}
	}
	// Expect ECHILD from waitid since we're not our own parent.
	if !errors.Is(err, syscall.ECHILD) {
		return os.NewSyscallError("pidfd_wait", err)
	}

	// Check pidfd_send_signal works (should be able to send 0 to itself).
	if err := PidFDSendSignal(fd, 0); err != nil {
		return os.NewSyscallError("pidfd_send_signal", err)
	}

	return nil
}
