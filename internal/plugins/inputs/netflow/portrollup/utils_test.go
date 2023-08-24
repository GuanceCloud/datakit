// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package portrollup

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPortToString(t *testing.T) {
	assert.Equal(t, "65535", PortToString(math.MaxUint16))
	assert.Equal(t, "10", PortToString(10))
	assert.Equal(t, "0", PortToString(0))
	assert.Equal(t, "*", PortToString(-1))
	assert.Equal(t, "invalid", PortToString(-10))
}
