// Package tailer wraps logging file collection
package tailer

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/encoding"
)

const (
	// 定期寻找符合条件的新文件.
	scanNewFileInterval = time.Second * 10

	defaultSource   = "default"
	defaultMaxLines = 1000
)

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
	// 匹配正则表达式
	// 符合此正则匹配的数据，将被认定为有效数据。否则会累积追加到上一条有效数据的末尾
	// 例如 ^\d{4}-\d{2}-\d{2} 行首匹配 YYYY-MM-DD 时间格式
	//
	// 如果为空，则默认使用 ^\S 即匹配每行开始处非空白字符
	MultilineMatch string
	//  多行匹配的最大行数，避免出现某一行过长导致程序爆栈。默认 1000
	MultilineMaxLines int

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

	if opt.GlobalTags == nil {
		opt.GlobalTags = make(map[string]string)
	}

	if opt.MultilineMaxLines == 0 {
		opt.MultilineMaxLines = defaultMaxLines
	}

	opt.GlobalTags["service"] = opt.Service
	opt.log = logger.SLogger(opt.InputName)

	if _, err := encoding.NewDecoder(opt.CharacterEncoding); err != nil {
		return err
	}

	if _, err := NewMultiline(opt.MultilineMatch, opt.MultilineMaxLines); err != nil {
		return err
	}
	if filepath.Base(opt.Pipeline) != opt.Pipeline {
		return fmt.Errorf("invalid pipeline! the pipeline conf is file name like: nginx.p or xxx.p")
	}
	return nil
}

type Tailer struct {
	opt *Option

	fileList map[string]*Single

	filePatterns   []string
	ignorePatterns []string

	stop chan struct{}
	mu   sync.Mutex
	wg   sync.WaitGroup
}

func NewTailer(filePatterns []string, opt *Option, ignorePatterns ...[]string) (*Tailer, error) {
	if len(filePatterns) == 0 {
		return nil, fmt.Errorf("filePatterns is empty")
	}

	t := Tailer{
		opt:          opt,
		filePatterns: filePatterns,
		ignorePatterns: func() []string {
			if len(ignorePatterns) > 0 {
				return ignorePatterns[0]
			}
			return nil
		}(),
		fileList: make(map[string]*Single),
		stop:     make(chan struct{}),
	}

	if t.opt == nil {
		t.opt = &Option{}
	}

	if err := t.opt.init(); err != nil {
		return nil, err
	}
	return &t, nil
}

func (t *Tailer) Start() {
	ticker := time.NewTicker(scanNewFileInterval)
	defer ticker.Stop()

	// 立即执行一次，而不是等到tick到达
	t.scan()

	for {
		select {
		case <-t.stop:
			t.closeAll()
			t.removeAll()

			t.opt.log.Infof("waiting for all tailers to exit")
			t.wg.Wait()

			t.opt.log.Info("exit")
			return

		case <-ticker.C:
			t.scan()
			t.opt.log.Debugf("list of recivering: %v", t.getFileList())
		}
	}
}

func (t *Tailer) scan() {
	filelist, err := NewProvider().SearchFiles(t.filePatterns).IgnoreFiles(t.ignorePatterns).Result()
	if err != nil {
		t.opt.log.Warn(err)
	}

	t.cleanExpriedFile(filelist)

	for _, filename := range filelist {
		if t.fileInFileList(filename) {
			continue
		}
		t.wg.Add(1)
		go func(filename string) {
			defer t.wg.Done()
			defer t.removeFromFileList(filename)

			tl, err := NewTailerSingle(filename, t.opt)
			if err != nil {
				t.opt.log.Errorf("new tailer file %s error: %s", filename, err)
				return
			}

			t.addToFileList(filename, tl)

			tl.Run()
		}(filename)
	}
}

// cleanExpriedFile 清除过期文件，过期的定义包括被 remove/rename/truncate 导致文件不可用，其中 truncate 必须小于文件当前的 offset
// Tailer 已保存当前文件的列表（currentFileList），和函数参数 newFileList 比对，取 newFileList 对于 currentFileList 的差集，即为要被 clean 的对象.
func (t *Tailer) cleanExpriedFile(newFileList []string) {
	for _, oldFilename := range t.getFileList() {
		shouldClean := false

		tl := t.getTailerSingle(oldFilename)
		if tl != nil {
			didRotate, err := DidRotate(tl.file, tl.currentOffset())
			if err != nil {
				t.opt.log.Warnf("didRotate error: %s", err)
			}
			if didRotate {
				shouldClean = true
			}
		}

		func() {
			for _, newFilename := range newFileList {
				if oldFilename == newFilename {
					return
				}
			}
			shouldClean = true
		}()

		if shouldClean {
			t.closeFromFileList(oldFilename)
			t.opt.log.Debugf("maybe file %s already not exist or truncate, exit", oldFilename)
		}
	}
}

func (t *Tailer) Close() {
	select {
	case <-t.stop:
		// pass
	default:
		close(t.stop)
	}
}

func (t *Tailer) addToFileList(filename string, tl *Single) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.fileList[filename] = tl
}

func (t *Tailer) removeFromFileList(filename string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.fileList, filename)
}

func (t *Tailer) closeFromFileList(filename string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	tl, ok := t.fileList[filename]
	if !ok {
		return
	}
	tl.Close()
}

func (t *Tailer) fileInFileList(filename string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.fileList[filename]
	return ok
}

func (t *Tailer) getTailerSingle(filename string) *Single {
	t.mu.Lock()
	defer t.mu.Unlock()
	tl := t.fileList[filename]
	return tl
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

func (t *Tailer) closeAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, tl := range t.fileList {
		tl.Close()
	}
}

func (t *Tailer) removeAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for filename := range t.fileList {
		delete(t.fileList, filename)
	}
}
