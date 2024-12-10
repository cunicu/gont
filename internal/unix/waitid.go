// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package unix

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

func Waitid(idType int, id int, info *SiginfoChld, options int, rusage *unix.Rusage) (err error) {
	if _, _, e1 := unix.Syscall6(unix.SYS_WAITID, uintptr(idType), uintptr(id), uintptr(unsafe.Pointer(info)), uintptr(options), uintptr(unsafe.Pointer(rusage)), 0); e1 != 0 {
		return e1
	}

	return nil
}
