package tailf

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/hpcloud/tail"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

type tailer struct {
	tf         *Tailf
	notifyChan chan notifyType

	filename string
	source   string
	tags     map[string]string

	tail *tail.Tail
	pipe *pipeline.Pipeline

	textLine    bytes.Buffer
	tailerOpen  bool
	channelOpen bool
}

func newTailer(tl *Tailf, filename string) *tailer {
	t := tailer{
		tf:         tl,
		filename:   filename,
		source:     tl.Source,
		notifyChan: make(chan notifyType),
	}

	t.tags = func() map[string]string {
		var m = make(map[string]string)

		for k, v := range tl.Tags {
			m[k] = v
		}

		if _, ok := m["filename"]; !ok {
			m["filename"] = filename
		}
		return m
	}()

	t.tailerOpen = true
	t.channelOpen = true

	return &t
}

func (t *tailer) getNotifyChan() chan notifyType {
	return t.notifyChan
}

func (t *tailer) run() {
	var err error

	t.tail, err = tail.TailFile(t.filename, t.tf.tailerConf)
	if err != nil {
		t.tf.log.Error("failed of build tailer, err:%s", err)
		return
	}
	defer t.tail.Cleanup()

	if t.tf.Pipeline != "" {
		t.pipe, err = pipeline.NewPipelineFromFile(t.tf.Pipeline)
		if err != nil {
			t.tf.log.Error("failed of pipeline, err:%s", err)
			return
		}
	}

	t.receiving()
}

func (t *tailer) receiving() {
	t.tf.log.Debugf("start recivering data from the file %s", t.filename)

	ticker := time.NewTicker(checkFileExistInterval)
	defer ticker.Stop()

	var line *tail.Line

	for {
		line = nil

		// FIXME: 4个case是否过多？
		select {
		case <-datakit.Exit.Wait():
			t.tf.log.Debugf("tailing source:%s, file %s is ending", t.source, t.filename)
			return

		case n := <-t.notifyChan:
			switch n {
			case renameNotify:
				t.tf.log.Warnf("file %s was rename", t.filename)
				return
			default:
				// nil
			}

		// 为什么不使用 notify 的方式监控文件删除，反而采用 Lstat() ？
		//
		// notify 只有当文件引用计数为 0 时，才会认为此文件已经被删除，从而触发 remove event
		// 在此处，datakit 打开文件后保存句柄，即使 rm 删除文件，该文件的引用计数依旧是 1，因为 datakit 在占用
		// 从而导致，datakit 占用文件无法删除，无法删除就收不到 remove event，此 goroutine 就会长久存在
		// 且极端条件下，长时间运行，可能会导致磁盘容量不够的情况，因为占用容量的文件在此被引用，新数据无法覆盖
		// 以上结论仅限于 linux

		case <-ticker.C:
			_, statErr := os.Lstat(t.filename)
			if os.IsNotExist(statErr) {
				t.tf.log.Warnf("file %s is not exist", t.filename)
				return
			}

		case line, t.tailerOpen = <-t.tail.Lines:
			if !t.tailerOpen {
				t.channelOpen = false
			}

			if line != nil {
				t.tf.log.Debugf("get %d bytes from source:%s file:%s", len(line.Text), t.source, t.filename)
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
			t.tf.log.Errorf("decode error, %s", err)
			continue
		}

		var fields = make(map[string]interface{})

		if t.pipe != nil {
			fields, err = t.pipe.Run(text).Result()
			if err != nil {
				// 当pipe.Run() err不为空时，fields含有message字段
				// 等同于fields["message"] = text
				t.tf.log.Errorf("run pipeline error, %s", err)
			}
		} else {
			fields["message"] = text
		}

		ts, err := takeTime(fields)
		if err != nil {
			ts = time.Now()
			t.tf.log.Errorf("%s", err)
		}

		if err := checkFieldsLength(fields, maxFieldsLength); err != nil {
			// 只有在碰到非 message 字段，且长度超过最大限制时才会返回 error
			// 防止通过 pipeline 添加巨长字段的恶意行为
			t.tf.log.Error(err)
			continue
		}

		addStatus(fields)

		// use t.source as input-name, make it more distinguishable for multiple tailf instances
		if err := io.HighFreqFeedEx(inputName, io.Logging, t.source, t.tags, fields, ts); err != nil {
			t.tf.log.Error(err)
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

func (t *tailer) multiline(line *tail.Line) (text string, status multilineStatus) {
	if line != nil {
		text = strings.TrimRight(line.Text, "\r")

		if t.tf.multiline.IsEnabled() {
			if text = t.tf.multiline.ProcessLine(text, &t.textLine); text == "" {
				status = _continue
				return
			}
		}
	}

	if line == nil || !t.channelOpen || !t.tailerOpen {
		if text += t.tf.multiline.Flush(&t.textLine); text == "" {
			if !t.channelOpen {
				status = _return
				t.tf.log.Warnf("tailing %s data channel is closed", t.filename)
				return
			}

			status = _continue
			return
		}
	}

	if line != nil && line.Err != nil {
		t.tf.log.Errorf("tailing %q: %s", t.filename, line.Err.Error())
		status = _continue
		return
	}

	status = _next
	return
}

func (t *tailer) decode(text string) (str string, err error) {
	return t.tf.decoder.String(text)
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
