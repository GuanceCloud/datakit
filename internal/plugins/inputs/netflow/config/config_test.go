// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/common"
)

func TestListenerConfig_Addr(t *testing.T) {
	listenerConfig := ListenerConfig{
		FlowType: common.TypeNetFlow9,
		BindHost: "127.0.0.1",
		Port:     1234,
	}
	assert.Equal(t, "127.0.0.1:1234", listenerConfig.Addr())
}
