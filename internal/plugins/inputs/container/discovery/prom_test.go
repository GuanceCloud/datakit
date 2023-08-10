// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPromRunnerWithToml(t *testing.T) {
	t.Run("parse-prom-01", func(t *testing.T) {
		cfg := `
[[inputs.prom]]
  urls = ["http://172.16.0.2:8080/metrics"]
  source = "testing-prom"
  interval = "10s"
  [inputs.prom.tags]
    some_tag = "some_value"
    more_tag = "some_other_value"
`
		r, err := newPromRunnerWithTomlConfig(cfg)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(r))
		assert.Equal(t, "testing-prom", r[0].conf.Source)
	})

	t.Run("parse-prom-01", func(t *testing.T) {
		cfg := `
[[inputs.prom]]
  urls = ["http://172.16.0.2:8080/metrics"]
  source = "testing-prom"
  interval = "10s"
  [inputs.prom.tags]
    some_tag = "some_value"
    more_tag = "some_other_value"

[[inputs.prom]]
  urls = ["http://172.16.0.3:8080/metrics"]
  source = "testing-prom-02"
  interval = "10s"
`
		r, err := newPromRunnerWithTomlConfig(cfg)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(r))
		assert.Equal(t, "testing-prom", r[0].conf.Source)
		assert.Equal(t, "testing-prom-02", r[1].conf.Source)
	})
}
