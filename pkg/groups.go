// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

type DeviceGroup int

const (
	DeviceGroupDefault    DeviceGroup = 0
	DeviceGroupSouthBound DeviceGroup = 1000 + iota
	DeviceGroupNorthBound
)
