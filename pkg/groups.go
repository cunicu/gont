package gont

type DeviceGroup int

const (
	Default       DeviceGroup = 0
	NATSouthBound DeviceGroup = 1000 + iota
	NATNorthBound
)
