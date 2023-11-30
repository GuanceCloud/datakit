// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tailer wraps logging file collection
package tailer

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

const (
	// 定期寻找符合条件的新文件.
	scanNewFileInterval = time.Second * 10
	defaultSource       = "default"
)

var l = logger.DefaultSLogger("socketLog")

type ForwardFunc func(filename, text string) error

type Option struct {
	// 忽略这些status，如果数据的status在此列表中，数据将不再上传
	// ex: "info"
	//     "debug"
	IgnoreStatus []string
	// logFile      []string
	Sockets []string
	// 默认值是 $Source + `_log`
	InputName string
	// 数据来源，默认值为'default'
	Source string
	// service，默认值为 $Source
	Service string
	// pipeline脚本路径，如果为空则不使用pipeline
	Pipeline string
	// 解释文件内容时所使用的的字符编码，如果设置为空，将不进行转码处理
	// ex: "utf-8"
	//     "utf-16le"
	//     "utf-16be"
	//     "gbk"
	//     "gb18030"
	//     "none"
	//     ""
	CharacterEncoding string

	// Depercated
	MaximumLength int

	// 匹配正则表达式
	// 符合此正则匹配的数据，将被认定为有效数据。否则会累积追加到上一条有效数据的末尾
	// 例如 ^\d{4}-\d{2}-\d{2} 行首匹配 YYYY-MM-DD 时间格式
	// 如果为空，则默认使用 ^\S 即匹配每行开始处非空白字符
	// 这是一个列表
	MultilinePatterns []string

	log *logger.Logger

	// 添加tag
	GlobalTags map[string]string
	// 是否从文件起始处开始读取
	// 注意，如果打开此项，可能会导致大量数据重复
	FromBeginning bool
	// 是否删除文本中的ansi转义码，默认为false，即不删除
	RemoveAnsiEscapeCodes bool
	// 是否关闭添加默认status字段列，包括status字段的固定转换行为，例如'd'->'debug'
	DisableAddStatusField bool
	// 是否关闭高频IO
	DisableHighFreqIODdata bool
	// 日志文本的另一种发送方式（和Feed冲突）
	ForwardFunc   ForwardFunc
	IgnoreDeadLog time.Duration

	// 是否开启阻塞发送模式
	BlockingMode bool

	MaxForceFlushLimit int

	MaxMultilineLifeDuration time.Duration

	Done <-chan interface{}

	Feeder dkio.Feeder

	Mode Mode
}

type Mode uint8

const (
	FileMode Mode = iota + 1
	DockerMode
	ContainerdMode
)

func (mode Mode) String() string {
	switch mode {
	case FileMode:
		return "file"
	case DockerMode:
		return "docker-json"
	case ContainerdMode:
		return "cri-log"
	}
	return "unknown"
}

func (opt *Option) Init() error {
	if opt.Source == "" {
		opt.Source = defaultSource
	}

	if opt.Service == "" {
		opt.Service = opt.Source
	}

	if opt.InputName == "" {
		opt.InputName = "logging/" + opt.Source
	}

	if opt.GlobalTags == nil {
		opt.GlobalTags = make(map[string]string)
	}

	if opt.Feeder == nil {
		opt.Feeder = dkio.DefaultFeeder()
	}

	if opt.MaxForceFlushLimit == 0 {
		opt.MaxForceFlushLimit = 10
	}

	opt.GlobalTags["service"] = opt.Service
	opt.log = logger.SLogger(opt.InputName)

	if opt.Pipeline != "" && filepath.Base(opt.Pipeline) != opt.Pipeline {
		opt.log.Warnf("invalid pipeline! the pipeline conf is file name like: nginx.p or xxx.p")
	}
	return nil
}

type Tailer struct {
	opt *Option

	fileList map[string]interface{}

	filePatterns   []string
	ignorePatterns []string

	stop chan interface{}
	mu   sync.Mutex
	g    *goroutine.Group
}

func NewTailer(filePatterns []string, opt *Option, ignorePatterns ...[]string) (*Tailer, error) {
	if len(filePatterns) == 0 {
		return nil, fmt.Errorf("filePatterns cannot be empty")
	}

	// copy tags, avoid to change the source tags
	sourceTags := opt.GlobalTags
	tags := make(map[string]string)
	for k, v := range sourceTags {
		tags[k] = v
	}
	opt.GlobalTags = tags

	t := Tailer{
		opt:          opt,
		filePatterns: filePatterns,
		ignorePatterns: func() []string {
			if len(ignorePatterns) > 0 {
				return ignorePatterns[0]
			}
			return nil
		}(),
		stop:     make(chan interface{}),
		fileList: make(map[string]interface{}),
		g:        goroutine.NewGroup(goroutine.Option{Name: "tailer"}),
	}

	if t.opt == nil {
		t.opt = &Option{}
	}
	if err := t.opt.Init(); err != nil {
		return nil, err
	}

	return &t, nil
}

func (t *Tailer) Start() {
	ticker := time.NewTicker(scanNewFileInterval)
	defer ticker.Stop()

	for {
		if t.scan() {
			t.opt.log.Infof("all tailers end...")
			_ = t.g.Wait()
			t.opt.log.Info("all exit")
			return
		}

		t.opt.log.Debugf("list of recivering: %v", t.getFileList())

		select {
		case <-t.stop:
			t.opt.log.Infof("waiting for all tailers to exit")
			_ = t.g.Wait()
			t.opt.log.Info("all exit")
			return

		case <-ticker.C:
		}
	}
}

func (t *Tailer) scan() (ended bool) {
	filelist, err := NewProvider().SearchFiles(t.filePatterns).IgnoreFiles(t.ignorePatterns).Result()
	if err != nil {
		t.opt.log.Warn(err)
	}

	for _, filename := range filelist {
		if !t.opt.FromBeginning {
			if t.opt.IgnoreDeadLog > 0 && !FileIsActive(filename, t.opt.IgnoreDeadLog) {
				continue
			}
		}

		if t.inFileList(filename) {
			continue
		}

		t.opt.log.Infof("new logging file %s with source %s", filename, t.opt.Source)

		func(filename string) {
			g.Go(func(ctx context.Context) error {
				defer t.removeFromFileList(filename)

				tl, err := NewTailerSingle(filename, t.opt)
				if err != nil {
					t.opt.log.Errorf("new tailer file %s error: %s", filename, err)
					return nil
				}

				t.addToFileList(filename)

				tl.Run()
				return nil
			})
		}(filename)
	}

	return false
}

func (t *Tailer) Close() {
	select {
	case <-t.stop:
		return
	default:
		close(t.stop)
	}
}

func (t *Tailer) addToFileList(filename string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.fileList[filename] = nil
}

func (t *Tailer) removeFromFileList(filename string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.fileList, filename)
}

func (t *Tailer) inFileList(filename string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.fileList[filename]
	return ok
}

func (t *Tailer) getFileList() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	var list []string
	for filename := range t.fileList {
		list = append(list, filename)
	}
	return list
}

//nolint:gochecknoinits
func init() {
	setupMetrics()
}
