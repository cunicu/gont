// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"runtime/debug"
	"time"
)

func ReadVCSInfos() (bool, string, bool, time.Time) {
	if info, ok := debug.ReadBuildInfo(); ok {
		rev := "unknown"
		dirty := false
		btime := time.Time{}

		for _, v := range info.Settings {
			switch v.Key {
			case "vcs.revision":
				rev = v.Value
			case "vcs.modified":
				dirty = v.Value == "true"
			case "vcs.time":
				btime, _ = time.Parse(time.RFC3339, v.Value)
			}
		}

		return true, rev, dirty, btime
	} else {
		return false, "", false, time.Time{}
	}
}
