// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

func useTestDataway(t *testing.T, replacement *dataway.Dataway) {
	t.Helper()

	original := config.Cfg.Dataway
	t.Cleanup(func() {
		config.Cfg.Dataway = original
	})
	config.Cfg.Dataway = replacement
}

func TestUseTestDatawayRestoresGlobalState(t *testing.T) {
	original := config.Cfg.Dataway
	replacement := dataway.NewDefaultDataway()

	t.Run("replace", func(t *testing.T) {
		useTestDataway(t, replacement)
		if config.Cfg.Dataway != replacement {
			t.Fatalf("got dataway %p, want replacement %p", config.Cfg.Dataway, replacement)
		}
	})

	if config.Cfg.Dataway != original {
		t.Fatalf("got dataway %p after cleanup, want original %p", config.Cfg.Dataway, original)
	}
}
