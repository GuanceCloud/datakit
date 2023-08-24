// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package enrichment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapEtherType(t *testing.T) {
	assert.Equal(t, "", MapEtherType(0))
	assert.Equal(t, "", MapEtherType(0x8888))
	assert.Equal(t, "IPv4", MapEtherType(0x0800))
	assert.Equal(t, "IPv6", MapEtherType(0x86DD))
}
