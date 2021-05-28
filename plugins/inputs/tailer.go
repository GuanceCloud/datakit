package inputs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/glob"
	"github.com/hpcloud/tail"
	"github.com/mattn/go-zglob"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

// // example_01:
//
//     tailer, err := NewTailer(tailerOption_point)
//     if err != nil {
//         return err
//     }
//     go tailer.Run()
//
// 详见 tailf/inputs.go

const (
	// 定期寻找符合条件的新文件
	findNewFileInterval = time.Second * 10

	// 定期检查当前文件是否存在
	checkFileExistInterval = time.Minute * 10

	// pipeline关键字段
	pipelineTimeField = "time"

	// ES value can be at most 32766 bytes long
	maxFieldsLength = 32766

	defaultSource = "default"

	defaultMatch = `^\S`

	// TODO: 传参决定数据上传到哪个category
	OPT_METRIC  = 1
	OPT_LOGGING = 2
)

type TailerOption struct {
	// 需要采集的文件路径列表
	Files []string `toml:"files"`

	// glob 忽略的文件路径
	IgnoreFiles []string `toml:"ignore"`

	// 数据来源，默认值为'default'
	Source string `toml:"source"`

	// service，默认值为 $Source
	Service string `toml:"service"`

	// pipeline脚本路径，如果为空则不使用pipeline
	Pipeline string `toml:"pipeline"`

	// 忽略这些status，如果数据的status在此列表中，数据将不再上传
	// ex: "info"
	//     "debug"
	IgnoreStatus []string `toml:"ignore_status"`

	// 是否从文件起始处开始读取
	// 注意，如果打开此项，可能会导致大量数据重复
	FromBeginning bool `toml:"-"`

	// 解释文件内容时所使用的的字符编码，如果设置为空，将不进行转码处理
	// ex: "utf-8"
	//     "utf-16le"
	//     "utf-16be"
	//     "gbk"
	//     "gb18030"
	//     ""
	CharacterEncoding string `toml:"character_encoding"`

	// 匹配正则表达式
	// 符合此正则匹配的数据，将被认定为有效数据。否则会累积追加到上一条有效数据的末尾
	// 例如 ^\d{4}-\d{2}-\d{2} 行首匹配 YYYY-MM-DD 时间格式
	//
	// 如果为空，则默认使用 ^\S 即匹配每行开始处非空白字符
	Match string `toml:"match"`

	// 是否关闭添加默认status字段列，包括status字段的固定转换行为，例如'd'->'debug'
	DisableAddStatusField bool `toml:"-"`

	// 是否关闭高频IO
	DisableHighFreqIODdata bool `toml:"-"`

	// TODO:
	// 上传到该category，默认是 OPT_LOGGING
	Category int `toml:"-"`

	// 自定义添加tag，默认会添加 "filename"
	Tags map[string]string `toml:"tags"`

	StopChan chan interface{} `toml:"-"`
}

func fillTailerOption(opt *TailerOption) *TailerOption {
	if opt == nil {
		opt = &TailerOption{}
	}

	if opt.Source == "" {
		opt.Source = defaultSource
	}

	if opt.Service == "" {
		opt.Service = opt.Source
	}

	if opt.Match == "" {
		opt.Match = defaultMatch
	}

	if opt.Tags == nil {
		opt.Tags = make(map[string]string)
	}
	opt.Tags["service"] = opt.Service

	if opt.StopChan == nil {
		opt.StopChan = make(chan interface{})
	}

	return opt
}

type Tailer struct {
	Option    *TailerOption
	InputName string

	tailConf tail.Config

	decoder           *encoding.Decoder
	multilineInstance *Multiline
	watcher           *Watcher

	log *logger.Logger
	wg  sync.WaitGroup
}

func NewTailer(opt *TailerOption) (*Tailer, error) {
	var tailer = Tailer{Option: opt}

	if err := tailer.init(); err != nil {
		return nil, err
	}

	return &tailer, nil
}

func (t *Tailer) init() error {
	var err error

	if len(t.Option.Files) == 0 {
		return fmt.Errorf("files is nil")
	}

	t.Option = fillTailerOption(t.Option)

	var seek *tail.SeekInfo
	if !t.Option.FromBeginning {
		seek = &tail.SeekInfo{
			Whence: 2, // seek is 2
			Offset: 0,
		}
	}
	t.tailConf = tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  seek,
		MustExist: true,
		Poll:      false, // default watch method is "inotify"
		Logger:    tail.DiscardingLogger,
	}

	if t.decoder, err = NewDecoder(t.Option.CharacterEncoding); err != nil {
		return err
	}

	multilineConfig := &MultilineConfig{
		Pattern:        t.Option.Match,
		InvertMatch:    true,
		MatchWhichLine: "previous",
	}
	if t.multilineInstance, err = multilineConfig.NewMultiline(); err != nil {
		return err
	}

	if t.watcher, err = NewWatcher(); err != nil {
		return err
	}

	if t.InputName == "" {
		t.InputName = t.Option.Source + "_log"
	}

	t.log = logger.SLogger(t.InputName)

	return nil
}

func (t *Tailer) Run() {
	t.log.Infof("tail %s input started", t.InputName)

	ticker := time.NewTicker(findNewFileInterval)
	defer ticker.Stop()

	ctx, watcherCancel := context.WithCancel(context.Background())
	go t.watcher.Watching(ctx)

	// 立即执行一次，而不是等到tick到达
	t.do()

	// 如果开启 FromBeginning 选项，将停止 ticker，此 case 分支将不再进入，也不再发现新文件
	if t.Option.FromBeginning {
		ticker.Stop()
	}

	for {
		select {
		case <-t.Option.StopChan:
			t.log.Infof("waiting for all tailers to exit")
			watcherCancel()
			t.watcher.Close()
			t.wg.Wait()
			t.log.Info("exit")
			return

		case <-ticker.C:
			t.do()
		}
	}
}

func (t *Tailer) do() {
	fileList := getFileList(t.Option.Files, t.Option.IgnoreFiles)

	for _, filename := range fileList {
		if exist := t.watcher.IsExist(filename); exist {
			continue
		}
		t.wg.Add(1)
		go func(fn string) {
			defer t.wg.Done()
			t.tailingFile(fn)
		}(filename)
	}
}

func (t *Tailer) Close() error {
	select {
	case <-t.Option.StopChan:
		// pass
	default:
		close(t.Option.StopChan)
	}
	return nil
}

func (t *Tailer) tailingFile(filename string) {
	t.log.Debugf("start tail, %s", filename)

	instence, err := newTailerSingle(t, filename)
	if err != nil {
		t.log.Errorf("new tailer %s error: %s", filename, err)
	}
	defer instence.clean()

	ctx, cancel := context.WithCancel(context.Background())
	if err := t.watcher.Add(filename, cancel); err != nil {
		t.log.Warnf("add watcher file %s error: %s", filename, err)
	}

	// 阻塞
	instence.receiving(ctx)

	if err := t.watcher.Remove(filename); err != nil {
		t.log.Warnf("remove watcher file %s error: %s", filename, err)
	}

	t.log.Debugf("remove file %s from the running list", filename)
}

// tailerSingle
// 每一个文件有自己独立的tailerSingle
// tailerSingle 共用来自父类 Tailer 的成员变量，包括 Option、Multiline 等
type tailerSingle struct {
	tl *Tailer

	filename string
	tags     map[string]string

	tail *tail.Tail
	pipe *pipeline.Pipeline

	textLine    bytes.Buffer
	tailerOpen  bool
	channelOpen bool
}

func newTailerSingle(tl *Tailer, filename string) (*tailerSingle, error) {
	t := &tailerSingle{
		tl:          tl,
		filename:    filename,
		tailerOpen:  true,
		channelOpen: true,
	}

	t.tags = func() map[string]string {
		var m = make(map[string]string)

		for k, v := range tl.Option.Tags {
			m[k] = v
		}

		if _, ok := m["filename"]; !ok {
			m["filename"] = filepath.Base(filename)
		}
		return m
	}()

	var err error

	t.tail, err = tail.TailFile(t.filename, t.tl.tailConf)
	if err != nil {
		return nil, err
	}

	if t.tl.Option.Pipeline != "" {
		t.pipe, err = pipeline.NewPipelineFromFile(t.tl.Option.Pipeline)
		if err != nil {
			t.tl.log.Warnf("failed of pipeline, err: %s", err)
		}
	}

	return t, nil
}

func (t *tailerSingle) clean() {
	if t.tail != nil {
		t.tail.Cleanup()
	}
}

func (t *tailerSingle) receiving(ctx context.Context) {
	t.tl.log.Debugf("start recivering data from the file %s", t.filename)

	var line *tail.Line

	for {
		line = nil

		select {
		case <-ctx.Done():
			t.tl.log.Debugf("tailing source:%s, file %s is ending", t.tl.Option.Source, t.filename)
			return

		case line, t.tailerOpen = <-t.tail.Lines:
			if !t.tailerOpen {
				t.channelOpen = false
			}

			if line != nil {
				t.tl.log.Debugf("get %d bytes from source:%s file:%s", len(line.Text), t.tl.Option.Source, t.filename)
			}
		}

		text, status := t.multiline(line)
		switch status {
		case _return:
			return
		case _continue:
			continue
		case _next:
			//pass
		}

		var err error

		text, err = t.decode(text)
		if err != nil {
			t.tl.log.Errorf("decode error, %s", err)
			continue
		}

		var fields = make(map[string]interface{})

		if t.pipe != nil {
			fields, err = t.pipe.Run(text).Result()
			if err != nil {
				// 当pipe.Run() err不为空时，fields含有message字段
				// 等同于fields["message"] = text
				t.tl.log.Errorf("run pipeline error, %s", err)
			}
		} else {
			fields["message"] = text
		}

		if len(fields) == 0 {
			l.Debug("fields is empty, maybe the use of pipeline")
			continue
		}

		// 检查数据是否过长
		// 只有在碰到非 message 字段，且长度超过最大限制时才会返回 error
		// 防止通过 pipeline 添加巨长字段的恶意行为
		if err := checkFieldsLength(fields, maxFieldsLength); err != nil {
			t.tl.log.Error(err)
			continue
		}

		// 添加默认 status 和 status 映射
		if !t.tl.Option.DisableAddStatusField {
			addStatus(fields)
		}

		// 过滤指定status
		if status, ok := fields["status"].(string); ok {
			if contains(t.tl.Option.IgnoreStatus, status) {
				continue
			}
		}

		// 提取 time 字段
		ts, err := takeTime(fields)
		if err != nil {
			ts = time.Now()
			t.tl.log.Error(err)
		}

		// 使用 source 当做行协议的 measurement
		// 使用 inputName 用作 io 上传来源，Tailer inputName 默认为 source+"_log"
		pt, err := io.MakePoint(t.tl.Option.Source, t.tags, fields, ts)
		if err != nil {
			t.tl.log.Error(err)
		} else {
			if err := io.Feed(
				t.tl.InputName,
				datakit.Logging,
				[]*io.Point{pt},
				&io.Option{HighFreq: !t.tl.Option.DisableHighFreqIODdata},
			); err != nil {
				t.tl.log.Error(err)
			}
		}
	}
}

type multilineStatus int

const (
	// tail channel 关闭，执行 return
	_return multilineStatus = iota
	// multiline 判断数据为多行，将数据存入缓存，继续读取下一行
	_continue
	// multiline 判断多行数据结束，将缓存中的数据放出，继续执行后续处理
	_next
)

func (t *tailerSingle) multiline(line *tail.Line) (text string, status multilineStatus) {
	if line != nil {
		text = strings.TrimRight(line.Text, "\r")

		if t.tl.multilineInstance.IsEnabled() {
			if text = t.tl.multilineInstance.ProcessLine(text, &t.textLine); text == "" {
				status = _continue
				return
			}
		}
	}

	if line == nil || !t.channelOpen || !t.tailerOpen {
		if text += t.tl.multilineInstance.Flush(&t.textLine); text == "" {
			if !t.channelOpen {
				status = _return
				t.tl.log.Warnf("tailing %s data channel is closed", t.filename)
				return
			}

			status = _continue
			return
		}
	}

	if line != nil && line.Err != nil {
		t.tl.log.Errorf("tailing %q: %s", t.filename, line.Err.Error())
		status = _continue
		return
	}

	status = _next
	return
}

func (t *tailerSingle) decode(text string) (str string, err error) {
	return t.tl.decoder.String(text)
}

func takeTime(fields map[string]interface{}) (ts time.Time, err error) {
	// time should be nano-second
	if v, ok := fields[pipelineTimeField]; ok {
		nanots, ok := v.(int64)
		if !ok {
			err = fmt.Errorf("invalid filed `%s: %v', should be nano-second, but got `%s'",
				pipelineTimeField, v, reflect.TypeOf(v).String())
			return
		}

		ts = time.Unix(nanots/int64(time.Second), nanots%int64(time.Second))
		delete(fields, pipelineTimeField)
	} else {
		ts = time.Now()
	}

	return
}

// checkFieldsLength 指定字段长度 "小于等于" maxlength
func checkFieldsLength(fields map[string]interface{}, maxlength int) error {
	for k, v := range fields {
		switch vv := v.(type) {
		// FIXME:
		// need  "case []byte" ?
		case string:
			if len(vv) <= maxlength {
				continue
			}
			if k == "message" {
				fields[k] = vv[:maxlength]
			} else {
				return fmt.Errorf("fields: %s, length=%d, out of maximum length", k, len(vv))
			}
		default:
			// nil
		}
	}
	return nil
}

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

func addStatus(fields map[string]interface{}) {
	// map 有 "status" 字段
	statusField, ok := fields["status"]
	if !ok {
		fields["status"] = "info"
		return
	}
	// "status" 类型必须是 string
	statusStr, ok := statusField.(string)
	if !ok {
		fields["status"] = "info"
		return
	}

	// 查询 statusMap 枚举表并替换
	if v, ok := statusMap[strings.ToLower(statusStr)]; !ok {
		fields["status"] = "info"
	} else {
		fields["status"] = v
	}
}

//
// ============================= multiline ==================================
//

type Multiline struct {
	config        *MultilineConfig
	enabled       bool
	patternRegexp *regexp.Regexp
}

type MultilineConfig struct {
	Pattern        string
	MatchWhichLine string
	InvertMatch    bool
}

const (
	// Previous => Append current line to previous line
	Previous = "previous"
	// Next => Next line will be appended to current line
	Next = "next"
)

func (m *MultilineConfig) NewMultiline() (*Multiline, error) {
	enabled := false
	var r *regexp.Regexp
	var err error

	if m.Pattern != "" {
		enabled = true
		if r, err = regexp.Compile(m.Pattern); err != nil {
			return nil, err
		}

		if m.MatchWhichLine != Previous && m.MatchWhichLine != Next {
			m.MatchWhichLine = Previous
		}
	}

	return &Multiline{
		config:        m,
		enabled:       enabled,
		patternRegexp: r,
	}, nil
}

func (m *Multiline) IsEnabled() bool {
	return m.enabled
}
func (m *Multiline) ProcessLine(text string, buffer *bytes.Buffer) string {
	if m.matchString(text) {
		buffer.WriteString("\n")
		buffer.WriteString(text)
		return ""
	}

	if m.config.MatchWhichLine == Previous {
		previousText := buffer.String()
		buffer.Reset()
		buffer.WriteString(text)
		text = previousText
	} else {
		// Next
		if buffer.Len() > 0 {
			buffer.WriteString(text)
			text = buffer.String()
			buffer.Reset()
		}
	}

	return text
}
func (m *Multiline) Flush(buffer *bytes.Buffer) string {
	if buffer.Len() == 0 {
		return ""
	}
	text := buffer.String()
	buffer.Reset()
	return text
}

func (m *Multiline) matchString(text string) bool {
	return m.patternRegexp.MatchString(text) != m.config.InvertMatch
}

//
// ========================== watcher =========================
//

type Watcher struct {
	watcher *fsnotify.Watcher
	list    sync.Map
}

func NewWatcher() (*Watcher, error) {
	var err error
	var f = &Watcher{}

	f.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (w *Watcher) Close() error {
	w.list.Range(func(key, value interface{}) bool {
		if cancel, ok := value.(context.CancelFunc); ok {
			cancel()
		}
		return true
	})
	return w.watcher.Close()
}

func (w *Watcher) Add(filename string, cancel context.CancelFunc) error {
	if err := w.watcher.Add(filename); err != nil {
		return err
	}
	w.list.Store(filename, cancel)
	return nil
}

func (w *Watcher) Remove(filename string) error {
	w.list.Delete(filename)
	return w.watcher.Remove(filename)
}

func (w *Watcher) cancelCtx(filename string) {
	value, ok := w.list.Load(filename)
	if !ok {
		return
	}
	if cancel, ok := value.(context.CancelFunc); ok {
		cancel()
	}
}

func (w *Watcher) IsExist(file string) bool {
	_, ok := w.list.Load(file)
	return ok
}

func (w *Watcher) Watching(ctx context.Context) {
	var checkFileExistInterval = time.Second * 5
	tick := time.NewTicker(checkFileExistInterval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				continue
			}
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				w.cancelCtx(event.Name)
			}

		case <-tick.C:
			// 为什么不使用 notify 的方式监控文件删除，反而采用 Lstat() ？
			//
			// notify 只有当文件引用计数为 0 时，才会认为此文件已经被删除，从而触发 remove event
			// 在此处，datakit 打开文件后保存句柄，即使 rm 删除文件，该文件的引用计数依旧是 1，因为 datakit 在占用
			// 从而导致，datakit 占用文件无法删除，无法删除就收不到 remove event，此 goroutine 就会长久存在
			// 且极端条件下，长时间运行，可能会导致磁盘容量不够的情况，因为占用容量的文件在此被引用，新数据无法覆盖
			// 以上结论仅限于 linux
			w.list.Range(func(key, value interface{}) bool {
				filename, ok := key.(string)
				if !ok {
					return true
				}

				_, statErr := os.Lstat(filename)
				if os.IsNotExist(statErr) {
					if cancel, ok := value.(context.CancelFunc); ok {
						cancel()
					}
					return true
				}

				return true
			})

		case _, ok := <-w.watcher.Errors:
			if !ok {
				continue
			}
		}
	}
}

//
// ========================== decode =========================
//

func NewDecoder(enc string) (*encoding.Decoder, error) {
	switch enc {
	case "utf-8":
		return unicode.UTF8.NewDecoder(), nil
	case "utf-16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder(), nil
	case "utf-16be":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder(), nil
	case "gbk":
		return simplifiedchinese.GBK.NewDecoder(), nil
	case "gb18030":
		return simplifiedchinese.GB18030.NewDecoder(), nil
	case "none", "":
		return encoding.Nop.NewDecoder(), nil
	}
	return nil, errors.New("unknown character encoding")
}

//
// ========================= unit =============================
//

func checkPipeLine(path string) error {
	if path == "" {
		return nil
	}
	_, err := pipeline.NewPipelineFromFile(path)
	return err
}
func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func contains(array []string, str string) bool {
	for _, a := range array {
		if a == str {
			return true
		}
	}
	return false
}

func getFileList(filesGlob, ignoreGlob []string) []string {
	var matches = make(map[string]interface{})

	var filesMatches []string
	for _, f := range filesGlob {
		matche, err := zglob.Glob(f)
		if err != nil {
			continue
		}
		filesMatches = append(filesMatches, matche...)
	}

	var ignores []glob.Glob
	for _, ig := range ignoreGlob {
		g, err := glob.Compile(ig)
		if err != nil {
			continue
		}
		ignores = append(ignores, g)
	}

	for _, f := range filesMatches {
		for _, g := range ignores {
			if g.Match(f) {
				goto next
			}
		}
		matches[f] = nil
	next:
	}

	// unique
	var list []string
	for matche := range matches {
		list = append(list, matche)
	}

	return list
}
