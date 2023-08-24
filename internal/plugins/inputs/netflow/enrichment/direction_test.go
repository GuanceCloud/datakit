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

func TestRemapDirection(t *testing.T) {
	assert.Equal(t, "ingress", RemapDirection(uint32(0)))
	assert.Equal(t, "egress", RemapDirection(uint32(1)))
	assert.Equal(t, "ingress", RemapDirection(uint32(99))) // invalid direction will default to ingress
}
