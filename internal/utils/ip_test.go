// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"net"
	"testing"

	"cunicu.li/gont/v2/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestIPAddressRangeV4(t *testing.T) {
	_, src, err := net.ParseCIDR("192.168.0.0/24")
	require.NoError(t, err)

	fromAddr, toAddr := utils.AddressRange(src)

	expFromAddr := net.ParseIP("192.168.0.0")
	expToAddr := net.ParseIP("192.168.1.0")

	require.True(t, fromAddr.Equal(expFromAddr))
	require.True(t, toAddr.Equal(expToAddr))
}

func TestIPAddressRangeV6(t *testing.T) {
	_, src, err := net.ParseCIDR("fc00::/8")
	require.NoError(t, err)

	fromAddr, toAddr := utils.AddressRange(src)

	expFromAddr := net.ParseIP("fc00::")
	expToAddr := net.ParseIP("fd00::")

	require.True(t, fromAddr.Equal(expFromAddr))
	require.True(t, toAddr.Equal(expToAddr))
}
