// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"encoding/json"
	"strings"
	"testing"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	netpathinput "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netpath"
)

func TestNewNetPathDialTaskFromOnlinePayload(t *testing.T) {
	taskJSON := `{
		"external_id":"netpath-task",
		"name":"example path",
		"status":"OK",
		"frequency":"1m",
		"post_url":"https://openway.example.com?token=tkn_test",
		"protocol":"tcp",
		"host":"example.com",
		"port":"443",
		"advance_options":{
			"timeout":"1s",
			"source_name":"hangzhou",
			"target_name":"example",
			"max_ttl":30,
			"e2e_queries":10,
			"traceroute_queries":3
		},
		"success_when":[{
			"e2e_rtt_avg":[{"op":"lt","target":"500ms"}]
		}],
		"success_when_logic":"and"
	}`

	task, err := (&Input{}).newTaskFromClassJSON(dt.ClassNetPath, taskJSON)
	require.NoError(t, err)
	_, ok := task.(*dt.NetPathTask)
	require.True(t, ok)
	assert.Equal(t, dt.ClassNetPath, task.Class())
	assert.Equal(t, netPathMetricName, task.MetricName())
	assert.Equal(t, "netpath-task", task.GetExternalID())
	assert.Equal(t, []string{"example.com"}, mustNetPathHosts(t, task))
	require.NoError(t, task.RenderTemplateAndInit(nil))
	reasons, success := task.CheckResult()
	assert.False(t, success)
	assert.Empty(t, reasons)
}

func TestNetPathRenderTemplateAndInitValidatesEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		protocol string
		port     string
		wantErr  string
	}{
		{
			name:     "empty host",
			protocol: "icmp",
			wantErr:  "host should not be empty",
		},
		{
			name:     "unsupported protocol",
			host:     "example.com",
			protocol: "http",
			wantErr:  `unsupported protocol "http"`,
		},
		{
			name:     "icmp port",
			host:     "example.com",
			protocol: "icmp",
			port:     "443",
			wantErr:  "port must be omitted for ICMP",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			task := &dt.NetPathTask{
				Task: &dt.Task{
					ExternalID: "netpath-task",
					CurStatus:  "OK",
					Frequency:  "1m",
				},
				Host:     test.host,
				Protocol: test.protocol,
				Port:     test.port,
				SuccessWhen: []*dt.NetPathSuccess{{
					E2EStatus: []*dt.NetPathCondition{{
						Op:     "eq",
						Target: json.RawMessage(`"reached"`),
					}},
				}},
			}
			raw, err := json.Marshal(task)
			require.NoError(t, err)
			parsed, err := dt.NewTask(string(raw), &dt.NetPathTask{})
			require.NoError(t, err)

			err = parsed.RenderTemplateAndInit(nil)
			require.EqualError(t, err, test.wantErr)
		})
	}
}

func TestNetPathRenderTemplateAndInitRejectsNullNumericTarget(t *testing.T) {
	raw := `{
		"external_id":"netpath-task",
		"status":"OK",
		"frequency":"1m",
		"protocol":"icmp",
		"host":"example.com",
		"advance_options":{"timeout":"1s"},
		"success_when":[{"hop_count":[{"op":"eq","target":null}]}]
	}`
	task, err := dt.NewTask(raw, &dt.NetPathTask{})
	require.NoError(t, err)

	err = task.RenderTemplateAndInit(nil)
	require.EqualError(t, err, "hop_count: target must be numeric")
}

func TestNetPathCheckTaskDefersTemplateValidation(t *testing.T) {
	raw := `{
		"external_id":"netpath-task",
		"status":"OK",
		"frequency":"1m",
		"protocol":"{{protocol}}",
		"host":"example.com",
		"port":"{{port}}",
		"advance_options":{"timeout":"1s"},
		"success_when":[{"e2e_status":[{"op":"eq","target":"{{expected_status}}"}]}],
		"config_vars":[
			{"name":"protocol","value":"tcp","secure":false},
			{"name":"port","value":"443","secure":false},
			{"name":"expected_status","value":"reached","secure":false}
		]
	}`
	task, err := dt.NewTask(raw, &dt.NetPathTask{})
	require.NoError(t, err)

	require.NoError(t, task.CheckTask())
	require.NoError(t, task.RenderTemplateAndInit(nil))

	netPathTask := task.(*dt.NetPathTask)
	assert.Equal(t, "tcp", netPathTask.Protocol)
	assert.Equal(t, "443", netPathTask.Port)

	invalidTask, err := dt.NewTask(strings.Replace(raw, `"value":"443"`, `"value":"0"`, 1), &dt.NetPathTask{})
	require.NoError(t, err)
	require.NoError(t, invalidTask.CheckTask())
	require.EqualError(t, invalidTask.RenderTemplateAndInit(nil), "port must be between 1 and 65535")
}

func TestNetPathTaskAssertions(t *testing.T) {
	raw := `{
		"external_id":"netpath-task",
		"name":"example path",
		"status":"OK",
		"frequency":"1m",
		"post_url":"https://openway.example.com?token=tkn_test",
		"protocol":"tcp",
		"host":"example.com",
		"port":"443",
		"advance_options":{"timeout":"1s"},
		"success_when":[{
			"e2e_rtt_avg":[{"op":"lt","target":"500ms"}],
			"e2e_probe_loss_percent":[{"op":"leq","target":5}],
			"e2e_status":[{"op":"eq","target":"reached"}]
		}],
		"success_when_logic":"and"
	}`
	task, err := dt.NewTask(raw, &dt.NetPathTask{})
	require.NoError(t, err)
	netPathTask, ok := task.(*dt.NetPathTask)
	require.True(t, ok)
	executor := newNetPathExecutor(netPathTask)

	reasons, success := executor.evaluate(netpathinput.DialProbeResult{
		Tags: map[string]string{"e2e_status": "reached"},
		Fields: map[string]interface{}{
			"e2e_rtt_avg":            float64(250_000),
			"e2e_probe_loss_percent": float64(0),
		},
	})
	assert.True(t, success)
	assert.Empty(t, reasons)

	reasons, success = executor.evaluate(netpathinput.DialProbeResult{
		Tags:   map[string]string{"e2e_status": "failed"},
		Fields: map[string]interface{}{},
	})
	assert.False(t, success)
	assert.NotEmpty(t, reasons)

	executor.setResult(netpathinput.DialProbeResult{
		Tags: map[string]string{
			"probe_source_ip": "192.0.2.10",
			"probe_dest_ip":   "203.0.113.10",
		},
		Fields: map[string]interface{}{
			"e2e_dest_ip": "203.0.113.11",
		},
	}, nil, true)
	tags, _ := executor.GetResults()
	assert.Equal(t, "example path", tags["name"])
	assert.Equal(t, "example path", tags["task_name"])
	assert.Equal(t, netPathTaskSource, tags["task_source"])
	assert.Equal(t, "example.com", tags["dest_host"])
	assert.Equal(t, "443", tags["dest_port"])
	assert.Equal(t, "192.0.2.10", tags["src_ip"])
	assert.Equal(t, "*", tags["src_port"])
	assert.Equal(t, "192.0.2.10", tags["probe_source_ip"])
	assert.Equal(t, "203.0.113.10", tags["dest_ip"])
	assert.Equal(t, "203.0.113.10", tags["dst_ip"])
	assert.Equal(t, "443", tags["dst_port"])
	assert.Equal(t, "example.com", tags["dst_domain"])
}

func TestNetPathDurationConditionFailureHasUnit(t *testing.T) {
	err := checkNetPathCondition(
		"e2e_rtt_avg",
		&dt.NetPathCondition{Op: "lt", Target: json.RawMessage(`"500ms"`)},
		netpathinput.DialProbeResult{Fields: map[string]interface{}{"e2e_rtt_avg": float64(750_000)}},
	)
	assert.EqualError(t, err, "e2e_rtt_avg: 750ms does not satisfy lt 500ms")
}

func TestNetPathMeasurementAndOneShotClass(t *testing.T) {
	info := (&netPathMeasurement{}).Info()
	require.NotNil(t, info)
	assert.Equal(t, netPathMetricName, info.Name)
	assert.Contains(t, info.Tags, "protocol")
	assert.Contains(t, info.Tags, "path_key")
	assert.Contains(t, info.Tags, "task_name")
	assert.Contains(t, info.Tags, "task_source")
	assert.Contains(t, info.Tags, "dest_host")
	assert.Contains(t, info.Tags, "dest_port")
	assert.Contains(t, info.Tags, "dest_ip")
	assert.Contains(t, info.Tags, "src_ip")
	assert.Contains(t, info.Tags, "src_port")
	assert.Contains(t, info.Tags, "source_host")
	assert.Contains(t, info.Tags, "probe_source_ip")
	assert.Contains(t, info.Tags, "dst_ip")
	assert.Contains(t, info.Tags, "dst_port")
	assert.Contains(t, info.Tags, "dst_domain")
	assert.Contains(t, info.Fields, "traceroute")
	assert.Contains(t, info.Fields, "task_id")
	assert.Equal(t, inputs.DurationUS, info.Fields["duration"].(*inputs.FieldInfo).Unit)
	assert.Equal(t, inputs.DurationUS, info.Fields["e2e_rtt_avg"].(*inputs.FieldInfo).Unit)
	assert.Equal(t, inputs.Percent, info.Fields["e2e_probe_loss_percent"].(*inputs.FieldInfo).Unit)
	assert.Equal(t, inputs.NCount, info.Fields["hop_count"].(*inputs.FieldInfo).Unit)
	assert.Equal(t, dt.ClassNetPath, oneShotMetricProtocol(dt.ClassNetPath))
}

func TestNetPathTaskIPPortDisplayName(t *testing.T) {
	raw := `{
		"external_id":"netpath-task",
		"name":"192.0.2.10",
		"status":"OK",
		"frequency":"1m",
		"protocol":"tcp",
		"host":"192.0.2.10",
		"port":"443",
		"advance_options":{"timeout":"1s"},
		"success_when":[{"e2e_status":[{"op":"eq","target":"reached"}]}]
	}`
	task, err := dt.NewTask(raw, &dt.NetPathTask{})
	require.NoError(t, err)
	executor := newNetPathExecutor(task.(*dt.NetPathTask))
	executor.setResult(netpathinput.DialProbeResult{
		Tags:   map[string]string{},
		Fields: map[string]interface{}{},
	}, nil, true)

	tags, _ := executor.GetResults()
	assert.Equal(t, "192.0.2.10", tags["name"])
	assert.Equal(t, "192.0.2.10:443", tags["task_name"])
	assert.Equal(t, "192.0.2.10", tags["dest_host"])
	assert.Equal(t, "443", tags["dest_port"])
	assert.Equal(t, "192.0.2.10", tags["dest_ip"])
	assert.Equal(t, "192.0.2.10", tags["dst_ip"])
	assert.Equal(t, "443", tags["dst_port"])
	assert.NotContains(t, tags, "dst_domain")
}

func mustNetPathHosts(t *testing.T, task interface {
	GetHostName() ([]string, error)
}) []string {
	t.Helper()
	hosts, err := task.GetHostName()
	require.NoError(t, err)
	return hosts
}

func TestNetPathTaskSanitizesSecrets(t *testing.T) {
	raw := `{
		"external_id":"netpath-task",
		"status":"stop",
		"frequency":"1m",
		"access_key":"secret-ak",
		"post_url":"https://openway.example.com?token=secret-token",
		"protocol":"icmp",
		"host":"example.com",
		"advance_options":{"timeout":"1s"},
		"success_when":[{"e2e_status":[{"op":"eq","target":"reached"}]}],
		"config_vars":[{"name":"password","value":"secret-value","secure":true}]
	}`
	taskRaw, err := dt.NewTask(raw, &dt.NetPathTask{})
	require.NoError(t, err)
	task := taskRaw.(*dt.NetPathTask)
	executor := newNetPathExecutor(task)

	var sanitized map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(executor.sanitizedTask()), &sanitized))
	assert.NotContains(t, sanitized, "access_key")
	assert.NotContains(t, sanitized, "post_url")
	variables := sanitized["config_vars"].([]interface{})
	assert.NotContains(t, variables[0].(map[string]interface{}), "value")
}
