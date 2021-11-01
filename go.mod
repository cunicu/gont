module github.com/stv0g/gont

go 1.17

require (
	github.com/go-ping/ping v0.0.0-20211014180314-6e2b003bffdd
	github.com/google/nftables v0.0.0-20210916140115-16a134723a96
	github.com/sirupsen/logrus v1.8.1
	github.com/vishvananda/netlink v1.1.1-0.20210924202909-187053b97868
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1
	kernel.org/pub/linux/libs/security/libcap/cap v1.2.59
)

require (
	github.com/google/uuid v1.2.0 // indirect
	github.com/koneu/natend v0.0.0-20150829182554-ec0926ea948d // indirect
	github.com/mdlayher/netlink v0.0.0-20191009155606-de872b0d824b // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/net v0.0.0-20210316092652-d523dce5a7f4 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/text v0.3.7 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.59 // indirect
)

replace github.com/vishvananda/netlink => github.com/stv0g/netlink v1.1.1-gont

replace github.com/vishvananda/netns => github.com/stv0g/netns v0.0.0-20211014154538-c4d5d2062cf1
