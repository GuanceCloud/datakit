// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package windowsremote

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIPsFromCIDRs(t *testing.T) {
	var (
		in = []string{
			"172.16.10.0/29",
			"192.168.1.0/30",
		}
		out = []string{
			"172.16.10.1",
			"172.16.10.2",
			"172.16.10.3",
			"172.16.10.4",
			"172.16.10.5",
			"172.16.10.6",
			"192.168.1.1",
			"192.168.1.2",
		}
	)

	res, err := getIPsFromCIDRs(in)
	assert.NoError(t, err)
	assert.Equal(t, out, res)
}
