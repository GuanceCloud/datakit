// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !datakit_elinker && !datakit_lite && !datakit_aws_lambda && with_inputs
// +build !datakit_elinker,!datakit_lite,!datakit_aws_lambda,with_inputs

package inputs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	inputregistry "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestDefaultInputRegistrationMatrix(t *testing.T) {
	_, dialtestingRegistered := inputregistry.AllInputs["dialtesting"]
	_, netpathRegistered := inputregistry.AllInputs["netpath"]

	assert.True(t, dialtestingRegistered)
	assert.True(t, netpathRegistered)
}
