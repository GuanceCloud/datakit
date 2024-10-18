// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"time"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

type option struct {
	// sockets
	sockets []string
	// 忽略这些文件
	ignorePatterns []string
	// 忽略这些status，如果数据的status在此列表中，数据将不再上传
	// e.g. "info","debug"
	ignoreStatus []string
	// 数据来源，默认值为'default'
	source string
	// service，默认值为 $Source
	service string
	// pipeline脚本路径，如果为空则不使用pipeline
	pipeline string
	// 解释文件内容时所使用的的字符编码，如果设置为空，将不进行转码处理
	// e.g. "utf-8","utf-16le","utf-16be","gbk","gb18030"
	characterEncoding string
	// 添加 debug 字段
	enableDebugFields bool

	// 匹配正则表达式
	// 符合此正则匹配的数据，将被认定为有效数据。否则会累积追加到上一条有效数据的末尾
	// 例如 ^\d{4}-\d{2}-\d{2} 行首匹配 YYYY-MM-DD 时间格式
	// 如果为空，则默认使用 ^\S 即匹配每行开始处非空白字符
	multilinePatterns []string
	// 最大多行存在时间，避免堆积过久
	maxMultilineLifeDuration time.Duration

	// 是否从文件起始处开始读取，如果打开此项，可能会导致大量数据重复
	fromBeginning bool
	// 是否删除文本中的ansi转义码，默认为false，即不删除
	removeAnsiEscapeCodes bool
	// 是否关闭添加默认status字段列，包括status字段的固定转换行为，例如'd'->'debug'
	disableAddStatusField bool
	// 日志文本的另一种发送方式（和Feed冲突）
	forwardFunc ForwardFunc
	// 判定不活跃文件
	ignoreDeadLog time.Duration
	// 添加额外tag
	extraTags map[string]string

	// 连续 N 次采集为空，就强制 flush 已有数据
	maxForceFlushLimit int
	// 如果要采集的文件 size 小于此值，将使用 from_bgeinning，单位字节
	fileFromBeginningThresholdSize int64

	setDone bool
	done    <-chan interface{}
	mode    Mode

	feeder dkio.Feeder
}

type Option func(*option)

func WithSockets(arr []string) Option        { return func(opt *option) { opt.sockets = arr } }
func WithIgnorePatterns(arr []string) Option { return func(opt *option) { opt.ignorePatterns = arr } }
func WithIgnoreStatus(arr []string) Option   { return func(opt *option) { opt.ignoreStatus = arr } }
func WithPipeline(s string) Option           { return func(opt *option) { opt.pipeline = s } }
func WithCharacterEncoding(s string) Option  { return func(opt *option) { opt.characterEncoding = s } }
func WithFromBeginning(b bool) Option        { return func(opt *option) { opt.fromBeginning = b } }
func WithEnableDebugFields(b bool) Option    { return func(opt *option) { opt.enableDebugFields = b } }
func WithTextParserMode(mode Mode) Option    { return func(opt *option) { opt.mode = mode } }

func WithSource(s string) Option {
	return func(opt *option) {
		if s == "" {
			return
		}
		opt.source = s
		if opt.service == "" {
			WithService(s)
		}
	}
}

func WithService(s string) Option {
	return func(opt *option) {
		if s == "" {
			s = opt.source
		}
		opt.service = s
		if opt.extraTags == nil {
			opt.extraTags = make(map[string]string)
		}
		opt.extraTags["service"] = opt.service
	}
}

func WithMultilinePatterns(arr []string) Option {
	return func(opt *option) { opt.multilinePatterns = arr }
}

func WithMaxMultilineLifeDuration(dur time.Duration) Option {
	return func(opt *option) {
		if dur > 0 {
			opt.maxMultilineLifeDuration = dur
		}
	}
}

func WithRemoveAnsiEscapeCodes(b bool) Option {
	return func(opt *option) { opt.removeAnsiEscapeCodes = b }
}

func WithDisableAddStatusField(b bool) Option {
	return func(opt *option) { opt.disableAddStatusField = b }
}

func WithIgnoreDeadLog(dur time.Duration) Option {
	return func(opt *option) {
		if dur > 0 {
			opt.ignoreDeadLog = dur
		}
	}
}

func WithFileFromBeginningThresholdSize(n int64) Option {
	return func(opt *option) {
		if n > 0 {
			opt.fileFromBeginningThresholdSize = n
		}
	}
}

func WithMaxForceFlushLimit(n int) Option {
	return func(opt *option) {
		if n > 0 {
			opt.maxForceFlushLimit = n
		}
	}
}

func WithGlobalTags(m map[string]string) Option {
	return func(opt *option) {
		for k, v := range m {
			opt.extraTags[k] = v
		}
	}
}

func WithTag(key, value string) Option {
	return func(opt *option) {
		if opt.extraTags == nil {
			opt.extraTags = make(map[string]string)
		}
		opt.extraTags[key] = value
	}
}

func WithDone(ch <-chan interface{}) Option {
	return func(opt *option) { opt.setDone = true; opt.done = ch }
}
func WithForwardFunc(fn ForwardFunc) Option { return func(opt *option) { opt.forwardFunc = fn } }
func WithFeeder(feeder dkio.Feeder) Option  { return func(opt *option) { opt.feeder = feeder } }

func defaultOption() *option {
	return &option{
		source:                         "default",
		extraTags:                      map[string]string{"service": "default"},
		maxForceFlushLimit:             10,
		fileFromBeginningThresholdSize: 1000 * 1000 * 1, // 1 MB
		done:                           make(<-chan interface{}),
		feeder:                         dkio.DefaultFeeder(),
	}
}

type ForwardFunc func(filename, text string, fields map[string]interface{}) error

type Mode uint8

const (
	FileMode Mode = iota + 1
	DockerJSONLogMode
	CriLogdMode
)

func (mode Mode) String() string {
	switch mode {
	case FileMode:
		return "file"
	case DockerJSONLogMode:
		return "docker-json"
	case CriLogdMode:
		return "cri-log"
	}
	return "unknown"
}
