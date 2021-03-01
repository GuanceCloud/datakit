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

const (
	pipelineTimeField = "time"
)

type tailer struct {
	tf *Tailf

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
	t := tailer{tf: tl, filename: filename, source: tl.Source}

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

	t.receiver()
}

func (t *tailer) receiver() {
	ticker := time.NewTicker(checkFileExistInterval)
	defer ticker.Stop()

	var line *tail.Line

	for {
		line = nil

		select {
		case <-datakit.Exit.Wait():
			t.tf.log.Debugf("Tailing source:%s, file %s is ending", t.source, t.filename)
			return

		case line, t.tailerOpen = <-t.tail.Lines:
			if !t.tailerOpen {
				t.channelOpen = false
			}

			if line != nil {
				t.tf.log.Debugf("get %d bytes from %s.%s", len(line.Text), t.source, t.filename)
			}

		case <-ticker.C:
			_, statErr := os.Lstat(t.filename)
			if os.IsNotExist(statErr) {
				t.tf.log.Warnf("check file %s is not exist", t.filename)
				return
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
		addStatus(fields)

		if err := io.NamedFeedEx(inputName, io.Logging, t.source, t.tags, fields, ts); err != nil {
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
				t.tf.log.Warnf("Tailing %s data channel is closed", t.filename)
				return
			}

			status = _continue
			return
		}
	}

	if line != nil && line.Err != nil {
		t.tf.log.Errorf("Tailing %q: %s", t.filename, line.Err.Error())
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
