package options

import nl "github.com/vishvananda/netlink"

type MulticastSnooping bool
type VLANFiltering bool
type AgingTime uint32
type HelloTime uint32

func (mcs MulticastSnooping) Apply(b *nl.Bridge) {
	v := bool(mcs)
	b.MulticastSnooping = &v
}

func (vf VLANFiltering) Apply(b *nl.Bridge) {
	v := bool(vf)
	b.VlanFiltering = &v
}

func (at AgingTime) Apply(b *nl.Bridge) {
	v := uint32(at)
	b.AgeingTime = &v
}

func (ht HelloTime) Apply(b *nl.Bridge) {
	v := uint32(ht)
	b.HelloTime = &v
}
