// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package unix

import "syscall"

func Ptrace(request int, pid int, addr uintptr, data uintptr) error {
	if _, _, e1 := syscall.Syscall6(syscall.SYS_PTRACE, uintptr(request), uintptr(pid), uintptr(addr), uintptr(data), 0, 0); e1 != 0 {
		return syscall.Errno(e1)
	}

	return nil
}
