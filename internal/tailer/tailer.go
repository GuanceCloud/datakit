package tailer

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/encoding"
)

const (
	// 定期寻找符合条件的新文件
	scannsNewFileInterval = time.Second * 10

	// 定期检查当前文件是否存在
	checkFileExistInterval = time.Minute * 10

	defaultSource = "default"

	defaultMatch = `^\S`
)

type Option struct {
	// 默认值是 $Source + `_log`
	InputName string

	// 数据来源，默认值为'default'
	Source string

	// service，默认值为 $Source
	Service string

	// pipeline脚本路径，如果为空则不使用pipeline
	Pipeline string

	// 忽略这些status，如果数据的status在此列表中，数据将不再上传
	// ex: "info"
	//     "debug"
	IgnoreStatus []string

	// 是否从文件起始处开始读取
	// 注意，如果打开此项，可能会导致大量数据重复
	FromBeginning bool

	// 解释文件内容时所使用的的字符编码，如果设置为空，将不进行转码处理
	// ex: "utf-8"
	//     "utf-16le"
	//     "utf-16be"
	//     "gbk"
	//     "gb18030"
	//     "none"
	//     ""
	CharacterEncoding string

	// 匹配正则表达式
	// 符合此正则匹配的数据，将被认定为有效数据。否则会累积追加到上一条有效数据的末尾
	// 例如 ^\d{4}-\d{2}-\d{2} 行首匹配 YYYY-MM-DD 时间格式
	//
	// 如果为空，则默认使用 ^\S 即匹配每行开始处非空白字符
	Match string

	// 是否关闭添加默认status字段列，包括status字段的固定转换行为，例如'd'->'debug'
	DisableAddStatusField bool

	// 是否关闭高频IO
	DisableHighFreqIODdata bool

	// 添加tag
	GlobalTags map[string]string

	done chan string
	log  *logger.Logger
}

func (opt *Option) init() error {
	if opt.Source == "" {
		opt.Source = defaultSource
	}

	if opt.Service == "" {
		opt.Service = opt.Source
	}

	if opt.InputName == "" {
		opt.InputName = opt.Source + "_log"
	}

	if opt.Match == "" {
		opt.Match = defaultMatch
	}

	if opt.GlobalTags == nil {
		opt.GlobalTags = make(map[string]string)
	}

	opt.GlobalTags["service"] = opt.Service
	opt.done = make(chan string)
	opt.log = logger.SLogger(opt.InputName)

	var err error
	_, err = encoding.NewDecoder(opt.CharacterEncoding)
	_, err = NewMultiline(opt.Match)

	return err
}

type Tailer struct {
	opt     *Option
	watcher *Watcher

	pathNames       []string
	ignorePathNames []string

	stop chan interface{}
}

func NewTailer(pathNames []string, opt *Option, ignorePathNames ...[]string) (*Tailer, error) {
	if len(pathNames) == 0 {
		return nil, fmt.Errorf("pathNames is empty")
	}

	t := Tailer{
		opt:       opt,
		pathNames: pathNames,
		stop:      make(chan interface{}),
	}

	if t.opt == nil {
		t.opt = &Option{}
	}
	if len(ignorePathNames) > 0 {
		t.ignorePathNames = ignorePathNames[0]
	}

	var err error
	t.watcher, err = NewWatcher()
	if err != nil {
		return nil, err
	}

	if err = t.opt.init(); err != nil {
		return nil, err
	}

	return &t, nil
}

func (t *Tailer) Start() {
	ticker := time.NewTicker(scannsNewFileInterval)
	defer ticker.Stop()

	ctx, watcherCancel := context.WithCancel(context.Background())
	go t.watcher.Watching(ctx)

	// 立即执行一次，而不是等到tick到达
	t.do()

	for {
		select {
		case name := <-t.opt.done:
			t.watcher.Remove(name)
			t.opt.log.Debugf("tailer %s exit", name)
		case <-t.stop:
			watcherCancel()
			t.watcher.Close()
			t.opt.log.Infof("waiting for all tailers to exit")
			t.opt.log.Info("exit")
			return

		case <-ticker.C:
			t.opt.log.Debugf("list of recivering: %v", t.watcher.List())
			t.do()
		}
	}
}

func (t *Tailer) do() {
	fileList := NewFileList(t.pathNames).Ignore(t.ignorePathNames).List()

	for _, filename := range fileList {
		if exist := t.watcher.IsExist(filename); exist {
			continue
		}

		tl, err := NewTailerSingle(filename, t.opt)
		if err != nil {
			t.opt.log.Errorf("new tailer file %s error: %s", filename, err)
			continue
		}

		if err := t.watcher.Add(filename, tl); err != nil {
			t.opt.log.Error("add watcher file %s error: %s", filename, err)
			// add 失败将不运行此 tailer，避免出现未在册记录的行为
			continue
		}

		tl.Start()
	}
}

func (t *Tailer) Close() error {
	select {
	case <-t.stop:
		// pass
	default:
		close(t.stop)
	}
	return nil
}
