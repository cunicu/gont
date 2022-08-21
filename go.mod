module github.com/stv0g/gont

go 1.19

require (
	github.com/go-ping/ping v1.1.0
	github.com/google/gopacket v1.1.19
	github.com/google/nftables v0.0.0-20220808154552-2eca00135732
	github.com/vishvananda/netlink v1.2.1-beta.2
	github.com/vishvananda/netns v0.0.0-20211101163701-50045581ed74
	go.uber.org/zap v1.22.0
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e
	golang.org/x/net v0.0.0-20220812174116-3211cb980234
	golang.org/x/sys v0.0.0-20220818161305-2296e01440c6
	kernel.org/pub/linux/libs/security/libcap/cap v1.2.65
)

require (
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/josharian/native v1.0.0 // indirect
	github.com/mdlayher/netlink v1.6.0 // indirect
	github.com/mdlayher/socket v0.2.3 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/sync v0.0.0-20220819030929-7fc1605a5dde // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.65 // indirect
)

// Temporary, until https://github.com/google/gopacket/pull/1042 is merged
replace github.com/google/gopacket => github.com/stv0g/gopacket v0.0.0-20220819110231-82599fdade4a
