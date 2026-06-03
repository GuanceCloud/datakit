// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunMonitorRejectsUnknownModule(t *testing.T) {
	err := RunMonitor(MonitorOptions{Module: "filter,not_exist"})
	require.ErrorContains(t, err, "has no module:[not_exist],check please")
}

func TestExistsModuleAcceptsShortAndLongNames(t *testing.T) {
	require.Empty(t, existsModule([]string{"F", "filter", "In", "inputs", "WAL", "wal"}))
}
