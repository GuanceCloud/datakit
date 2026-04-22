// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestInternalNetwork(t *testing.T) {
	dialWorker = &worker{
		sender: &mockSender{},
	}
	ipt := &Input{
		DisableInternalNetworkTask: true,
	}

	child := &dialtesting.HTTPTask{
		Method: "GET",
		Task: &dialtesting.Task{
			Name:      "test",
			Frequency: "1ms",
		},
		URL: "http://127.0.0.1:9529",
		SuccessWhen: []*dialtesting.HTTPSuccess{
			{
				StatusCode: []*dialtesting.SuccessOption{
					{
						Is: "200",
					},
				},
			},
		},
	}

	task, err := dialtesting.NewTask("", child)
	assert.NoError(t, err)

	dialer := newDialer(task, ipt)
	assert.Error(t, dialer.run())

	child = &dialtesting.HTTPTask{
		Method: "GET",
		Task: &dialtesting.Task{
			Name:      "test",
			PostURL:   "http://xxxxx?token=xxxxxx",
			Frequency: "1ms",
		},
		URL: "http://8.8.8.8",
		AdvanceOptions: &dialtesting.HTTPAdvanceOption{
			RequestTimeout: "1s",
		},
		SuccessWhen: []*dialtesting.HTTPSuccess{
			{
				StatusCode: []*dialtesting.SuccessOption{
					{
						Is: "200",
					},
				},
			},
		},
	}

	task, err = dialtesting.NewTask("", child)

	assert.NoError(t, err)

	dialer = newDialer(task, ipt)
	go func() {
		time.Sleep(100 * time.Millisecond)
		task.SetStatus(dialtesting.StatusStop)
		dialer.updateCh <- task
	}()
	assert.NoError(t, dialer.run())
}

func TestCheckInternalNetwork(t *testing.T) {
	newTask := func(t *testing.T, host string) dialtesting.ITask {
		t.Helper()

		task, err := dialtesting.NewTask("", &dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				Name:      "test",
				Frequency: "1s",
			},
			URL: host,
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)
		return task
	}

	t.Run("allow when internal network check disabled", func(t *testing.T) {
		ipt := defaultInput()
		ipt.DisableInternalNetworkTask = false

		task := newTask(t, "http://localhost")
		d := newDialer(task, ipt)

		assert.NoError(t, d.checkInternalNetwork())
		assert.NoError(t, task.RenderTemplateAndInit(nil))
		assert.NoError(t, task.Run())
	})

	t.Run("nil task before run is ignored", func(t *testing.T) {
		ipt := defaultInput()
		task := newTask(t, "http://localhost")
		d := newDialer(task, ipt)

		assert.NoError(t, d.checkInternalNetwork())
		assert.NotPanics(t, func() {
			task.SetBeforeRun(nil)
			assert.NoError(t, d.checkInternalNetwork())
		})
	})

	t.Run("return error on invalid host", func(t *testing.T) {
		ipt := defaultInput()
		ipt.DisableInternalNetworkTask = true

		task := newTask(t, "http://①0.43.239.255:5000")
		d := newDialer(task, ipt)

		assert.NoError(t, d.checkInternalNetwork())
		assert.NoError(t, task.RenderTemplateAndInit(nil))
		assert.NoError(t, task.Run())

		_, fields := task.GetResults()
		assert.Contains(t, fields["fail_reason"], "dest host is not valid")
	})

	t.Run("deny internal host", func(t *testing.T) {
		ipt := defaultInput()
		ipt.DisableInternalNetworkTask = true

		task := newTask(t, "http://127.0.0.1")
		d := newDialer(task, ipt)

		assert.NoError(t, d.checkInternalNetwork())
		assert.NoError(t, task.RenderTemplateAndInit(nil))
		assert.NoError(t, task.Run())

		_, fields := task.GetResults()
		assert.Contains(t, fields["fail_reason"], "is not allowed to be tested")
	})

	t.Run("empty host list is allowed", func(t *testing.T) {
		ipt := defaultInput()
		ipt.DisableInternalNetworkTask = true
		task := &headlessTaskStub{}
		d := newDialer(task, ipt)

		assert.NoError(t, d.checkInternalNetwork())
	})
}

func TestPopulateDFLabelTags(t *testing.T) {
	cases := []struct {
		Title  string
		Label  string
		Expect map[string]string
	}{
		{
			Title:  "no need to extract tags",
			Label:  "test",
			Expect: map[string]string{LabelDF: `["test"]`},
		},
		{
			Title:  "empty label",
			Label:  "",
			Expect: map[string]string{LabelDF: `[]`},
		},
		{
			Title:  "extract tags",
			Label:  "test,f1:2,f2:3:3",
			Expect: map[string]string{LabelDF: `["test","f1:2","f2:3:3"]`, "f1": "2", "f2": "3:3"},
		},
		{
			Title:  "new label format",
			Label:  "[\"tag1:value1\",\"tag2:value2\",\"tag3:value3\"]",
			Expect: map[string]string{LabelDF: "[\"tag1:value1\",\"tag2:value2\",\"tag3:value3\"]", "tag1": "value1", "tag2": "value2", "tag3": "value3"},
		},
	}
	for _, tc := range cases {
		tags := make(map[string]string)
		populateDFLabelTags(tc.Label, tags)

		assert.EqualValues(t, tc.Expect, tags)
	}
}

func TestDispatchTasks(t *testing.T) {
	cases := []struct {
		name    string
		payload []byte
	}{
		{
			name: "invalid region info",
			payload: []byte(`{
				"content": {
					"region": "invalid-region-info"
				}
			}`),
		},
		{
			name: "invalid variables info",
			payload: []byte(`{
					"content": {
						"variables": 123
					}
				}`),
		},
		{
			name: "invalid variables json string",
			payload: []byte(`{
					"content": {
						"variables": "{invalid-json}"
					}
				}`),
		},
		{
			name: "invalid task data",
			payload: []byte(`{
				"content": {
					"HTTP": [123]
				}
			}`),
		},
		{
			name: "unknown task type",
			payload: []byte(`{
				"content": {
					"UNKNOWN": ["{}"]
				}
			}`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := defaultInput()

			assert.NotPanics(t, func() {
				err := ipt.dispatchTasks(tc.payload)
				assert.NoError(t, err)
			})
		})
	}

	t.Run("stop status task should not create dialer", func(t *testing.T) {
		ipt := defaultInput()

		taskJSON, err := json.Marshal(&dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID: "stop-task",
				Name:       "stop-task",
				PostURL:    "http://example.com?token=test",
				CurStatus:  dialtesting.StatusStop,
				Frequency:  "1s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				dialtesting.ClassHTTP: []string{string(taskJSON)},
			},
		})
		assert.NoError(t, err)

		assert.NotPanics(t, func() {
			err := ipt.dispatchTasks(payload)
			assert.NoError(t, err)
		})

		found := false
		ipt.curTasks.Range(func(key, value any) bool {
			found = true
			return false
		})
		assert.False(t, found)
	})

	t.Run("headless tasks are ignored", func(t *testing.T) {
		ipt := defaultInput()

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				dialtesting.ClassHeadless: []string{`{"name":"browser-task"}`},
			},
		})
		assert.NoError(t, err)

		assert.NoError(t, ipt.dispatchTasks(payload))

		found := false
		ipt.curTasks.Range(func(key, value any) bool {
			found = true
			return false
		})
		assert.False(t, found)
	})

	t.Run("deprecated dns and other tasks are ignored", func(t *testing.T) {
		ipt := defaultInput()

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				dialtesting.ClassDNS:   []string{`{"name":"dns-task"}`},
				dialtesting.ClassOther: []string{`{"name":"other-task"}`},
			},
		})
		assert.NoError(t, err)

		assert.NoError(t, ipt.dispatchTasks(payload))

		found := false
		ipt.curTasks.Range(func(key, value any) bool {
			found = true
			return false
		})
		assert.False(t, found)
	})

	t.Run("invalid class payload type is ignored", func(t *testing.T) {
		ipt := defaultInput()

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				dialtesting.ClassHTTP: map[string]interface{}{"task": "invalid"},
			},
		})
		assert.NoError(t, err)

		assert.NoError(t, ipt.dispatchTasks(payload))

		found := false
		ipt.curTasks.Range(func(key, value any) bool {
			found = true
			return false
		})
		assert.False(t, found)
	})

	t.Run("invalid task json string is ignored", func(t *testing.T) {
		ipt := defaultInput()

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				dialtesting.ClassHTTP: []string{`{invalid-json}`},
			},
		})
		assert.NoError(t, err)

		assert.NoError(t, ipt.dispatchTasks(payload))

		found := false
		ipt.curTasks.Range(func(key, value any) bool {
			found = true
			return false
		})
		assert.False(t, found)
	})

	t.Run("region info ignores unsupported value types", func(t *testing.T) {
		ipt := defaultInput()

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				RegionInfo: map[string]interface{}{
					"name":     "cn-sh",
					"priority": 1,
					"meta": map[string]interface{}{
						"nested": true,
					},
					"internal": true,
				},
			},
		})
		assert.NoError(t, err)

		assert.NoError(t, ipt.dispatchTasks(payload))
		assert.Equal(t, "cn-sh", ipt.regionName)
		assert.Equal(t, "true", ipt.RegionTags["internal"])
		_, hasPriority := ipt.RegionTags["priority"]
		_, hasMeta := ipt.RegionTags["meta"]
		assert.False(t, hasPriority)
		assert.False(t, hasMeta)
	})

	t.Run("apply region and variables info", func(t *testing.T) {
		ipt := defaultInput()

		vars, err := json.Marshal([]dialtesting.Variable{
			{
				UUID:            "var-1",
				Value:           "value-1",
				TaskID:          "task-1",
				OwnerExternalID: "owner-1",
				UpdatedAt:       123,
			},
		})
		assert.NoError(t, err)

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				RegionInfo: map[string]interface{}{
					"name":      "shanghai",
					"name_en":   "shanghai-en",
					"status":    "online",
					"isp":       "telecom",
					"internal":  true,
					"available": false,
				},
				VariablesInfo: string(vars),
			},
		})
		assert.NoError(t, err)

		assert.NoError(t, ipt.dispatchTasks(payload))
		assert.Equal(t, "shanghai", ipt.regionName)
		assert.Equal(t, "shanghai-en", ipt.regionNameEn)
		assert.Equal(t, "telecom", ipt.RegionTags["isp"])
		assert.Equal(t, "true", ipt.RegionTags["internal"])
		assert.Equal(t, "false", ipt.RegionTags["available"])
		_, gotVars := ipt.variables.getVariables([]string{"var-1"})
		if assert.Contains(t, gotVars, "var-1") {
			assert.Equal(t, "value-1", gotVars["var-1"].Value)
		}
	})

	t.Run("create new task stores dialer", func(t *testing.T) {
		ipt := defaultInput()

		taskJSON, err := json.Marshal(&dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID: "new-task",
				Name:       "new-task",
				PostURL:    "http://example.com?token=test",
				Frequency:  "1s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				dialtesting.ClassHTTP: []string{string(taskJSON)},
			},
		})
		assert.NoError(t, err)

		assert.NoError(t, ipt.dispatchTasks(payload))

		found := false
		ipt.curTasks.Range(func(key, value any) bool {
			d, ok := value.(*dialer)
			if ok && d != nil && d.task != nil && d.task.GetExternalID() == "new-task" {
				found = true
				return false
			}
			return true
		})
		assert.True(t, found)

		ipt.semStop.Close()
		time.Sleep(20 * time.Millisecond)
	})

	t.Run("task update time advances input position", func(t *testing.T) {
		ipt := defaultInput()
		ipt.pos = 10

		taskJSON, err := json.Marshal(&dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID: "pos-task",
				Name:       "pos-task",
				PostURL:    "http://example.com?token=test",
				Frequency:  "1s",
				UpdateTime: 123,
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				dialtesting.ClassHTTP: []string{string(taskJSON)},
			},
		})
		assert.NoError(t, err)

		assert.NoError(t, ipt.dispatchTasks(payload))
		assert.EqualValues(t, 123, ipt.pos)

		ipt.semStop.Close()
		time.Sleep(20 * time.Millisecond)
	})

	t.Run("update existing task", func(t *testing.T) {
		ipt := defaultInput()
		existing, err := dialtesting.NewTask("", &dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID: "existing-task",
				Name:       "existing-task",
				PostURL:    "http://example.com?token=test",
				Frequency:  "1s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		d := newDialer(existing, ipt)
		ipt.curTasks.Store(existing.ID(), d)

		taskJSON, err := json.Marshal(&dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID: "existing-task",
				Name:       "existing-task",
				PostURL:    "http://example.com?token=test",
				Frequency:  "2s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				dialtesting.ClassHTTP: []string{string(taskJSON)},
			},
		})
		assert.NoError(t, err)

		assert.NoError(t, ipt.dispatchTasks(payload))

		select {
		case got := <-d.updateCh:
			assert.Equal(t, "2s", got.GetFrequency())
		case <-time.After(time.Second):
			t.Fatal("expected task update to be pushed to existing dialer")
		}
	})

	t.Run("stop update removes existing task", func(t *testing.T) {
		ipt := defaultInput()

		existing, err := dialtesting.NewTask("", &dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID: "stop-existing-task",
				Name:       "stop-existing-task",
				PostURL:    "http://example.com?token=test",
				Frequency:  "1s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		d := newDialer(existing, ipt)
		ipt.curTasks.Store(existing.ID(), d)

		taskJSON, err := json.Marshal(&dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID: "stop-existing-task",
				Name:       "stop-existing-task",
				PostURL:    "http://example.com?token=test",
				CurStatus:  dialtesting.StatusStop,
				Frequency:  "1s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				dialtesting.ClassHTTP: []string{string(taskJSON)},
			},
		})
		assert.NoError(t, err)

		assert.NoError(t, ipt.dispatchTasks(payload))
		_, ok := ipt.curTasks.Load(existing.ID())
		assert.False(t, ok)
	})

	t.Run("task with too many failures is removed instead of updated", func(t *testing.T) {
		ipt := defaultInput()

		existing, err := dialtesting.NewTask("", &dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID: "failed-task",
				Name:       "failed-task",
				PostURL:    "http://example.com?token=test",
				Frequency:  "1s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		d := newDialer(existing, ipt)
		d.failCnt = MaxFails
		ipt.curTasks.Store(existing.ID(), d)

		taskJSON, err := json.Marshal(&dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID: "failed-task",
				Name:       "failed-task",
				PostURL:    "http://example.com?token=test",
				Frequency:  "2s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				dialtesting.ClassHTTP: []string{string(taskJSON)},
			},
		})
		assert.NoError(t, err)

		assert.NoError(t, ipt.dispatchTasks(payload))

		_, ok := ipt.curTasks.Load(existing.ID())
		assert.False(t, ok)
	})

	t.Run("update task error keeps existing dialer", func(t *testing.T) {
		ipt := defaultInput()

		existing, err := dialtesting.NewTask("", &dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID: "closed-task",
				Name:       "closed-task",
				PostURL:    "http://example.com?token=test",
				Frequency:  "1s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		d := newDialer(existing, ipt)
		close(d.stopCh)
		ipt.curTasks.Store(existing.ID(), d)

		taskJSON, err := json.Marshal(&dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID: "closed-task",
				Name:       "closed-task",
				PostURL:    "http://example.com?token=test",
				Frequency:  "2s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		payload, err := json.Marshal(map[string]interface{}{
			"content": map[string]interface{}{
				dialtesting.ClassHTTP: []string{string(taskJSON)},
			},
		})
		assert.NoError(t, err)

		assert.NoError(t, ipt.dispatchTasks(payload))

		value, ok := ipt.curTasks.Load(existing.ID())
		assert.True(t, ok)
		assert.Same(t, d, value)
	})
}

func TestReadEnv(t *testing.T) {
	t.Run("read valid envs", func(t *testing.T) {
		ipt := defaultInput()

		ipt.ReadEnv(map[string]string{
			"ENV_INPUT_DIALTESTING_AK":                                  "test-ak",
			"ENV_INPUT_DIALTESTING_SK":                                  "test-sk",
			"ENV_INPUT_DIALTESTING_REGION_ID":                           "test-region",
			"ENV_INPUT_DIALTESTING_SERVER":                              "http://example.com",
			"ENV_INPUT_DIALTESTING_ELECTION":                            "true",
			"ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK":       "true",
			"ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST": `["10.0.0.0/8","192.168.0.0/16"]`,
		})

		assert.Equal(t, "test-ak", ipt.AK)
		assert.Equal(t, "test-sk", ipt.SK)
		assert.Equal(t, "test-region", ipt.RegionID)
		assert.Equal(t, "http://example.com", ipt.Server)
		assert.True(t, ipt.Election)
		assert.True(t, ipt.DisableInternalNetworkTask)
		assert.Equal(t, []string{"10.0.0.0/8", "192.168.0.0/16"}, ipt.DisabledInternalNetworkCIDRList)
	})

	t.Run("invalid boolean env keeps defaults", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Election = false
		ipt.DisableInternalNetworkTask = false

		ipt.ReadEnv(map[string]string{
			"ENV_INPUT_DIALTESTING_ELECTION":                      "not-bool",
			"ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK": "not-bool",
		})

		assert.False(t, ipt.Election)
		assert.False(t, ipt.DisableInternalNetworkTask)
	})

	t.Run("invalid cidr list json is ignored", func(t *testing.T) {
		ipt := defaultInput()
		ipt.DisabledInternalNetworkCIDRList = nil

		ipt.ReadEnv(map[string]string{
			"ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK":       "true",
			"ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST": `invalid-json`,
		})

		assert.True(t, ipt.DisableInternalNetworkTask)
		assert.Nil(t, ipt.DisabledInternalNetworkCIDRList)
	})

	t.Run("cidr list is ignored when disable switch is false", func(t *testing.T) {
		ipt := &Input{}
		ipt.DisabledInternalNetworkCIDRList = nil

		ipt.ReadEnv(map[string]string{
			"ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK":       "false",
			"ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST": `["10.0.0.0/8"]`,
		})

		assert.False(t, ipt.DisableInternalNetworkTask)
		assert.Nil(t, ipt.DisabledInternalNetworkCIDRList)
	})
}

func TestVariable(t *testing.T) {
	t.Run("set and get variables", func(t *testing.T) {
		v := Variable{
			data:     map[string]dialtesting.Variable{},
			taskData: map[string]map[string]dialtesting.Variable{},
		}

		v.setVariables([]dialtesting.Variable{
			{
				UUID:            "var-1",
				Value:           "value-1",
				Secure:          true,
				UpdatedAt:       100,
				TaskID:          "task-1",
				OwnerExternalID: "owner-1",
			},
			{
				UUID:            "var-2",
				Value:           "value-2",
				UpdatedAt:       200,
				TaskID:          "task-1",
				OwnerExternalID: "owner-1",
			},
		})

		assert.EqualValues(t, 200, v.getLatestPos())

		pos, vars := v.getVariables([]string{"var-1", "var-2", "not-exist"})
		assert.EqualValues(t, 200, pos)
		assert.Len(t, vars, 2)
		assert.Equal(t, "value-1", vars["var-1"].Value)
		assert.True(t, vars["var-1"].Secure)
		assert.Equal(t, "value-2", vars["var-2"].Value)
	})

	t.Run("get variables by task", func(t *testing.T) {
		v := Variable{
			data:     map[string]dialtesting.Variable{},
			taskData: map[string]map[string]dialtesting.Variable{},
		}

		task, err := dialtesting.NewTask("", &dialtesting.HTTPTask{
			Task: &dialtesting.Task{
				ExternalID:      "task-1",
				OwnerExternalID: "owner-1",
				Frequency:       "1s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		v.setVariables([]dialtesting.Variable{
			{
				UUID:            "var-1",
				Value:           "value-1",
				TaskID:          "task-1",
				OwnerExternalID: "owner-1",
			},
			{
				UUID:            "var-2",
				Value:           "value-2",
				TaskID:          "task-2",
				OwnerExternalID: "owner-1",
			},
		})

		vars := v.getVariablesByTask(task)
		assert.Len(t, vars, 1)
		assert.Equal(t, "value-1", vars["var-1"].Value)
	})

	t.Run("deleted variable is removed from task data", func(t *testing.T) {
		v := Variable{
			data:     map[string]dialtesting.Variable{},
			taskData: map[string]map[string]dialtesting.Variable{},
		}

		v.setVariables([]dialtesting.Variable{
			{
				UUID:            "var-1",
				Value:           "value-1",
				TaskID:          "task-1",
				OwnerExternalID: "owner-1",
			},
		})

		v.setVariables([]dialtesting.Variable{
			{
				UUID:            "var-1",
				TaskID:          "task-1",
				OwnerExternalID: "owner-1",
				DeletedAt:       1,
			},
		})

		assert.Empty(t, v.taskData[v.getTaskKey("owner-1", "task-1")])
	})

	t.Run("get task key", func(t *testing.T) {
		v := Variable{}
		assert.Equal(t, "owner-task", v.getTaskKey("owner", "task"))
	})

	t.Run("update variable value enqueues latest variable", func(t *testing.T) {
		v := Variable{
			updateVariableCh: make(chan dialtesting.Variable, 1),
		}

		variable := dialtesting.Variable{
			UUID:            "var-1",
			OwnerExternalID: "owner-1",
		}

		v.updateVariableValue(variable, "new-value", 2)

		select {
		case got := <-v.updateVariableCh:
			assert.Equal(t, "var-1", got.UUID)
			assert.Equal(t, "owner-1", got.OwnerExternalID)
			assert.Equal(t, "new-value", got.Value)
			assert.Equal(t, 2, got.FailCount)
			assert.NotZero(t, got.UpdatedAt)
		case <-time.After(time.Second):
			t.Fatal("expected variable update to be enqueued")
		}
	})

	t.Run("get variables with empty input", func(t *testing.T) {
		v := Variable{
			data:     map[string]dialtesting.Variable{},
			taskData: map[string]map[string]dialtesting.Variable{},
		}

		pos, vars := v.getVariables(nil)
		assert.EqualValues(t, 0, pos)
		assert.Empty(t, vars)
	})

	t.Run("get variables by task returns empty for unmatched owner", func(t *testing.T) {
		v := Variable{
			data:     map[string]dialtesting.Variable{},
			taskData: map[string]map[string]dialtesting.Variable{},
		}

		v.setVariables([]dialtesting.Variable{
			{
				UUID:            "var-1",
				Value:           "value-1",
				TaskID:          "task-1",
				OwnerExternalID: "owner-1",
			},
		})

		task, err := dialtesting.NewTask("", &dialtesting.HTTPTask{
			Task: &dialtesting.Task{
				ExternalID:      "task-1",
				OwnerExternalID: "owner-2",
				Frequency:       "1s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		assert.Empty(t, v.getVariablesByTask(task))
	})
}

func TestInputHelpers(t *testing.T) {
	t.Run("default input defaults", func(t *testing.T) {
		ipt := defaultInput()
		assert.NotNil(t, ipt.Tags)
		assert.NotNil(t, ipt.RegionTags)
		assert.NotNil(t, ipt.semStop)
		assert.Equal(t, 1000, ipt.MaxJobChanNumber)
		assert.Equal(t, 10000, ipt.MaxCachePointsNumber)
		assert.Equal(t, 10, ipt.MaxJobNumber)
		assert.True(t, ipt.DisableInternalNetworkTask)
	})

	t.Run("sample config and catalog", func(t *testing.T) {
		ipt := defaultInput()
		assert.Contains(t, ipt.SampleConfig(), "[[inputs.dialtesting]]")
		assert.Equal(t, "network", ipt.Catalog())
	})

	t.Run("available archs", func(t *testing.T) {
		ipt := defaultInput()
		assert.NotEmpty(t, ipt.AvailableArchs())
	})

	t.Run("sample measurement", func(t *testing.T) {
		ipt := defaultInput()
		ms := ipt.SampleMeasurement()
		assert.NotEmpty(t, ms)
		assert.Len(t, ms, 6)
	})

	t.Run("election enabled", func(t *testing.T) {
		ipt := defaultInput()
		assert.False(t, ipt.ElectionEnabled())

		ipt.Election = true
		assert.True(t, ipt.ElectionEnabled())
	})

	t.Run("pause and resume", func(t *testing.T) {
		ipt := defaultInput()

		assert.NoError(t, ipt.Pause())
		assert.True(t, ipt.pause.Load())

		assert.NoError(t, ipt.Resume())
		assert.False(t, ipt.pause.Load())
	})

	t.Run("terminate closes sem stop", func(t *testing.T) {
		ipt := defaultInput()
		sem := ipt.semStop

		ipt.Terminate()

		select {
		case <-sem.Wait():
		case <-time.After(time.Second):
			t.Fatal("expected sem stop to be closed")
		}
	})

	t.Run("setup worker in debug mode initializes once", func(t *testing.T) {
		oldWorker := dialWorker
		defer func() {
			dialWorker = oldWorker
			once = sync.Once{}
		}()

		dialWorker = nil
		once = sync.Once{}

		ipt := defaultInput()
		ipt.isDebugMode = true
		ipt.MaxJobNumber = 3
		ipt.MaxJobChanNumber = 4
		ipt.MaxCachePointsNumber = 5
		ipt.setupWorker()

		if assert.NotNil(t, dialWorker) {
			assert.IsType(t, &emptySender{}, dialWorker.sender)
			assert.Equal(t, 3, dialWorker.maxJobNumber)
			assert.Equal(t, 4, dialWorker.maxJobChanNumber)
			assert.Equal(t, 5, dialWorker.maxCachePointsNumber)
		}

		firstWorker := dialWorker
		ipt2 := defaultInput()
		ipt2.isDebugMode = true
		ipt2.MaxJobNumber = 99
		ipt2.setupWorker()

		assert.Same(t, firstWorker, dialWorker)
		assert.Equal(t, 3, dialWorker.maxJobNumber)
	})

	t.Run("setup worker in non debug mode initializes dw sender", func(t *testing.T) {
		oldWorker := dialWorker
		defer func() {
			dialWorker = oldWorker
			once = sync.Once{}
		}()

		dialWorker = nil
		once = sync.Once{}

		ipt := defaultInput()
		ipt.isDebugMode = false
		ipt.MaxJobNumber = 2
		ipt.MaxJobChanNumber = 3
		ipt.MaxCachePointsNumber = 4
		ipt.setupCli()
		ipt.setupWorker()

		if assert.NotNil(t, dialWorker) {
			assert.IsType(t, &dwSender{}, dialWorker.sender)
			assert.Equal(t, 2, dialWorker.maxJobNumber)
			assert.Equal(t, 3, dialWorker.maxJobChanNumber)
			assert.Equal(t, 4, dialWorker.maxCachePointsNumber)
		}
	})
}

func TestSignReq(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.NoError(t, err)

	req.Header.Set("Date", time.Now().Format(http.TimeFormat))
	req.Header.Set("Content-MD5", "d41d8cd98f00b204e9800998ecf8427e")

	assert.NotPanics(t, func() {
		signReq(req, "test-ak", "test-sk")
	})

	auth := req.Header.Get("Authorization")
	assert.Contains(t, auth, "DIAL_TESTING test-ak:")
	assert.NotEmpty(t, auth)

	req.Header.Set("Authorization", "legacy-auth")
	signReq(req, "test-ak", "test-sk")
	assert.Contains(t, req.Header.Get("Authorization"), "DIAL_TESTING test-ak:")
	assert.NotContains(t, req.Header.Get("Authorization"), "legacy-auth")
}

func TestPullTask(t *testing.T) {
	t.Run("invalid server url", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Server = "://invalid-url"

		b, err := ipt.pullTask()
		assert.Error(t, err)
		assert.Nil(t, b)
	})

	t.Run("server error returns immediately with current retry logic", func(t *testing.T) {
		var reqCount int
		ipt := defaultInput()
		ipt.Server = "http://example.com"
		ipt.RegionID = "test-region"
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				reqCount++
				if reqCount < 3 {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Status:     "500 Internal Server Error",
						Body:       io.NopCloser(strings.NewReader(`server error`)),
						Header:     make(http.Header),
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(strings.NewReader(`{"content":{}}`)),
					Header:     make(http.Header),
				}, nil
			}),
		}

		b, err := ipt.pullTask()
		assert.Error(t, err)
		assert.Nil(t, b)
		assert.Equal(t, 1, reqCount)
	})

	t.Run("stop retry on non server error", func(t *testing.T) {
		var reqCount int
		ipt := defaultInput()
		ipt.Server = "http://example.com"
		ipt.RegionID = "test-region"
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				reqCount++
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`bad request`)),
					Header:     make(http.Header),
				}, nil
			}),
		}

		b, err := ipt.pullTask()
		assert.Error(t, err)
		assert.Nil(t, b)
		assert.Equal(t, 1, reqCount)
	})
}

func TestPullHTTPTask(t *testing.T) {
	t.Run("success response", func(t *testing.T) {
		reqURL, err := url.Parse("http://example.com")
		assert.NoError(t, err)

		var gotPath, gotRegion, gotSince, gotVariableSince, gotAuth, gotDate, gotMD5, gotConn string
		ipt := defaultInput()
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.RegionID = "test-region"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotPath = r.URL.Path
				gotRegion = r.URL.Query().Get("region_id")
				gotSince = r.URL.Query().Get("since")
				gotVariableSince = r.URL.Query().Get("variable_since")
				gotAuth = r.Header.Get("Authorization")
				gotDate = r.Header.Get("Date")
				gotMD5 = r.Header.Get("Content-MD5")
				gotConn = r.Header.Get("Connection")
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(strings.NewReader(`{"content":{}}`)),
					Header:     make(http.Header),
				}, nil
			}),
		}

		body, statusCode, err := ipt.pullHTTPTask(reqURL, 10, 20)
		assert.NoError(t, err)
		assert.Equal(t, 2, statusCode)
		assert.JSONEq(t, `{"content":{}}`, string(body))
		assert.Equal(t, "/v1/task/pull", gotPath)
		assert.Equal(t, "test-region", gotRegion)
		assert.Equal(t, "10", gotSince)
		assert.Equal(t, "20", gotVariableSince)
		assert.Contains(t, gotAuth, "DIAL_TESTING test-ak:")
		assert.NotEmpty(t, gotDate)
		assert.NotEmpty(t, gotMD5)
		assert.Equal(t, "close", gotConn)
	})

	t.Run("server failure stops all tasks on region disabled", func(t *testing.T) {
		reqURL, err := url.Parse("http://example.com")
		assert.NoError(t, err)

		ipt := defaultInput()
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.RegionID = "test-region"
		ipt.Server = "http://example.com"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`kodo.RegionNotFoundOrDisabled`)),
					Header:     make(http.Header),
				}, nil
			}),
		}

		task := newHTTPDialtestingTask(t, "pull-http-task")
		d := newDialer(task, ipt)
		ipt.curTasks.Store("task", d)

		body, statusCode, err := ipt.pullHTTPTask(reqURL, 0, 0)
		assert.Error(t, err)
		assert.Nil(t, body)
		assert.Equal(t, 4, statusCode)

		found := false
		ipt.curTasks.Range(func(key, value any) bool {
			found = true
			return false
		})
		assert.False(t, found)
	})

	t.Run("non region-disabled failure keeps tasks", func(t *testing.T) {
		reqURL, err := url.Parse("http://example.com")
		assert.NoError(t, err)

		ipt := defaultInput()
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.RegionID = "test-region"
		ipt.Server = "http://example.com"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`bad request`)),
					Header:     make(http.Header),
				}, nil
			}),
		}

		task := newHTTPDialtestingTask(t, "keep-task")
		d := newDialer(task, ipt)
		ipt.curTasks.Store("task", d)

		body, statusCode, err := ipt.pullHTTPTask(reqURL, 0, 0)
		assert.Error(t, err)
		assert.Nil(t, body)
		assert.Equal(t, 4, statusCode)

		_, ok := ipt.curTasks.Load("task")
		assert.True(t, ok)
	})

	t.Run("client do error", func(t *testing.T) {
		reqURL, err := url.Parse("http://example.com")
		assert.NoError(t, err)

		ipt := defaultInput()
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.RegionID = "test-region"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("do failed")
			}),
		}

		body, statusCode, err := ipt.pullHTTPTask(reqURL, 0, 0)
		assert.Error(t, err)
		assert.Nil(t, body)
		assert.Equal(t, 5, statusCode)
	})

	t.Run("read body error", func(t *testing.T) {
		reqURL, err := url.Parse("http://example.com")
		assert.NoError(t, err)

		ipt := defaultInput()
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.RegionID = "test-region"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       &errReadCloser{},
					Header:     make(http.Header),
				}, nil
			}),
		}

		body, statusCode, err := ipt.pullHTTPTask(reqURL, 0, 0)
		assert.Error(t, err)
		assert.Nil(t, body)
		assert.Equal(t, 0, statusCode)
	})

	t.Run("authorization header is set on request", func(t *testing.T) {
		reqURL, err := url.Parse("http://example.com")
		assert.NoError(t, err)

		var authHeader string
		ipt := defaultInput()
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.RegionID = "test-region"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				authHeader = req.Header.Get("Authorization")
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(strings.NewReader(`{"content":{}}`)),
					Header:     make(http.Header),
				}, nil
			}),
		}

		_, _, err = ipt.pullHTTPTask(reqURL, 0, 0)
		assert.NoError(t, err)
		assert.Contains(t, authHeader, "DIAL_TESTING test-ak:")
	})
}

func TestUpdateRemoteVariables(t *testing.T) {
	t.Run("return early when reqURL is nil", func(t *testing.T) {
		v := &Variable{
			updateVariables: []dialtesting.Variable{
				{UUID: "var-1", Value: "value-1"},
			},
			ipt: defaultInput(),
		}

		v.updateRemoteVariables()
		assert.Empty(t, v.updateVariables)
	})

	t.Run("post variables to remote server", func(t *testing.T) {
		var got []dialtesting.Variable
		reqURL, err := url.Parse("http://example.com")
		assert.NoError(t, err)

		ipt := defaultInput()
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.Header.Get("Authorization"), "DIAL_TESTING test-ak:")
				defer r.Body.Close()
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&got))
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(strings.NewReader(`ok`)),
					Header:     make(http.Header),
				}, nil
			}),
		}

		v := &Variable{
			updateVariables: []dialtesting.Variable{
				{
					UUID:            "var-1",
					Value:           "value-1",
					OwnerExternalID: "owner-1",
				},
			},
			reqURL: reqURL,
			ipt:    ipt,
		}

		v.updateRemoteVariables()

		if assert.Len(t, got, 1) {
			assert.Equal(t, "var-1", got[0].UUID)
			assert.Equal(t, "value-1", got[0].Value)
			assert.Equal(t, "owner-1", got[0].OwnerExternalID)
		}
		assert.Empty(t, v.updateVariables)
	})

	t.Run("region disabled response also clears queued variables", func(t *testing.T) {
		reqURL, err := url.Parse("http://example.com")
		assert.NoError(t, err)

		ipt := defaultInput()
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.Server = "http://example.com"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`kodo.RegionNotFoundOrDisabled`)),
					Header:     make(http.Header),
				}, nil
			}),
		}

		v := &Variable{
			updateVariables: []dialtesting.Variable{
				{UUID: "var-1", Value: "value-1"},
			},
			reqURL: reqURL,
			ipt:    ipt,
		}

		v.updateRemoteVariables()
		assert.Empty(t, v.updateVariables)
	})

	t.Run("non 2xx response still clears queued variables", func(t *testing.T) {
		reqURL, err := url.Parse("http://example.com")
		assert.NoError(t, err)

		ipt := defaultInput()
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.Server = "http://example.com"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader(`bad request`)),
					Header:     make(http.Header),
				}, nil
			}),
		}

		v := &Variable{
			updateVariables: []dialtesting.Variable{
				{UUID: "var-1", Value: "value-1"},
			},
			reqURL: reqURL,
			ipt:    ipt,
		}

		v.updateRemoteVariables()
		assert.Empty(t, v.updateVariables)
	})

	t.Run("request build error still clears queued variables", func(t *testing.T) {
		reqURL, err := url.Parse("http://example.com")
		assert.NoError(t, err)
		reqURL.Host = string([]byte{0x7f})

		v := &Variable{
			updateVariables: []dialtesting.Variable{
				{UUID: "var-1", Value: "value-1"},
			},
			reqURL: reqURL,
			ipt:    defaultInput(),
		}

		v.updateRemoteVariables()
		assert.Empty(t, v.updateVariables)
	})

	t.Run("do error still clears queued variables", func(t *testing.T) {
		reqURL, err := url.Parse("http://example.com")
		assert.NoError(t, err)

		ipt := defaultInput()
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("do failed")
			}),
		}

		v := &Variable{
			updateVariables: []dialtesting.Variable{
				{UUID: "var-1", Value: "value-1"},
			},
			reqURL: reqURL,
			ipt:    ipt,
		}

		v.updateRemoteVariables()
		assert.Empty(t, v.updateVariables)
	})

	t.Run("read body error still clears queued variables", func(t *testing.T) {
		reqURL, err := url.Parse("http://example.com")
		assert.NoError(t, err)

		ipt := defaultInput()
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       &errReadCloser{},
					Header:     make(http.Header),
				}, nil
			}),
		}

		v := &Variable{
			updateVariables: []dialtesting.Variable{
				{UUID: "var-1", Value: "value-1"},
			},
			reqURL: reqURL,
			ipt:    ipt,
		}

		v.updateRemoteVariables()
		assert.Empty(t, v.updateVariables)
	})

	t.Run("empty queued variables returns directly", func(t *testing.T) {
		v := &Variable{
			ipt: defaultInput(),
		}

		assert.NotPanics(t, func() {
			v.updateRemoteVariables()
		})
		assert.Empty(t, v.updateVariables)
	})
}

func TestNewTaskRun(t *testing.T) {
	t.Run("use english region name for english workspace", func(t *testing.T) {
		ipt := defaultInput()
		ipt.RegionID = "region-id"
		ipt.regionName = "region-zh"
		ipt.regionNameEn = "region-en"

		task, err := dialtesting.NewTask("", &dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID:        "task-en",
				Name:              "task-en",
				WorkspaceLanguage: "en",
				Frequency:         "1s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		d, err := ipt.newTaskRun(task)
		assert.NoError(t, err)
		if assert.NotNil(t, d) {
			assert.Equal(t, "region-en", d.regionName)
			assert.NotNil(t, d.done)
		}

		ipt.semStop.Close()
		time.Sleep(20 * time.Millisecond)
	})

	t.Run("headless task returns error", func(t *testing.T) {
		ipt := defaultInput()

		task := &headlessTaskStub{}

		d, err := ipt.newTaskRun(task)
		assert.Nil(t, d)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "headless task deprecated")
	})

	t.Run("fallback to region id when display names are empty", func(t *testing.T) {
		ipt := defaultInput()
		ipt.RegionID = "region-id"

		task, err := dialtesting.NewTask("", &dialtesting.HTTPTask{
			Method: "GET",
			Task: &dialtesting.Task{
				ExternalID: "task-default-region",
				Name:       "task-default-region",
				Frequency:  "1s",
			},
			URL: "http://example.com",
			SuccessWhen: []*dialtesting.HTTPSuccess{
				{
					StatusCode: []*dialtesting.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		assert.NoError(t, err)

		d, err := ipt.newTaskRun(task)
		assert.NoError(t, err)
		if assert.NotNil(t, d) {
			assert.Equal(t, "region-id", d.regionName)
		}

		ipt.semStop.Close()
		time.Sleep(20 * time.Millisecond)
	})

	t.Run("unknown class returns error", func(t *testing.T) {
		ipt := defaultInput()

		task := &genericTaskStub{class: "UNKNOWN"}

		d, err := ipt.newTaskRun(task)
		assert.Nil(t, d)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task type")
	})

	t.Run("supported non-http classes can start", func(t *testing.T) {
		cases := []struct {
			name  string
			child dialtesting.TaskChild
		}{
			{
				name: "tcp",
				child: &dialtesting.TCPTask{
					Task: &dialtesting.Task{
						ExternalID: "task-tcp",
						Name:       "task-tcp",
						Frequency:  "1s",
					},
					Host: "example.com",
				},
			},
			{
				name: "icmp",
				child: &dialtesting.ICMPTask{
					Task: &dialtesting.Task{
						ExternalID: "task-icmp",
						Name:       "task-icmp",
						Frequency:  "1s",
					},
					Host: "example.com",
				},
			},
			{
				name: "websocket",
				child: &dialtesting.WebsocketTask{
					Task: &dialtesting.Task{
						ExternalID: "task-ws",
						Name:       "task-ws",
						Frequency:  "1s",
					},
					URL: "ws://example.com",
				},
			},
			{
				name: "grpc",
				child: &dialtesting.GRPCTask{
					Task: &dialtesting.Task{
						ExternalID: "task-grpc",
						Name:       "task-grpc",
						Frequency:  "1s",
					},
					Server: "127.0.0.1:9529",
				},
			},
			{
				name: "multi",
				child: &dialtesting.MultiTask{
					Task: &dialtesting.Task{
						ExternalID: "task-multi",
						Name:       "task-multi",
						Frequency:  "1s",
					},
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				ipt := defaultInput()
				ipt.RegionID = "region-id"

				task, err := dialtesting.NewTask("", tc.child)
				assert.NoError(t, err)

				d, err := ipt.newTaskRun(task)
				assert.NoError(t, err)
				if assert.NotNil(t, d) {
					assert.Equal(t, task.Class(), d.class)
					assert.Equal(t, "region-id", d.regionName)
					assert.NotNil(t, d.done)
				}

				ipt.semStop.Close()
				time.Sleep(20 * time.Millisecond)
			})
		}
	})
}

type headlessTaskStub struct {
	option map[string]string
}

func (*headlessTaskStub) ID() string                                              { return "headless-task" }
func (*headlessTaskStub) Status() string                                          { return "" }
func (*headlessTaskStub) Run() error                                              { return nil }
func (*headlessTaskStub) Clear()                                                  {}
func (*headlessTaskStub) CheckResult() ([]string, bool)                           { return nil, false }
func (*headlessTaskStub) Class() string                                           { return dialtesting.ClassHeadless }
func (*headlessTaskStub) GetResults() (map[string]string, map[string]interface{}) { return nil, nil }
func (*headlessTaskStub) PostURLStr() string                                      { return "" }
func (*headlessTaskStub) MetricName() string                                      { return "" }
func (*headlessTaskStub) Stop()                                                   {}
func (*headlessTaskStub) RegionName() string                                      { return "" }
func (*headlessTaskStub) AccessKey() string                                       { return "" }
func (*headlessTaskStub) Check() error                                            { return nil }
func (*headlessTaskStub) CheckTask() error                                        { return nil }
func (*headlessTaskStub) UpdateTimeUs() int64                                     { return 0 }
func (*headlessTaskStub) GetFrequency() string                                    { return "" }
func (*headlessTaskStub) GetOwnerExternalID() string                              { return "" }
func (*headlessTaskStub) GetExternalID() string                                   { return "" }
func (*headlessTaskStub) SetOwnerExternalID(string)                               {}
func (*headlessTaskStub) GetLineData() string                                     { return "" }
func (*headlessTaskStub) GetHostName() ([]string, error)                          { return nil, nil }
func (*headlessTaskStub) GetWorkspaceLanguage() string                            { return "zh" }
func (*headlessTaskStub) GetDFLabel() string                                      { return "" }
func (*headlessTaskStub) GetScheduleType() string                                 { return "" }
func (*headlessTaskStub) GetCrontab() string                                      { return "" }
func (h *headlessTaskStub) SetOption(opt map[string]string)                       { h.option = opt }
func (h *headlessTaskStub) GetOption() map[string]string                          { return h.option }
func (*headlessTaskStub) SetRegionID(string)                                      {}
func (*headlessTaskStub) SetAk(string)                                            {}
func (*headlessTaskStub) SetStatus(string)                                        {}
func (*headlessTaskStub) SetUpdateTime(int64)                                     {}
func (*headlessTaskStub) SetChild(dialtesting.TaskChild)                          {}
func (*headlessTaskStub) SetTaskJSONString(string)                                {}
func (*headlessTaskStub) GetTaskJSONString() string                               { return "" }
func (*headlessTaskStub) SetDisabled(uint8)                                       {}
func (*headlessTaskStub) GetVariableValue(dialtesting.Variable) (string, error)   { return "", nil }
func (*headlessTaskStub) GetGlobalVars() []string                                 { return nil }
func (*headlessTaskStub) RenderTemplateAndInit(map[string]dialtesting.Variable) error {
	return nil
}
func (*headlessTaskStub) AddExtractedVar(*dialtesting.ConfigVar)     {}
func (*headlessTaskStub) SetCustomVars([]*dialtesting.ConfigVar)     {}
func (*headlessTaskStub) GetPostScriptVars() dialtesting.Vars        { return nil }
func (*headlessTaskStub) GetIsTemplate() bool                        { return false }
func (*headlessTaskStub) SetIsTemplate(bool)                         {}
func (*headlessTaskStub) SetBeforeRun(func(*dialtesting.Task) error) {}
func (*headlessTaskStub) String() string                             { return "headless-task" }

type genericTaskStub struct {
	headlessTaskStub
	class string
}

func (g *genericTaskStub) ID() string    { return "generic-task" }
func (g *genericTaskStub) Class() string { return g.class }

func TestDoLocalTask(t *testing.T) {
	t.Run("missing file returns directly", func(t *testing.T) {
		ipt := defaultInput()

		assert.NotPanics(t, func() {
			ipt.doLocalTask("/path/not/exist.json")
		})
	})

	t.Run("invalid json file returns directly", func(t *testing.T) {
		ipt := defaultInput()

		f, err := os.CreateTemp(t.TempDir(), "dialtesting-invalid-*.json")
		assert.NoError(t, err)
		_, err = f.WriteString("{invalid-json}")
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		assert.NotPanics(t, func() {
			ipt.doLocalTask(f.Name())
		})
	})

	t.Run("valid file dispatches tasks before exit", func(t *testing.T) {
		oldExit := datakit.Exit
		datakit.Exit = cliutils.NewSem()
		defer func() {
			datakit.Exit.Close()
			datakit.Exit = oldExit
		}()

		ipt := defaultInput()
		ipt.RegionID = "local-region"

		f, err := os.CreateTemp(t.TempDir(), "dialtesting-valid-*.json")
		assert.NoError(t, err)
		_, err = f.WriteString(`{
			"HTTP": [
				{
					"name": "local-http-task",
					"external_id": "local-http-task",
					"method": "GET",
					"url": "http://example.com",
					"post_url": "http://example.com?token=test-token",
					"status": "ok",
					"frequency": "10s",
					"region": "local-region",
					"owner_external_id": "owner",
					"success_when": [
						{
							"status_code": [
								{"is": "200"}
							]
						}
					]
				}
			]
		}`)
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		go func() {
			time.Sleep(20 * time.Millisecond)
			datakit.Exit.Close()
		}()

		assert.NotPanics(t, func() {
			ipt.doLocalTask(f.Name())
		})

		found := false
		ipt.curTasks.Range(func(_, value any) bool {
			d, ok := value.(*dialer)
			if ok && d.task.GetExternalID() == "local-http-task" {
				found = true
				return false
			}
			return true
		})
		assert.True(t, found)
	})
}

func TestVariableRun(t *testing.T) {
	t.Run("nil input exits early", func(t *testing.T) {
		v := &Variable{
			updateVariableCh: make(chan dialtesting.Variable, 1),
		}

		assert.NotPanics(t, func() {
			v.run()
			time.Sleep(20 * time.Millisecond)
		})
		assert.Nil(t, v.reqURL)
	})

	t.Run("invalid server url exits early", func(t *testing.T) {
		v := &Variable{
			updateVariableCh: make(chan dialtesting.Variable, 1),
			ipt: &Input{
				Server: "://invalid-url",
			},
		}

		assert.NotPanics(t, func() {
			v.run()
			time.Sleep(20 * time.Millisecond)
		})
		assert.Nil(t, v.reqURL)
	})
}

func TestDoServerTask(t *testing.T) {
	t.Run("pause mode stops all tasks and clears pos", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Server = "http://example.com"
		ipt.PullInterval = "1m"
		ipt.pause.Store(true)
		ipt.pos = 123

		task := newHTTPDialtestingTask(t, "paused-task")
		d := newDialer(task, ipt)
		ipt.curTasks.Store(task.ID(), d)
		ipt.semStop.Close()

		assert.NotPanics(t, func() {
			ipt.doServerTask()
		})
		assert.Zero(t, ipt.pos)

		found := false
		ipt.curTasks.Range(func(key, value any) bool {
			found = true
			return false
		})
		assert.False(t, found)
	})

	t.Run("pause mode exits on datakit exit", func(t *testing.T) {
		oldExit := datakit.Exit
		datakit.Exit = cliutils.NewSem()
		defer func() {
			datakit.Exit.Close()
			datakit.Exit = oldExit
		}()

		ipt := defaultInput()
		ipt.Server = "http://example.com"
		ipt.PullInterval = "1m"
		ipt.pause.Store(true)
		datakit.Exit.Close()

		assert.NotPanics(t, func() {
			ipt.doServerTask()
		})
	})

	t.Run("invalid pull interval falls back and exits on sem stop", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Server = "http://example.com"
		ipt.PullInterval = "invalid"
		ipt.pause.Store(true)
		ipt.semStop.Close()

		assert.NotPanics(t, func() {
			ipt.doServerTask()
		})
		assert.NotNil(t, ipt.variables.ipt)
	})

	t.Run("non pause branch exits cleanly when pull task fails and sem stop is closed", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Server = "://invalid-url"
		ipt.PullInterval = "1m"
		ipt.semStop.Close()

		assert.NotPanics(t, func() {
			ipt.doServerTask()
		})
	})

	t.Run("too small pull interval falls back and exits on sem stop", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Server = "http://example.com"
		ipt.PullInterval = "1s"
		ipt.pause.Store(true)
		ipt.semStop.Close()

		assert.NotPanics(t, func() {
			ipt.doServerTask()
		})
	})

	t.Run("too large pull interval falls back and exits on sem stop", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Server = "http://example.com"
		ipt.PullInterval = "25h"
		ipt.pause.Store(true)
		ipt.semStop.Close()

		assert.NotPanics(t, func() {
			ipt.doServerTask()
		})
	})

	t.Run("pulls and dispatches one round before sem stop exit", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Server = "http://example.com"
		ipt.RegionID = "test-region"
		ipt.AK = "test-ak"
		ipt.SK = "test-sk"
		ipt.PullInterval = "1m"
		ipt.cli = &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body: io.NopCloser(strings.NewReader(`{
						"content": {
							"region": {
								"name": "region-cn",
								"name_en": "region-en",
								"isp": "telecom"
							}
						}
					}`)),
					Header: make(http.Header),
				}, nil
			}),
		}
		ipt.semStop.Close()

		assert.NotPanics(t, func() {
			ipt.doServerTask()
		})
		assert.Equal(t, "region-cn", ipt.regionName)
		assert.Equal(t, "region-en", ipt.regionNameEn)
		assert.Equal(t, "telecom", ipt.RegionTags["isp"])
	})
}

func TestRun(t *testing.T) {
	t.Run("invalid task exec interval keeps zero and fills default sleep", func(t *testing.T) {
		oldWorker := dialWorker
		oldMaxSendFailCount := MaxSendFailCount
		defer func() {
			dialWorker = oldWorker
			once = sync.Once{}
			MaxSendFailCount = oldMaxSendFailCount
		}()

		dialWorker = nil
		once = sync.Once{}

		ipt := defaultInput()
		ipt.isDebugMode = true
		ipt.Server = "://invalid-url"
		ipt.TaskExecTimeInterval = "invalid"
		ipt.MaxSendFailCount = 7
		ipt.MaxSendFailSleepTime = nil

		assert.NotPanics(t, func() {
			ipt.Run()
		})
		assert.Zero(t, ipt.taskExecTimeInterval)
		if assert.NotNil(t, ipt.MaxSendFailSleepTime) {
			assert.Equal(t, 30*time.Minute, ipt.MaxSendFailSleepTime.Duration)
		}
		assert.Equal(t, 7, MaxSendFailCount)
	})

	t.Run("valid task exec interval is parsed", func(t *testing.T) {
		oldWorker := dialWorker
		defer func() {
			dialWorker = oldWorker
			once = sync.Once{}
		}()

		dialWorker = nil
		once = sync.Once{}

		ipt := defaultInput()
		ipt.isDebugMode = true
		ipt.Server = "ftp://example.com"
		ipt.TaskExecTimeInterval = "2s"

		assert.NotPanics(t, func() {
			ipt.Run()
		})
		assert.Equal(t, 2*time.Second, ipt.taskExecTimeInterval)
		assert.Equal(t, ipt.RegionID, ipt.regionName)
		assert.True(t, ipt.isDebugMode)
	})

	t.Run("empty scheme falls back to local path", func(t *testing.T) {
		oldWorker := dialWorker
		defer func() {
			dialWorker = oldWorker
			once = sync.Once{}
		}()

		dialWorker = nil
		once = sync.Once{}

		ipt := defaultInput()
		ipt.isDebugMode = true
		ipt.Server = "/path/not/exist.json"

		assert.NotPanics(t, func() {
			ipt.Run()
		})
	})

	t.Run("http scheme enters server mode", func(t *testing.T) {
		oldWorker := dialWorker
		defer func() {
			dialWorker = oldWorker
			once = sync.Once{}
		}()

		dialWorker = nil
		once = sync.Once{}

		ipt := defaultInput()
		ipt.isDebugMode = true
		ipt.Server = "http://example.com"
		ipt.PullInterval = "invalid"
		ipt.pause.Store(true)
		ipt.semStop.Close()

		assert.NotPanics(t, func() {
			ipt.Run()
		})
		assert.True(t, ipt.isServerMode)
	})

	t.Run("file scheme uses local mode", func(t *testing.T) {
		oldWorker := dialWorker
		defer func() {
			dialWorker = oldWorker
			once = sync.Once{}
		}()

		dialWorker = nil
		once = sync.Once{}

		ipt := defaultInput()
		ipt.isDebugMode = true
		ipt.Server = "file:///path/not/exist.json"

		assert.NotPanics(t, func() {
			ipt.Run()
		})
		assert.False(t, ipt.isServerMode)
	})

	t.Run("unsupported scheme keeps local flags unchanged", func(t *testing.T) {
		oldWorker := dialWorker
		defer func() {
			dialWorker = oldWorker
			once = sync.Once{}
		}()

		dialWorker = nil
		once = sync.Once{}

		ipt := defaultInput()
		ipt.isDebugMode = true
		ipt.Server = "ftp://example.com"

		assert.NotPanics(t, func() {
			ipt.Run()
		})
		assert.False(t, ipt.isServerMode)
	})
}

func TestDebugRun(t *testing.T) {
	oldExit := datakit.Exit
	datakit.Exit = cliutils.NewSem()
	defer func() {
		datakit.Exit.Close()
		datakit.Exit = oldExit
	}()

	ipt := defaultInput()
	ipt.Server = "://invalid-url"
	datakit.Exit.Close()

	assert.NotPanics(t, func() {
		ipt.DebugRun()
	})
	assert.True(t, ipt.isDebugMode)
}

func TestWorkerRunConsumer(t *testing.T) {
	t.Run("consumer sends queued point", func(t *testing.T) {
		oldExit := datakit.Exit
		datakit.Exit = cliutils.NewSem()
		defer func() {
			datakit.Exit.Close()
			datakit.Exit = oldExit
		}()

		collectPointsCache = nil
		w := &worker{
			sender:               &mockSender{},
			maxJobNumber:         1,
			maxJobChanNumber:     1,
			maxCachePointsNumber: 1,
			flushInterval:        time.Hour,
		}
		w.init()

		pt := point.NewPoint("dialtesting", point.NewKVs(map[string]interface{}{"value": 1}))
		w.jobChans <- &jobData{
			regionName: "test-region",
			class:      "HTTP",
			url:        "http://example.com",
			pt:         pt,
		}

		assert.Eventually(t, func() bool {
			return len(collectPointsCache) == 1
		}, time.Second, 10*time.Millisecond)

		datakit.Exit.Close()
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type errReadCloser struct{}

func (*errReadCloser) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func (*errReadCloser) Close() error { return nil }

func TestStopAlltask(t *testing.T) {
	ipt := defaultInput()

	task1, err := dialtesting.NewTask("", &dialtesting.HTTPTask{
		Task: &dialtesting.Task{
			ExternalID: "task-1",
			Name:       "task-1",
			Frequency:  "1s",
		},
		URL: "http://example.com",
		SuccessWhen: []*dialtesting.HTTPSuccess{
			{
				StatusCode: []*dialtesting.SuccessOption{
					{Is: "200"},
				},
			},
		},
	})
	assert.NoError(t, err)

	task2, err := dialtesting.NewTask("", &dialtesting.HTTPTask{
		Task: &dialtesting.Task{
			ExternalID: "task-2",
			Name:       "task-2",
			Frequency:  "1s",
		},
		URL: "http://example.com",
		SuccessWhen: []*dialtesting.HTTPSuccess{
			{
				StatusCode: []*dialtesting.SuccessOption{
					{Is: "200"},
				},
			},
		},
	})
	assert.NoError(t, err)

	d1 := newDialer(task1, ipt)
	d2 := newDialer(task2, ipt)

	ipt.curTasks.Store(task1.ID(), d1)
	ipt.curTasks.Store(task2.ID(), d2)

	assert.NotPanics(t, func() {
		ipt.stopAlltask()
	})

	found := false
	ipt.curTasks.Range(func(key, value any) bool {
		found = true
		return false
	})
	assert.False(t, found)
}

func TestStopAlltaskEmpty(t *testing.T) {
	ipt := defaultInput()

	assert.NotPanics(t, func() {
		ipt.stopAlltask()
	})
}

func TestSetupCli(t *testing.T) {
	oldProxy := config.Cfg.Dataway.HTTPProxy
	defer func() {
		config.Cfg.Dataway.HTTPProxy = oldProxy
	}()

	t.Run("default timeout", func(t *testing.T) {
		ipt := defaultInput()
		ipt.TimeOut = nil
		config.Cfg.Dataway.HTTPProxy = ""

		assert.NotPanics(t, func() {
			ipt.setupCli()
		})
		assert.NotNil(t, ipt.cli)
	})

	t.Run("custom timeout", func(t *testing.T) {
		ipt := defaultInput()
		ipt.TimeOut = &datakit.Duration{Duration: 5 * time.Second}
		config.Cfg.Dataway.HTTPProxy = ""

		assert.NotPanics(t, func() {
			ipt.setupCli()
		})
		assert.NotNil(t, ipt.cli)
	})

	t.Run("invalid proxy is ignored", func(t *testing.T) {
		ipt := defaultInput()
		config.Cfg.Dataway.HTTPProxy = "://invalid-proxy"

		assert.NotPanics(t, func() {
			ipt.setupCli()
		})
		assert.NotNil(t, ipt.cli)
	})

	t.Run("valid proxy is accepted", func(t *testing.T) {
		ipt := defaultInput()
		config.Cfg.Dataway.HTTPProxy = "http://127.0.0.1:8080"

		assert.NotPanics(t, func() {
			ipt.setupCli()
		})
		assert.NotNil(t, ipt.cli)
	})
}
