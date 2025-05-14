// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package winevent collect Windows event metrics
//
//nolint:lll
package winevent

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	sample = `
[[inputs.windows_event]]
  xpath_query = '''
  <QueryList>
    <Query Id="0" Path="Security">
      <Select Path="Security">*</Select>
      <Suppress Path="Security">*[System[( (EventID &gt;= 5152 and EventID &lt;= 5158) or EventID=5379 or EventID=4672)]]</Suppress>
    </Query>
    <Query Id="1" Path="Application">
      <Select Path="Application">*[System[(Level &lt; 4)]]</Select>
    </Query>
    <Query Id="2" Path="Windows PowerShell">
      <Select Path="Windows PowerShell">*[System[(Level &lt; 4)]]</Select>
    </Query>
    <Query Id="3" Path="System">
      <Select Path="System">*</Select>
    </Query>
    <Query Id="4" Path="Setup">
      <Select Path="Setup">*</Select>
    </Query>
  </QueryList>
  '''

  # event_fetch_size is the number of events to fetch per query.
  event_fetch_size = 5

  [inputs.windows_event.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...`

	inputName = "windows_event"
)

//nolint:unused
type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Cat:  point.Logging,
		Fields: map[string]interface{}{
			"event_id":        &inputs.FieldInfo{Desc: "Event ID", DataType: inputs.String},
			"event_record_id": &inputs.FieldInfo{Desc: "Event record ID", DataType: inputs.String},
			"status":          &inputs.FieldInfo{Desc: "Log level", DataType: inputs.String},
			"event_source":    &inputs.FieldInfo{Desc: "Windows event source", DataType: inputs.String},
			"version":         &inputs.FieldInfo{Desc: "Version", DataType: inputs.String},
			"task":            &inputs.FieldInfo{Desc: "Task category", DataType: inputs.String},
			"keyword":         &inputs.FieldInfo{Desc: "Keyword", DataType: inputs.String},
			"process_id":      &inputs.FieldInfo{Desc: "Process ID", DataType: inputs.Int},
			"channel":         &inputs.FieldInfo{Desc: "Channel", DataType: inputs.String},
			"computer":        &inputs.FieldInfo{Desc: "Computer", DataType: inputs.String},
			"message":         &inputs.FieldInfo{Desc: "Event content", DataType: inputs.String},
			"level":           &inputs.FieldInfo{Desc: "Level", DataType: inputs.String},
			"total_message":   &inputs.FieldInfo{Desc: "Full text of the event", DataType: inputs.String},
		},
	}
}
