// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package unix

//nolint:stylecheck,revive
type SiginfoChld struct {
	Signo  int32     // Signal number.
	Errno  int32     // If non-zero, an errno value associated with this signal.
	Code   int32     // Signal code.
	_      int32     // Padding
	Pid    int       // Which child.
	Status int       // Exit value or signal.
	Uid    int       // Real user ID of sending process.
	_      [100]byte // Padding
}
