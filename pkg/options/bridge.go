// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"time"

	nl "github.com/vishvananda/netlink"
)

// MulticastSnooping configures multicast snooping.
type MulticastSnooping bool

func (mcs MulticastSnooping) Apply(b *nl.Bridge) {
	v := bool(mcs)
	b.MulticastSnooping = &v
}

// VLANfiltering configures VLAN filtering.
// When disabled, the bridge will not consider the VLAN tag when handling packets
type VLANFiltering bool

func (vf VLANFiltering) Apply(b *nl.Bridge) {
	v := bool(vf)
	b.VlanFiltering = &v
}

// AgingTime configures the bridge's FDB entries ageing time, ie the number of seconds a MAC address will be kept in the FDB after a packet has been received from that address.
// After this time has passed, entries are cleaned up.
type AgingTime time.Duration

func (at AgingTime) Apply(b *nl.Bridge) {
	att := time.Duration(at)
	v := uint32(att.Seconds())
	b.AgeingTime = &v
}

// HelloTime sets the time in seconds between hello packets sent by the bridge, when it is a root bridge or a designated bridges.
// Only relevant if STP is enabled. Valid values are between 1 and 10 seconds.
type HelloTime time.Duration

func (ht HelloTime) Apply(b *nl.Bridge) {
	htt := time.Duration(ht)
	v := uint32(htt.Seconds())
	b.HelloTime = &v
}
