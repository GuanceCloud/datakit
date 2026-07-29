// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestNetPathProbeDependencyDoesNotRegisterStandaloneInput(t *testing.T) {
	_, registered := inputs.AllInputs["netpath"]
	assert.False(t, registered)
}
