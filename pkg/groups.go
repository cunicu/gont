package gont

type DeviceGroup int

const (
	DeviceGroupDefault    DeviceGroup = 0
	DeviceGroupSouthBound DeviceGroup = 1000 + iota
	DeviceGroupNorthBound
)
