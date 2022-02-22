package ipdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIPCIDR(t *testing.T) {
	cases := []struct {
		title       string
		cidr        string
		expected    string
		isValidCidr bool
	}{
		{
			title:       "correct cidr",
			cidr:        "1.0.0.0/24",
			expected:    "000000010000000000000000",
			isValidCidr: true,
		},
		{
			title:       "correct cidr",
			cidr:        "1.0.0.0/26",
			expected:    "00000001000000000000000000",
			isValidCidr: true,
		},
		{
			title:       "invalid ip length",
			cidr:        "1.0.0.0.0/24",
			expected:    "000000010000000000000000",
			isValidCidr: false,
		},
		{
			title:       "invalid cidr",
			cidr:        "256.0.0.0/24",
			expected:    "",
			isValidCidr: false,
		},
		{
			title:       "invalid ip",
			cidr:        "a.0.0.0/24",
			expected:    "",
			isValidCidr: false,
		},
		{
			title:       "invalid mask",
			cidr:        "1.0.0.0/b",
			expected:    "",
			isValidCidr: false,
		},
	}

	for _, item := range cases {
		ip, err := ParseIPCIDR(item.cidr)
		if !item.isValidCidr {
			assert.Error(t, err)
		} else {
			assert.Equal(t, ip, item.expected)
		}
	}
}
