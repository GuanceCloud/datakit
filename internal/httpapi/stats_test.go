// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/election"
)

func TestFormatElectionInfo(t *testing.T) {
	t.Run("missing election metric", func(t *testing.T) {
		require.Equal(t, "not-ready", formatElectionInfo(nil))
	})

	t.Run("without update time", func(t *testing.T) {
		require.Equal(t, "ns1::success|host1", formatElectionInfo(&election.ElectionInfo{
			Namespace:  "ns1",
			Status:     "success",
			WhoElected: "host1",
		}))
	})

	t.Run("with update time", func(t *testing.T) {
		require.Contains(t, formatElectionInfo(&election.ElectionInfo{
			Namespace:  "ns1",
			Status:     "success",
			WhoElected: "host1",
			UpdateTime: 1,
		}), "ns1::success|host1(")
	})
}
