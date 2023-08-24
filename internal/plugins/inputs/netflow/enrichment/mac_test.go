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

func TestFormatMacAddress(t *testing.T) {
	assert.Equal(t, "82:a5:6e:a5:aa:99", FormatMacAddress(uint64(143647037565593)))
	assert.Equal(t, "00:00:00:00:00:00", FormatMacAddress(uint64(0)))
}
