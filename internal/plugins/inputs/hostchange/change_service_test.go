// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceCheckerModifyUsesUnifiedDiff(t *testing.T) {
	checker := &ServiceChecker{}
	changesByType := make(ServiceChangesByType)
	oldService := &Service{
		Name:        "example.service",
		Path:        "/etc/systemd/system/example.service",
		Content:     "old\n",
		ContentHash: 1,
	}
	newService := &Service{
		Name:        "example.service",
		Path:        "/etc/systemd/system/example.service",
		Content:     "new\n",
		ContentHash: 2,
	}

	checker.modify(changesByType)("example.service", []*Service{newService}, []*Service{oldService})

	items := changesByType[string(ChangeIDModifyService)]
	require.Len(t, items, 1)
	assert.Equal(t, `--- a/etc/systemd/system/example.service
+++ b/etc/systemd/system/example.service
@@ -1 +1 @@
-old
+new
`, items[0].DiffText)
}
