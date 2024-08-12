// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

type DeviceGroup uint32

const (
	DeviceGroupDefault    DeviceGroup = 0
	DeviceGroupSouthBound DeviceGroup = 1000
	DeviceGroupNorthBound DeviceGroup = 1001
)
