module github.com/stv0g/gont

go 1.17

require (
	github.com/go-ping/ping v0.0.0-20211130115550-779d1e919534
	github.com/google/nftables v0.0.0-20211209220838-6f19c4381e13
	github.com/vishvananda/netlink v1.2.1-beta.2
	github.com/vishvananda/netns v0.0.0-20211101163701-50045581ed74
	go.uber.org/zap v1.20.0
	golang.org/x/sys v0.0.0-20220111092808-5a964db01320
	kernel.org/pub/linux/libs/security/libcap/cap v1.2.62
)

require (
	github.com/BurntSushi/toml v1.0.0 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/josharian/native v0.0.0-20200817173448-b6b71def0850 // indirect
	github.com/koneu/natend v0.0.0-20150829182554-ec0926ea948d // indirect
	github.com/mdlayher/netlink v1.5.0 // indirect
	github.com/mdlayher/socket v0.1.1 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/mod v0.5.1 // indirect
	golang.org/x/net v0.0.0-20220114011407-0dd24b26b47d // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/tools v0.1.8 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	honnef.co/go/tools v0.2.2 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.62 // indirect
)

// replace github.com/vishvananda/netlink => github.com/stv0g/netlink v1.1.1-gont
