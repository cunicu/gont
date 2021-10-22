module github.com/stv0g/gont

go 1.17

require (
	github.com/go-ping/ping v0.0.0-20211014180314-6e2b003bffdd
	github.com/sirupsen/logrus v1.8.1
	github.com/vishvananda/netlink v1.1.1-0.20210924202909-187053b97868
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f
	golang.org/x/sys v0.0.0-20210315160823-c6e025ad8005
	kernel.org/pub/linux/libs/security/libcap/cap v1.2.59
)

require (
	github.com/google/uuid v1.2.0 // indirect
	golang.org/x/net v0.0.0-20210316092652-d523dce5a7f4 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.59 // indirect
)

replace github.com/vishvananda/netlink => github.com/stv0g/netlink v1.1.1-gont

replace github.com/vishvananda/netns => github.com/stv0g/netns v0.0.0-20211014154538-c4d5d2062cf1
