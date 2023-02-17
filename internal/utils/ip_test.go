// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"net"
	"testing"

	"github.com/stv0g/gont/internal/utils"
)

func TestIPAddressRangeV4(t *testing.T) {
	_, src, err := net.ParseCIDR("192.168.0.0/24")
	if err != nil {
		t.Fatal(err)
	}

	fromAddr, toAddr := utils.AddressRange(src)

	expFromAddr := net.ParseIP("192.168.0.0")
	expToAddr := net.ParseIP("192.168.1.0")

	if !fromAddr.Equal(expFromAddr) {
		t.Fail()
	}

	if !toAddr.Equal(expToAddr) {
		t.Fail()
	}
}

func TestIPAddressRangeV6(t *testing.T) {
	_, src, err := net.ParseCIDR("fc00::/8")
	if err != nil {
		t.Fatal(err)
	}

	fromAddr, toAddr := utils.AddressRange(src)

	expFromAddr := net.ParseIP("fc00::")
	expToAddr := net.ParseIP("fd00::")

	if !fromAddr.Equal(expFromAddr) {
		t.Fail()
	}

	if !toAddr.Equal(expToAddr) {
		t.Fail()
	}
}
