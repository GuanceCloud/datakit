// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestReadEnv(t *T.T) {
	t.Run("basic", func(t *T.T) {
		ipt := defaultInput()
		envs := map[string]string{
			"ENV_INPUT_HOSTOBJECT_DISABLE_CLOUD_PROVIDER_SYNC": "true",
		}
		ipt.ReadEnv(envs)

		assert.True(t, ipt.DisableCloudProviderSync)
	})
}
