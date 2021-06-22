package container

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

type LogFilters []*LogFilter

type LogFilter struct {
	FilterMessage   []string `toml:"filter_message"`
	FilterMultiline string   `toml:"-"`
	Source          string   `toml:"source"`
	Service         string   `toml:"service"`
	Pipeline        string   `toml:"pipeline"`

	pipelinePool sync.Pool

	multilinePattern *regexp.Regexp
	messagePattern   []*regexp.Regexp
}

func (lf *LogFilter) Init() error {
	if lf.Service == "" {
		lf.Service = lf.Source
	}

	if lf.FilterMultiline != "" {
		pattern, err := regexp.Compile(lf.FilterMultiline)
		if err != nil {
			return fmt.Errorf("config FilterMultiline, error: %s", err)
		}
		lf.multilinePattern = pattern
	}

	// regexp
	for idx, m := range lf.FilterMessage {
		pattern, err := regexp.Compile(m)
		if err != nil {
			return fmt.Errorf("config FilterMessage index[%d], error: %s", idx, err)
		}
		lf.messagePattern = append(lf.messagePattern, pattern)
	}

	// pipeline 不是并发安全，无法支持多个 goroutine 使用同一个 pipeline 对象
	// 所以在此处使用 pool
	// 另，regexp 是并发安全的
	lf.pipelinePool = sync.Pool{
		New: func() interface{} {
			if lf.Pipeline == "" {
				return nil
			}

			// 即使 pipeline 配置错误，也不会影响全局
			p, err := pipeline.NewPipelineFromFile(filepath.Join(datakit.PipelineDir, lf.Pipeline))
			if err != nil {
				l.Debugf("new pipeline error: %s", err)
				return nil
			}
			return p
		},
	}

	return nil
}
func (lf *LogFilter) RunPipeline(message string) (map[string]interface{}, error) {
	pipe := lf.pipelinePool.Get()
	// pipe 为空指针（即没有配置 pipeline），将返回默认值
	if pipe == nil {
		return map[string]interface{}{"message": message}, nil
	}

	return pipe.(*pipeline.Pipeline).Run(message).Result()
}

func (lf *LogFilter) MatchMessage(message string) bool {
	for _, pattern := range lf.messagePattern {
		if pattern.MatchString(message) {
			return true
		}
	}
	return false
}

func (lf *LogFilter) MatchMultiline(message string) bool {
	if lf.multilinePattern == nil {
		return false
	}
	return lf.multilinePattern.MatchString(message)
}
