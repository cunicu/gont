// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package unix

const (
	CLD_EXITED    = 1 // Child has exited
	CLD_KILLED    = 2 // Child was killed
	CLD_DUMPED    = 3 // Child terminated abnormally
	CLD_TRAPPED   = 4 // Traced child has trapped
	CLD_STOPPED   = 5 // Child has stopped
	CLD_CONTINUED = 6 // Stopped child has continued
)
