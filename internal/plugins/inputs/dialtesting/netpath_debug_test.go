// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"net"
	"testing"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeNetPathResolvedIPValidator(t *testing.T) {
	assert.Nil(t, makeNetPathResolvedIPValidator(false, nil))

	validate := makeNetPathResolvedIPValidator(true, nil)
	require.NotNil(t, validate)
	require.NoError(t, validate(net.ParseIP("8.8.8.8")))
	require.Error(t, validate(net.ParseIP("10.0.0.1")))
	require.Error(t, validate(nil))

	validateConfigured := makeNetPathResolvedIPValidator(true, []string{"203.0.113.0/24"})
	require.Error(t, validateConfigured(net.ParseIP("203.0.113.10")))
	require.NoError(t, validateConfigured(net.ParseIP("10.0.0.1")))
}

func TestEnrichNetPathDebugResultMatchesPointMetadata(t *testing.T) {
	ipt := defaultInput()
	ipt.RegionID = "node-a"
	ipt.setRegionNames("hangzhou-node", "")
	ipt.RegionTags = map[string]string{"province": "zhejiang"}
	ipt.Tags = map[string]string{"custom_tag": "custom-value"}
	setNetPathDebugInput(ipt)
	t.Cleanup(func() {
		clearNetPathDebugInput(ipt)
	})

	rawTask, err := dt.NewTask(`{
		"external_id":"task-a",
		"name":"example path",
		"status":"OK",
		"protocol":"tcp",
		"host":"example.com",
		"port":"443",
		"advance_options":{"timeout":"1s"}
	}`, &dt.NetPathTask{})
	require.NoError(t, err)
	task := rawTask.(*dt.NetPathTask)
	tags := map[string]string{"status": "OK"}
	fields := map[string]interface{}{"traceroute": `{"runs":[]}`}

	enrichNetPathDebugResult(task, tags, fields)

	assert.Equal(t, "hangzhou-node", tags["node_name"])
	assert.Equal(t, "hangzhou-node", tags["source_name"])
	assert.Equal(t, "node-a", tags["node_id"])
	assert.Equal(t, "node-a", tags["source_host"])
	assert.Equal(t, makeNetPathDialtestingPathKey("node-a", "task-a"), tags["path_key"])
	assert.Equal(t, "zhejiang", tags["province"])
	assert.Equal(t, "custom-value", tags["custom_tag"])
	assert.Equal(t, "scheduled", tags["trigger_type"])
	assert.Contains(t, tags, "datakit_version")
	assert.Equal(t, int64(1), fields["seq_number"])
	assert.Equal(t, "task-a", fields["task_id"])
}
