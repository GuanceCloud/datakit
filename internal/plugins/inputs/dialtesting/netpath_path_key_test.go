// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"testing"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeNetPathDialtestingPathKey(t *testing.T) {
	key := makeNetPathDialtestingPathKey("node-a", "task-a")

	assert.Regexp(t, `^np-v1-[0-9a-f]{32}$`, key)
	assert.Equal(t, key, makeNetPathDialtestingPathKey(" node-a ", " task-a "))
	assert.NotEqual(t, key, makeNetPathDialtestingPathKey("node-b", "task-a"))
	assert.NotEqual(t, key, makeNetPathDialtestingPathKey("node-a", "task-b"))
	assert.Empty(t, makeNetPathDialtestingPathKey("", "task-a"))
	assert.Empty(t, makeNetPathDialtestingPathKey("node-a", ""))
}

func TestPointsFeedAddsNetPathPathKey(t *testing.T) {
	oldWorker := dialWorker
	t.Cleanup(func() { dialWorker = oldWorker })
	dialWorker = &worker{
		jobChans:   make(chan *jobData, 1),
		pointCache: map[string]*DataCache{},
		failInfo:   map[string]int{},
	}

	task := &runTaskStub{
		class:      dt.ClassNetPath,
		externalID: "task-a",
		resultTags: map[string]string{"path_key": "forged"},
		metricNameFunc: func() string {
			return netPathMetricName
		},
	}
	ipt := defaultInput()
	ipt.RegionID = "node-a"
	ipt.setRegionNames("public-node", "")

	d := newDialer(task, ipt)
	d.dialingTime = time.Unix(100, 0)
	d.pointsFeed("http://example.com/v1/write/logging?token=test")

	select {
	case job := <-dialWorker.jobChans:
		require.NotNil(t, job)
		line := job.pt.LineProto()
		assert.Contains(t, line, "path_key="+makeNetPathDialtestingPathKey("node-a", "task-a"))
		assert.Contains(t, line, "source_name=public-node")
		assert.Contains(t, line, "source_host=node-a")
		assert.NotContains(t, line, "path_key=forged")
	default:
		t.Fatal("expected point to be queued")
	}
}
