// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	T "testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

func TestMainCfgToml(t *T.T) {
	// used to show default datakit.conf
	t.Run("default", func(t *T.T) {
		c := DefaultConfig()
		c.Dataway.Sinkers = append(c.Dataway.Sinkers,
			&dataway.Sinker{
				Categories: []string{},
				Filters:    []string{},
			})
		t.Logf("conf: %s", c.String())
	})
}
