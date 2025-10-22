// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/encoding"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"
)

// config 日志采集器的配置结构体.
type config struct {
	// 网络套接字配置
	sockets []string
	// 忽略的文件模式，支持通配符
	ignorePatterns []string
	// 数据源名称，用于标识日志来源
	source string
	// 服务名称，默认为 source 的值
	service string
	// Pipeline 脚本路径，用于日志处理
	pipeline string
	// 存储索引名称
	storageIndex string
	// 字符编码，支持 utf-8、utf-16le、utf-16be、gbk、gb18030 等
	characterEncoding string

	// 最大同时打开的文件数量，-1 表示无限制
	maxOpenFiles int
	// 忽略不活跃文件的时间阈值
	ignoreDeadLog time.Duration
	// 是否从文件开头开始读取（可能导致重复数据）
	fromBeginning bool
	// 文件大小阈值，小于此值的文件将从开头读取
	fileSizeThreshold int64
	// 是否移除 ANSI 转义码
	removeAnsiEscapeCodes bool
	// 是否启用多行日志处理
	enableMultiline bool
	// 多行日志匹配模式，用于识别日志行开始
	// 例如：^\d{4}-\d{2}-\d{2} 匹配 YYYY-MM-DD 格式
	multilinePatterns []string
	// 多行日志最大长度限制
	maxMultilineLength int64

	// 自定义日志转发函数（与 Feed 冲突）
	forwardFunc ForwardFunc
	// 内部文件路径处理函数
	insideFilepathFunc func(string) string

	// 是否启用调试字段
	enableDebugFields bool
	// 额外的标签信息
	extraTags map[string]string
	// 字段白名单，只有在此列表中的字段才会被保留
	fieldWhitelist []string

	// 日志解析模式
	mode Mode
	// 数据输出器
	feeder dkio.Feeder

	// 已废弃：忽略的状态列表
	ignoredStatuses []string
	// 已废弃：是否禁用默认状态字段
	disableStatusField bool
}

// checkConfig 验证配置的有效性.
func checkConfig(cfg *config) error {
	// 验证字符编码
	if cfg.characterEncoding != "" {
		if _, err := encoding.NewDecoder(cfg.characterEncoding); err != nil {
			return fmt.Errorf("invalid character encoding '%s': %w", cfg.characterEncoding, err)
		}
	}

	// 验证多行模式
	if _, err := multiline.New(cfg.multilinePatterns); err != nil {
		return fmt.Errorf("invalid multiline patterns: %w", err)
	}

	// 验证数据源名称
	if cfg.source == "" {
		return fmt.Errorf("source cannot be empty")
	}

	// 验证最大打开文件数
	if cfg.maxOpenFiles < -1 {
		return fmt.Errorf("maxOpenFiles must be >= -1, got %d", cfg.maxOpenFiles)
	}

	// 验证文件大小阈值
	if cfg.fileSizeThreshold < 0 {
		return fmt.Errorf("fileSizeThreshold must be >= 0, got %d", cfg.fileSizeThreshold)
	}

	// 验证多行最大长度
	if cfg.maxMultilineLength < 0 {
		return fmt.Errorf("maxMultilineLength must be >= 0, got %d", cfg.maxMultilineLength)
	}

	return nil
}

// Option 配置选项函数类型.
type Option func(*config)

func WithIgnoredStatuses(arr []string) Option {
	return func(cfg *config) { cfg.ignoredStatuses = arr }
}

func WithDisableStatusField(b bool) Option {
	return func(cfg *config) { cfg.disableStatusField = b }
}

func WithSockets(arr []string) Option {
	return func(cfg *config) { cfg.sockets = arr }
}

func WithIgnorePatterns(arr []string) Option {
	return func(cfg *config) { cfg.ignorePatterns = arr }
}

func WithPipeline(s string) Option {
	return func(cfg *config) { cfg.pipeline = s }
}

func WithStorageIndex(name string) Option {
	return func(cfg *config) { cfg.storageIndex = name }
}

func WithCharacterEncoding(s string) Option {
	return func(cfg *config) { cfg.characterEncoding = s }
}

func WithFromBeginning(b bool) Option {
	return func(cfg *config) { cfg.fromBeginning = b }
}

func WithTextParserMode(mode Mode) Option {
	return func(cfg *config) { cfg.mode = mode }
}

func EnableDebugFields(b bool) Option {
	return func(cfg *config) { cfg.enableDebugFields = b }
}

func WithFieldWhitelist(list []string) Option {
	return func(cfg *config) { cfg.fieldWhitelist = list }
}

func WithSource(s string) Option {
	return func(cfg *config) {
		if s == "" {
			return
		}
		cfg.source = s
		if cfg.service == "" {
			WithService(s)(cfg)
		}
	}
}

func WithService(s string) Option {
	return func(cfg *config) {
		if s == "" {
			s = cfg.source
		}
		cfg.service = s
		if cfg.extraTags == nil {
			cfg.extraTags = make(map[string]string)
		}
		cfg.extraTags["service"] = cfg.service
	}
}

func EnableMultiline(b bool) Option {
	return func(cfg *config) { cfg.enableMultiline = b }
}

func WithMultilinePatterns(arr []string) Option {
	return func(cfg *config) { cfg.multilinePatterns = arr }
}

func WithMaxMultilineLength(n int64) Option {
	return func(cfg *config) { cfg.maxMultilineLength = n }
}

func WithRemoveAnsiEscapeCodes(b bool) Option {
	return func(cfg *config) { cfg.removeAnsiEscapeCodes = b }
}

func WithMaxOpenFiles(n int) Option {
	return func(cfg *config) {
		if n > 0 || n == -1 {
			cfg.maxOpenFiles = n
		}
	}
}

func WithIgnoreDeadLog(dur time.Duration) Option {
	return func(cfg *config) {
		if dur > 0 {
			cfg.ignoreDeadLog = dur
		}
	}
}

func WithFileSizeThreshold(n int64) Option {
	return func(cfg *config) {
		if n > 0 {
			cfg.fileSizeThreshold = n
		}
	}
}

func WithExtraTags(m map[string]string) Option {
	return func(cfg *config) {
		// 保留 service 标签，删除其他标签
		service := cfg.extraTags["service"]
		cfg.extraTags = map[string]string{"service": service}
		for k, v := range m {
			cfg.extraTags[k] = v
		}
	}
}

func AddTag(key, value string) Option {
	return func(cfg *config) {
		if cfg.extraTags == nil {
			cfg.extraTags = make(map[string]string)
		}
		cfg.extraTags[key] = value
	}
}

func WithInsideFilepathFunc(fn func(path string) string) Option {
	return func(cfg *config) {
		cfg.insideFilepathFunc = fn
	}
}

func WithForwardFunc(fn ForwardFunc) Option {
	return func(cfg *config) { cfg.forwardFunc = fn }
}

func WithFeeder(feeder dkio.Feeder) Option {
	return func(cfg *config) { cfg.feeder = feeder }
}

func defaultConfig() *config {
	return &config{
		source:            "default",
		extraTags:         map[string]string{"service": "default"},
		fileSizeThreshold: 1000 * 1000 * 20, // 20 MB
		maxOpenFiles:      defaultMaxOpenFiles,
		feeder:            dkio.DefaultFeeder(),
	}
}

func buildConfig(opts []Option) *config {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
	return cfg
}

func mergeOptions(oldOpts, newOpts []Option) []Option {
	cfg := buildConfig(oldOpts)
	for _, opt := range newOpts {
		opt(cfg)
	}
	return []Option{
		WithSockets(cfg.sockets),
		WithIgnorePatterns(cfg.ignorePatterns),
		WithSource(cfg.source),
		WithService(cfg.service),
		WithPipeline(cfg.pipeline),
		WithStorageIndex(cfg.storageIndex),
		WithCharacterEncoding(cfg.characterEncoding),

		EnableDebugFields(cfg.enableDebugFields),
		EnableMultiline(cfg.enableMultiline),
		WithMultilinePatterns(cfg.multilinePatterns),
		WithMaxMultilineLength(cfg.maxMultilineLength),

		WithMaxOpenFiles(cfg.maxOpenFiles),
		WithIgnoreDeadLog(cfg.ignoreDeadLog),
		WithFileSizeThreshold(cfg.fileSizeThreshold),
		WithFromBeginning(cfg.fromBeginning),
		WithRemoveAnsiEscapeCodes(cfg.removeAnsiEscapeCodes),

		WithExtraTags(cfg.extraTags),
		WithFieldWhitelist(cfg.fieldWhitelist),

		WithForwardFunc(cfg.forwardFunc),
		WithInsideFilepathFunc(cfg.insideFilepathFunc),

		WithTextParserMode(cfg.mode),
		WithFeeder(cfg.feeder),
	}
}

type ForwardFunc func(filename, text string, fields map[string]interface{}) error

type Mode uint8

const (
	// FileMode 普通文件模式.
	FileMode Mode = iota + 1
	// DockerJSONLogMode Docker JSON 日志模式.
	DockerJSONLogMode
	// CriLogdMode CRI 日志模式.
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
