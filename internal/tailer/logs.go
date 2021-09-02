package tailer

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	// pipeline关键字段
	pipelineTimeField = "time"

	// ES value can be at most 32766 bytes long
	maxFieldsLength = 32766

	// 不使用高频IO
	disableHighFreqIODdata = false
)

type Logs struct {
	text   string
	fields map[string]interface{}
	ts     time.Time
	pt     *io.Point
	err    error
}

func NewLogs(text string) *Logs {
	return &Logs{text: text, fields: make(map[string]interface{})}
}

func (x *Logs) Pipeline(p *pipeline.Pipeline) *Logs {
	if x.err != nil || p == nil {
		x.fields["message"] = x.text
		return x
	}

	x.fields, x.err = p.Run(x.text).Result()

	if len(x.fields) == 0 {
		x.err = fmt.Errorf("fields is empty, maybe the use of delete_origin_data() of pipeline")
	}
	return x
}

// checkFieldsLength 检查数据是否过长
// 只有在碰到非 message 字段，且长度超过最大限制时才会返回 error
// 防止通过 pipeline 添加巨长字段的恶意行为
func (x *Logs) CheckFieldsLength() *Logs {
	if x.err != nil {
		return x
	}

	func() {
		for k, v := range x.fields {
			switch vv := v.(type) {
			case string:
				if len(vv) <= maxFieldsLength {
					continue
				}
				if k == "message" {
					x.fields[k] = vv[:maxFieldsLength]
				} else {
					x.err = fmt.Errorf("fields[%s], length=%d, out of maximum length", k, len(vv))
					return
				}
			default:
				// nil
			}
		}
	}()

	return x
}

const (
	DEFAULT_INFO = "info"
)

var statusMap = map[string]string{
	"f":        "emerg",
	"emerg":    "emerg",
	"a":        "alert",
	"alert":    "alert",
	"c":        "critical",
	"critical": "critical",
	"e":        "error",
	"error":    "error",
	"w":        "warning",
	"warning":  "warning",
	"i":        "info",
	"info":     "info",
	"d":        "debug",
	"trace":    "debug",
	"verbose":  "debug",
	"debug":    "debug",
	"o":        "OK",
	"s":        "OK",
	"ok":       "OK",
}

// addStatus 添加默认status和status映射
func (x *Logs) AddStatus(disable bool) *Logs {
	if x.err != nil || disable {
		return x
	}

	// 不包含 status 字段
	statusField, ok := x.fields["status"]
	if !ok {
		x.fields["status"] = DEFAULT_INFO
		return x
	}

	// status 类型必须是 string
	statusStr, ok := statusField.(string)
	if !ok {
		x.fields["status"] = DEFAULT_INFO
		return x
	}

	// 查询 statusMap 枚举表并替换
	if v, ok := statusMap[strings.ToLower(statusStr)]; !ok {
		x.fields["status"] = DEFAULT_INFO
	} else {
		x.fields["status"] = v
	}
	return x
}

// ignoreStatus 过滤指定status
func (x *Logs) IgnoreStatus(ignoreStatus []string) *Logs {
	if x.err != nil || len(ignoreStatus) == 0 {
		return x
	}

	if status, ok := x.fields["status"].(string); ok {
		for _, ignore := range ignoreStatus {
			if ignore == status {
				x.err = fmt.Errorf("this fields has been ignored, status:%s", status)
				break
			}
		}
	}

	return x
}

func (x *Logs) TakeTime() *Logs {
	if x.err != nil {
		return x
	}

	// time should be nano-second
	if v, ok := x.fields[pipelineTimeField]; ok {
		nanots, ok := v.(int64)
		if !ok {
			x.err = fmt.Errorf("invalid filed `%s: %v', should be nano-second, but got `%s'",
				pipelineTimeField, v, reflect.TypeOf(v).String())
			return x
		}

		x.ts = time.Unix(nanots/int64(time.Second), nanots%int64(time.Second))
		delete(x.fields, pipelineTimeField)
	} else {
		x.ts = time.Now()
	}

	return x
}

func (x *Logs) Point(measurement string, tags map[string]string) *Logs {
	if x.err != nil {
		return x
	}
	x.pt, x.err = io.MakePoint(measurement, tags, x.fields, x.ts)
	return x
}

func (x *Logs) Feed(inputName string) *Logs {
	if x.err != nil {
		return x
	}
	x.err = io.Feed(inputName,
		datakit.Logging,
		[]*io.Point{x.pt},
		&io.Option{HighFreq: disableHighFreqIODdata},
	)
	return x
}

func (x *Logs) Output() string {
	if x.pt == nil {
		return ""
	}
	return x.pt.String()
}

func (x *Logs) Error() error {
	return x.err
}

func feed(inputName, measurement string, tags map[string]string, message string) error {
	pt, err := io.MakePoint(measurement,
		tags,
		map[string]interface{}{"message": message},
		time.Now())

	if err != nil {
		return err
	}

	err = io.Feed(inputName,
		datakit.Logging,
		[]*io.Point{pt},
		&io.Option{HighFreq: disableHighFreqIODdata},
	)

	return err
}
