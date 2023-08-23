// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package profile defines profile related functions
package profile

import (
	"strings"
)

type Lang string

func (l Lang) String() string {
	return string(l)
}

const (
	Java    Lang = "Java"
	Python  Lang = "Python"
	Golang  Lang = "Golang"
	Ruby    Lang = "Ruby"
	NodeJS  Lang = "NodeJS"
	PHP     Lang = "PHP"
	DotNet  Lang = "DotNet"
	UnKnown Lang = "Unknown Language"
)

var langMaps = map[string]Lang{
	"java":    Java,
	"jvm":     Java,
	"python":  Python,
	"ruby":    Ruby,
	"node.js": NodeJS,
	"nodejs":  NodeJS,
	"php":     PHP,
	"c#":      DotNet,
	"csharp":  DotNet,
	"golang":  Golang,
	"node":    NodeJS,
	"go":      Golang,
}

func ResolveLanguage(runtimes []string) Lang {
	for _, r := range runtimes {
		r = strings.ToLower(r)
		for name, lang := range langMaps {
			if strings.Contains(r, name) {
				return lang
			}
		}
	}
	return UnKnown
}

type EventType string

type FlameGraph struct {
	Unit               string   `json:"unit"`                // 该火焰图上Value的单位 eg: ns, count...
	AvailableDimension []string `json:"available_dimension"` // 支持的展示维度， eg: method、method + line、thread....
	Dimension          string   `json:"dimension"`           // 当前展示维度
	RootFrame          IFrame   `json:"root_frame"`          // 火焰图根级图块
}

type IFrame interface {
	IsLeaf() bool
}

// Frame 代表火焰图的某个小块.
type Frame struct {
	Value    int64    `json:"value"`     // 该图块的值
	Method   string   `json:"method"`    // 方法名
	Line     int      `json:"line"`      // 代码行号
	SubFrame []IFrame `json:"sub_frame"` // 下级火焰图块, 直接下级图块的Value总和等于父级Value
}

func (f *Frame) IsLeaf() bool {
	return len(f.SubFrame) == 0
}

type GoFrame struct {
	*Frame
	SourceFile string `json:"source_file"` // 代码源文件名
}

type JavaFrame struct {
	*Frame
	Thread   string `json:"thread"`   // 所属线程
	Modifier string `json:"modifier"` // 方法修饰符：public, private, static ....
	Library  string `json:"library"`  // 所属库
	Package  string `json:"package"`  // 包名
	Class    string `json:"class"`    // 类名
}

type PythonFrame struct {
	*Frame
}

//nolint:stylecheck
type Profile struct {
	ProfileId   string                `json:"profile_id"`   // 使用google uuid作为profile唯一id
	Endpoint    string                `json:"endpoint"`     // from where eg: /profiling/v1/input
	AgentVer    string                `json:"agent_ver"`    // 客户端采用的agent库版本
	Service     string                `json:"service"`      // service name
	Env         string                `json:"env"`          // 运行环境
	Version     string                `json:"version"`      // 客户端代码设置的 version
	Start       int64                 `json:"start"`        // 开始采集的时间戳, Unit：ns
	End         int64                 `json:"end"`          // 采集结束的时间戳, Unit: ns
	Duration    int64                 `json:"duration"`     // 采集持续的时间, Unit: ns
	Runtime     string                `json:"runtime"`      // 语言运行时环境，eg: jvm, cpython, go, ...
	Language    Lang                  `json:"language"`     // 客户端语言, eg: Java, Golang, Python, ...
	LanguageVer string                `json:"language_ver"` // 客户端语言版本
	Format      string                `json:"format"`       // profile文件格式, jfr，pprof...
	Host        string                `json:"host"`
	Pid         int                   `json:"pid"`          // 进程号
	RuntimeId   string                `json:"runtime_id"`   // 运行id
	RuntimeOs   string                `json:"runtime_os"`   // os
	RuntimeArch string                `json:"runtime_arch"` // architecture
	Tags        []string              `json:"tags"`         // form field "tags[]"
	OssPath     []string              `json:"-"`
	Metrics     map[string]string     `json:"-"`
	Samples     map[EventType]*Sample `json:"-"`
}

type SampleValue struct {
	Type  string  `json:"type"`  // 值的类别， cpu, wall, inuse_space ...
	Value []int64 `json:"value"` // 具体数值
	Unit  string  `json:"unit"`  // 值的单位：seconds, nanoseconds, bytes ...
}

type TraceFunc struct {
	Name string `json:"name"` // 函数名
	File string `json:"file"` // 所在代码源文件名
	Line int    `json:"line"` // 所在源文件行号
}

//nolint:stylecheck
type SpanProfileMap struct {
	SpanId    string
	ProfileId string
	SpanStart int64
	SpanEnd   int64
	RuntimeId string
}

// Sample 代表profile中的某个样本详情.
//
//nolint:stylecheck
type Sample struct {
	Values        []SampleValue     `json:"values"`         // 采样到的一些数值，比如占用cpu时间，内存使用大小
	SpanId        string            `json:"span_id"`        // 关联的trace span id
	RootSpanId    string            `json:"root_span_id"`   // 关联的trace root span id
	TraceEndpoint string            `json:"trace_endpoint"` // 关联的span resource
	Labels        map[string]string `json:"labels"`         // 其他自定义Label
	StackTrace    []TraceFunc       `json:"stack_trace"`    // 调用堆栈
}
