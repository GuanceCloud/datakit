// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package goflowlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/common"
)

func TestStartFlowRoutine_invalidType(t *testing.T) {
	state, err := StartFlowRoutine("invalid", "my-hostname", 1234, 1, "my-ns", make(chan *common.Flow))
	assert.EqualError(t, err, "unknown flow type: invalid")
	assert.Nil(t, state)
}
