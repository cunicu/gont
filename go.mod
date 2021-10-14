module github.com/stv0g/gont

go 1.17

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/vishvananda/netlink v1.1.1-0.20210924202909-187053b97868
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f
	golang.org/x/sys v0.0.0-20200728102440-3e129f6d46b1
	kernel.org/pub/linux/libs/security/libcap/cap v1.2.59
)

require kernel.org/pub/linux/libs/security/libcap/psx v1.2.59 // indirect

replace github.com/vishvananda/netlink v1.1.1-0.20210924202909-187053b97868 => github.com/stv0g/netlink v1.1.1-ipset-family
replace github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f => github.com/stv0g/netns v0.0.0-docker-cgroups-v2

replace github.com/vishvananda/netlink => /home/stv0g_local/workspace/netlink
replace github.com/vishvananda/netns => /home/stv0g_local/workspace/netns
