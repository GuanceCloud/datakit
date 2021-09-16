package win_event

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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
  [inputs.windows_event.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ... `

	inputName = "windows_event"
)

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *Measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Fields: map[string]interface{}{},
		Tags: map[string]interface{}{
			"event_id":        inputs.NewTagInfo("事件 ID"),
			"event_record_id": inputs.NewTagInfo("事件记录 ID"),
			"source":          inputs.NewTagInfo("日志来源"),
			"event_source":    inputs.NewTagInfo("Windows 事件来源"),
			"version":         inputs.NewTagInfo("版本"),
			"task":            inputs.NewTagInfo("任务类别"),
			"keyword":         inputs.NewTagInfo("关键字"),
			"process_id":      inputs.NewTagInfo("进程 ID"),
			"channel":         inputs.NewTagInfo("Channel"),
			"computer":        inputs.NewTagInfo("计算机"),
			"message":         inputs.NewTagInfo("事件内容"),
			"level":           inputs.NewTagInfo("级别"),
			"total_message":   inputs.NewTagInfo("事件全文"),
		},
	}
}
