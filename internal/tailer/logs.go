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

	ignoreLogs bool
	errs       []error
}

func NewLogs(text string) *Logs {
	return &Logs{text: text, fields: make(map[string]interface{})}
}

func (x *Logs) Pipeline(p *pipeline.Pipeline) *Logs {
	if x.IsIgnored() {
		return x
	}

	if p == nil {
		x.fields["message"] = x.text
		return x
	}

	fields, err := p.Run(x.text).Result()
	if err != nil {
		x.AddErr(err)
	}

	for k, v := range fields {
		x.fields[k] = v
	}

	if len(x.fields) == 0 {
		x.AddErr(fmt.Errorf("fields is empty, maybe the use of delete_origin_data() of pipeline"))
	}

	return x
}

// checkFieldsLength 检查数据是否过长
// 只有在碰到非 message 字段，且长度超过最大限制时才会返回 error
// 防止通过 pipeline 添加巨长字段的恶意行为
func (x *Logs) CheckFieldsLength() *Logs {
	if x.IsIgnored() {
		return x
	}
	for k, v := range x.fields {
		vv, ok := v.(string)
		if !ok {
			continue
		}
		if len(vv) <= maxFieldsLength {
			continue
		}

		if k == "message" {
			x.fields[k] = vv[:maxFieldsLength]
		} else {
			delete(x.fields, k)
			x.AddErr(fmt.Errorf("discard fields[%s], length=%d, out of maximum length", k, len(vv)))
		}
	}

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
	if x.IsIgnored() || disable {
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
	if x.IsIgnored() || len(ignoreStatus) == 0 {
		return x
	}

	status, ok := x.fields["status"].(string)
	if !ok {
		return x
	}
	for _, ignore := range ignoreStatus {
		if ignore == status {
			x.ignoreLogs = true
			x.AddErr(fmt.Errorf("this fields has been ignored, status:%s", status))
			return x
		}
	}

	return x
}

func (x *Logs) TakeTime() *Logs {
	if x.IsIgnored() {
		return x
	}

	// time should be nano-second
	if v, ok := x.fields[pipelineTimeField]; ok {
		nanots, ok := v.(int64)
		if !ok {
			x.ts = time.Now()
			x.AddErr(fmt.Errorf("invalid filed `%s: %v', should be nano-second, but got `%s'",
				pipelineTimeField, v, reflect.TypeOf(v).String()))
		}

		x.ts = time.Unix(nanots/int64(time.Second), nanots%int64(time.Second))
		delete(x.fields, pipelineTimeField)
	} else {
		x.ts = time.Now()
	}

	return x
}

func (x *Logs) Point(measurement string, tags map[string]string) *Logs {
	if x.IsIgnored() {
		return x
	}
	pt, err := io.MakePoint(measurement, tags, x.fields, x.ts)
	if err != nil {
		x.AddErr(err)
	}
	x.pt = pt
	return x
}

func (x *Logs) Feed(inputName string) *Logs {
	if x.IsIgnored() {
		return x
	}
	if x.pt == nil {
		return x
	}

	err := io.Feed(inputName,
		datakit.Logging,
		[]*io.Point{x.pt},
		&io.Option{HighFreq: disableHighFreqIODdata},
	)
	if err != nil {
		x.AddErr(err)
	}
	return x
}

func (x *Logs) Output() string {
	if x.pt == nil {
		return ""
	}
	return x.pt.String()
}

func (x *Logs) MergeErrs() error {
	if len(x.errs) == 0 {
		return nil
	}
	return fmt.Errorf("%v", x.errs)
}

func (x *Logs) AddErr(err error) {
	x.errs = append(x.errs, err)
}

func (x *Logs) IsIgnored() bool {
	return x.ignoreLogs
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
