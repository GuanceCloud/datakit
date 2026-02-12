// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import (
	T "testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestMetricElectionInfo(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		electionStatusVec.Reset()
		reg := prometheus.NewRegistry()
		reg.MustRegister(electionStatusVec)

		electionStatusVec.WithLabelValues("host1", "id1", "ns1", StatusSuccess.String()).Set(1)
		electionStatusVec.WithLabelValues("host1", "id1", "ns1", StatusFail.String()).Set(2)
		electionStatusVec.WithLabelValues("host1", "id1", "ns1", StatusImpeached.String()).Set(3)

		mfs, err := reg.Gather()
		assert.NoError(t, err)
		ei := MetricElectionInfo(mfs[0])
		t.Logf("ei: %+#v", ei)

		assert.Equal(t, StatusImpeached.String(), ei.Status)
	})
}
