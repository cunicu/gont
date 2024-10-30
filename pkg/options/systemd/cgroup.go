// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package cgroup

import (
	"net"
	"time"

	sdbus "github.com/coreos/go-systemd/v22/dbus"
	"github.com/godbus/dbus/v5"
)

// CPU Accounting and Control

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#CPUAccounting=
func CPUAccounting(enable bool) Property {
	return Property(sdbus.Property{
		Name:  "CPUAccounting",
		Value: dbus.MakeVariant(enable),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#CPUWeight=weight
func CPUWeight(weight uint64) Property {
	return Property(sdbus.Property{
		Name:  "CPUWeight",
		Value: dbus.MakeVariant(weight),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#CPUWeight=weight
func StartupCPUWeight(weight uint64) Property {
	return Property(sdbus.Property{
		Name:  "StartupCPUWeight",
		Value: dbus.MakeVariant(weight),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#CPUQuota=
func CPUQuota(quota float32) Property {
	return Property(sdbus.Property{
		Name:  "CPUQuotaPerSecUSec",
		Value: dbus.MakeVariant(uint64(quota * 1e6)),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#CPUQuotaPeriodSec=
func CPUQuotaPeriod(period time.Duration) Property {
	return Property(sdbus.Property{
		Name:  "CPUQuotaPerSecUSec",
		Value: dbus.MakeVariant(uint64(period.Microseconds())), //nolint:gosec
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#AllowedCPUs=
func AllowedCPUs(mask uint64) Property {
	return Property(sdbus.Property{
		Name:  "AllowedCPUs",
		Value: makeCpuSet(mask),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#AllowedCPUs=
func StartupAllowedCPUs(mask uint64) Property {
	return Property(sdbus.Property{
		Name:  "StartupAllowedCPUs",
		Value: makeCpuSet(mask),
	})
}

// Memory Accounting and Control

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryAccounting=
func MemoryAccounting(enable bool) Property {
	return Property(sdbus.Property{
		Name:  "MemoryAccounting",
		Value: dbus.MakeVariant(enable),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryMin=bytes,%20MemoryLow=bytes
func MemoryMin(bytes uint64) Property {
	return Property(sdbus.Property{
		Name:  "MemoryMin",
		Value: dbus.MakeVariant(bytes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryMin=bytes,%20MemoryLow=bytes
func MemoryLow(bytes uint64) Property {
	return Property(sdbus.Property{
		Name:  "MemoryLow",
		Value: dbus.MakeVariant(bytes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryMin=bytes,%20MemoryLow=bytes
func StartupMemoryLow(bytes uint64) Property {
	return Property(sdbus.Property{
		Name:  "StartupMemoryLow",
		Value: dbus.MakeVariant(bytes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryMin=bytes,%20MemoryLow=bytes
func DefaultStartupMemoryLow(bytes uint64) Property {
	return Property(sdbus.Property{
		Name:  "DefaultStartupMemoryLow",
		Value: dbus.MakeVariant(bytes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryHigh=bytes
func MemoryHigh(bytes uint64) Property {
	return Property(sdbus.Property{
		Name:  "MemoryHigh",
		Value: dbus.MakeVariant(bytes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryHigh=bytes
func StartupMemoryHigh(bytes uint64) Property {
	return Property(sdbus.Property{
		Name:  "StartupMemoryHigh",
		Value: dbus.MakeVariant(bytes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryMax=bytes
func MemoryMax(bytes uint64) Property {
	return Property(sdbus.Property{
		Name:  "MemoryMax",
		Value: dbus.MakeVariant(bytes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryMax=bytes
func StartupMemoryMax(bytes uint64) Property {
	return Property(sdbus.Property{
		Name:  "StartupMemoryMax",
		Value: dbus.MakeVariant(bytes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemorySwapMax=bytes
func MemorySwapMax(bytes uint64) Property {
	return Property(sdbus.Property{
		Name:  "MemorySwapMax",
		Value: dbus.MakeVariant(bytes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemorySwapMax=bytes
func StartupMemorySwapMax(bytes uint64) Property {
	return Property(sdbus.Property{
		Name:  "StartupMemorySwapMax",
		Value: dbus.MakeVariant(bytes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryZSwapMax=bytes
func MemoryZSwapMax(bytes uint64) Property {
	return Property(sdbus.Property{
		Name:  "MemoryZSwapMax",
		Value: dbus.MakeVariant(bytes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryZSwapMax=bytes
func StartupMemoryZSwapMax(bytes uint64) Property {
	return Property(sdbus.Property{
		Name:  "StartupMemoryZSwapMax",
		Value: dbus.MakeVariant(bytes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryZSwapWriteback=
func MemoryZSwapWriteback(enable bool) Property {
	return Property(sdbus.Property{
		Name:  "MemoryZSwapWriteback",
		Value: dbus.MakeVariant(enable),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#AllowedMemoryNodes=
func AllowedMemoryNodes(mask uint64) Property {
	return Property(sdbus.Property{
		Name:  "AllowedMemoryNodes",
		Value: makeCpuSet(mask),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#AllowedMemoryNodes=
func StartupAllowedMemoryNodes(mask uint64) Property {
	return Property(sdbus.Property{
		Name:  "StartupAllowedMemoryNodes",
		Value: makeCpuSet(mask),
	})
}

// Process Accounting and Control

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#TasksAccounting=
func TasksAccounting(enable bool) Property {
	return Property(sdbus.Property{
		Name:  "TasksAccounting",
		Value: dbus.MakeVariant(enable),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#TasksMax=N
func TasksMax(max uint64) Property {
	return Property(sdbus.Property{
		Name:  "TasksMax",
		Value: dbus.MakeVariant(max),
	})
}

// IO Accounting and Control

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IOAccounting=
func IOAccounting(enable bool) Property {
	return Property(sdbus.Property{
		Name:  "IOAccounting",
		Value: dbus.MakeVariant(enable),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IOWeight=weight
func IOWeight(weight uint64) Property {
	return Property(sdbus.Property{
		Name:  "IOWeight",
		Value: dbus.MakeVariant(weight),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IOWeight=weight
func StartupIOWeight(weight uint64) Property {
	return Property(sdbus.Property{
		Name:  "StartupIOWeight",
		Value: dbus.MakeVariant(weight),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IODeviceWeight=device%20weight
// TODO: Implement
// func IODeviceWeight(device string, weight uint64) Property {
// 	return Property(dbus.Property{
// 		Name:  "IODeviceWeight",
// 		Value: dbus.MakeVariant(),
// 	})
// }

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IOReadBandwidthMax=device%20bytes
// TODO: Implement
// func IOReadBandwidthMax(device string, bytes uint64) Property {
// 	return Property(sdbus.Property{
// 		Name:  "IOReadBandwidthMax",
// 		Value: dbus.MakeVariant(),
// 	})
// }

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IOReadBandwidthMax=device%20bytes
// TODO: Implement
// func IOWriteBandwidthMax(device string, bytes uint64) Property {
// 	return Property(dbus.Property{
// 		Name:  "IOWriteBandwidthMax",
// 		Value: dbus.MakeVariant(),
// 	}
// }

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IOReadIOPSMax=device%20IOPS
// TODO: Implement
// func IOReadIOPSMax(device string, iops uint64) Property {
// 	return Property(sdbus.Property{
// 		Name:  "IOReadIOPSMax",
// 		Value: dbus.MakeVariant(),
// 	})
// }

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IOReadIOPSMax=device%20IOPS
// TODO: Implement
// func IOWriteIOPSMax(device string, iops uint64) Property {
// 	return Property(sdbus.Property{
// 		Name:  "IOWriteIOPSMax",
// 		Value: dbus.MakeVariant(),
// 	})
// }

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IODeviceLatencyTargetSec=device%20target
// TODO: Implement
// func IODeviceLatencyTarget(device string, target time.Duration) Property {
// 	return Property(sdbus.Property{
// 		Name:  "IODeviceLatencyTargetSec",
// 		Value: dbus.MakeVariant(),
// 	})
// }

// Network Accounting and Control

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IPAccounting=
func IPAccounting(enable bool) Property {
	return Property(sdbus.Property{
		Name:  "IPAccounting",
		Value: dbus.MakeVariant(enable),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IPAddressAllow=ADDRESS%5B/PREFIXLENGTH%5D%E2%80%A6
func IPAddressAllow(prefixes ...net.IPNet) Property {
	return Property(sdbus.Property{
		Name:  "IPAddressAllow",
		Value: makePrefixList(prefixes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IPAddressAllow=ADDRESS%5B/PREFIXLENGTH%5D%E2%80%A6
func IPAddressDeny(prefixes ...net.IPNet) Property {
	return Property(sdbus.Property{
		Name:  "IPAddressDeny",
		Value: makePrefixList(prefixes),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#SocketBindAllow=bind-rule
func SocketBindAllow(policies ...BindPolicy) Property {
	val, err := makeBindPolicies(policies)
	if err != nil {
		panic(err)
	}

	return Property(sdbus.Property{
		Name:  "SocketBindAllow",
		Value: val,
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#SocketBindAllow=bind-rule
func SocketBindDeny(policies ...BindPolicy) Property {
	val, err := makeBindPolicies(policies)
	if err != nil {
		panic(err)
	}

	return Property(sdbus.Property{
		Name:  "SocketBindDeny",
		Value: val,
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#RestrictNetworkInterfaces=
func RestrictNetworkInterfaces(allow bool, intfs ...string) Property {
	return Property(sdbus.Property{
		Name:  "RestrictNetworkInterfaces",
		Value: makeNetworkInterfaceAllowList(allow, intfs),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#NFTSet=family:table:set
// TODO: Implement
// type NFTSetDefinition struct {
// 	Source string
// 	Family string
// 	Table string
// 	Set string
// }
// func NFTSet(sets []NFTSetDefinition) Property {
// 	return Property(sdbus.Property{
// 		Name:  "NFTSet",
// 		Value: dbus.MakeVariant(),
// 	})
// }

// BPF Programs

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IPIngressFilterPath=BPF_FS_PROGRAM_PATH
func IPIngressFilterPath(paths ...string) Property {
	return Property(sdbus.Property{
		Name:  "IPIngressFilterPath",
		Value: dbus.MakeVariant(paths),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#IPIngressFilterPath=BPF_FS_PROGRAM_PATH
func IPEgressFilterPath(paths ...string) Property {
	return Property(sdbus.Property{
		Name:  "IPEgressFilterPath",
		Value: dbus.MakeVariant(paths),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#BPFProgram=type:program-path
type BPFProgramAttachType string

const (
	BPFAttachTypeIngress    BPFProgramAttachType = "ingress"
	BPFAttachTypeEgress     BPFProgramAttachType = "egress"
	BPFAttachTypeSockCreate BPFProgramAttachType = "sock_create"
	BPFAttachTypeSockOps    BPFProgramAttachType = "sock_ops"
	BPFAttachTypeDevice     BPFProgramAttachType = "device"
	BPFAttachTypeBind4      BPFProgramAttachType = "bind4"
	BPFAttachTypeBind6      BPFProgramAttachType = "bind6"
	BPFAttachTypeConnect4   BPFProgramAttachType = "connect4"
	BPFAttachTypeConnect6   BPFProgramAttachType = "connect6"
	BPFAttachTypePostBind4  BPFProgramAttachType = "post_bind4"
	BPFAttachTypePostBind6  BPFProgramAttachType = "post_bind6"
	BPFAttachTypeSendmsg4   BPFProgramAttachType = "sendmsg4"
	BPFAttachTypeSendmsg6   BPFProgramAttachType = "sendmsg6"
	BPFAttachTypeSysctl     BPFProgramAttachType = "sysctl"
	BPFAttachTypeRecvmsg4   BPFProgramAttachType = "recvmsg4"
	BPFAttachTypeRecvmsg6   BPFProgramAttachType = "recvmsg6"
	BPFAttachTypeGetsockopt BPFProgramAttachType = "getsockopt"
	BPFAttachTypeSetsockopt BPFProgramAttachType = "setsockopt"
)

type BPFProgramType struct {
	AttachType BPFProgramAttachType `dbus:"-"`
	Path       string               `dbus:"-"`
}

func BPFProgram(programs ...BPFProgramType) Property {
	return Property(sdbus.Property{
		Name:  "BPFProgram",
		Value: dbus.MakeVariant(programs),
	})
}

// Device Access

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#DeviceAllow=
// TODO: Implement
// func DeviceAllow() Property {
// 	return Property(sdbus.Property{
// 		Name:  "DeviceAllow",
// 		Value: dbus.MakeVariant(),
// 	})
// }

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#DevicePolicy=auto%7Cclosed%7Cstrict
type DevicePolicyType string

const (
	DevicePolicyAuto   DevicePolicyType = "auto"
	DevicePolicyClosed DevicePolicyType = "closed"
	DevicePolicyStrict DevicePolicyType = "strict"
)

func DevicePolicy(policy DevicePolicyType) Property {
	return Property(sdbus.Property{
		Name:  "DevicePolicy",
		Value: dbus.MakeVariant(policy),
	})
}

// Control Group Management

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#Slice=
// TODO: Implement
func Slice(slice string) Property {
	return Property(sdbus.Property{
		Name:  "Slice",
		Value: dbus.MakeVariant(slice),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#Delegate=
func Delegate(enable bool) Property {
	return Property(sdbus.Property{
		Name:  "Delegate",
		Value: dbus.MakeVariant(enable),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#DelegateSubgroup=
func DelegateSubgroup(group string) Property {
	return Property(sdbus.Property{
		Name:  "DelegateSubgroup",
		Value: dbus.MakeVariant(group),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#DisableControllers=
func DisableControllers(controllers ...string) Property {
	return Property(sdbus.Property{
		Name:  "DisableControllers",
		Value: dbus.MakeVariant(controllers),
	})
}

// Memory Pressure Control

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#ManagedOOMSwap=auto%7Ckill
// TODO: Implement
// func ManagedOOMSwap() Property {
// 	return Property(sdbus.Property{
// 		Name:  "ManagedOOMSwap",
// 		Value: dbus.MakeVariant(),
// 	})
// }

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#ManagedOOMSwap=auto%7Ckill
// TODO: Implement
// func ManagedOOMMemoryPressure() Property {
// 	return Property(sdbus.Property{
// 		Name:  "ManagedOOMMemoryPressure",
// 		Value: dbus.MakeVariant(),
// 	})
// }

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#ManagedOOMMemoryPressureLimit=
// TODO: Implement
// func ManagedOOMMemoryPressureLimit() Property {
// 	return Property(sdbus.Property{
// 		Name:  "ManagedOOMMemoryPressureLimit",
// 		Value: dbus.MakeVariant(),
// 	})
// }

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#ManagedOOMPreference=none%7Cavoid%7Comit
type ManagedOOMPreferenceMode string

const (
	ManagedOOMPreferenceNone  ManagedOOMPreferenceMode = "none"
	ManagedOOMPreferenceAvoid ManagedOOMPreferenceMode = "avoid"
	ManagedOOMPreferenceOmit  ManagedOOMPreferenceMode = "omit"
)

func ManagedOOMPreference(mode ManagedOOMPreferenceMode) Property {
	return Property(sdbus.Property{
		Name:  "ManagedOOMPreference",
		Value: dbus.MakeVariant(mode),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryPressureWatch=
type MemoryPressureWatchMode string

const (
	MemoryPressureWatchOff  MemoryPressureWatchMode = "off"
	MemoryPressureWatchOn   MemoryPressureWatchMode = "on"
	MemoryPressureWatchSkip MemoryPressureWatchMode = "skip"
)

func MemoryPressureWatch(mode MemoryPressureWatchMode) Property {
	return Property(sdbus.Property{
		Name:  "MemoryPressureWatch",
		Value: dbus.MakeVariant(mode),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#MemoryPressureThresholdSec=
func MemoryPressureThreshold(d time.Duration) Property {
	return Property(sdbus.Property{
		Name:  "MemoryPressureThresholdSec",
		Value: dbus.MakeVariant(uint64(d.Microseconds())), //nolint:gosec
	})
}

// Coredump Control

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html#CoredumpReceive=
func CoredumpReceive(enable bool) Property {
	return Property(sdbus.Property{
		Name:  "CoredumpReceive",
		Value: dbus.MakeVariant(enable),
	})
}
