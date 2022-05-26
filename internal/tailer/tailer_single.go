// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pborman/ansi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/encoding"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/multiline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
)

const (
	defaultSleepDuration = time.Second
	readBuffSize         = 1024 * 4
	timeoutDuration      = time.Second * 3
)

type Single struct {
	opt                *Option
	file               *os.File
	watcher            *fsnotify.Watcher
	filepath, filename string

	decoder *encoding.Decoder
	mult    *multiline.Multiline

	readBuff []byte

	tags   map[string]string
	stopCh chan struct{}
}

func NewTailerSingle(filename string, opt *Option) (*Single, error) {
	if opt == nil {
		return nil, fmt.Errorf("option cannot be null pointer")
	}

	t := &Single{
		stopCh: make(chan struct{}, 1),
		opt:    opt,
	}

	var err error
	if opt.CharacterEncoding != "" {
		t.decoder, err = encoding.NewDecoder(opt.CharacterEncoding)
		if err != nil {
			return nil, err
		}
	}
	t.mult, err = multiline.New(opt.MultilineMatch, opt.MultilineMaxLines)
	if err != nil {
		return nil, err
	}

	t.file, err = os.Open(filename) //nolint:gosec
	if err != nil {
		if os.IsNotExist(err) {
			filename = filepath.Join("/rootfs", filename)
			t.file, err = os.Open(filename) //nolint:gosec
			if err != nil {
				return nil, err
			}
		}
		return nil, err
	}

	if !opt.FromBeginning {
		if _, err := t.file.Seek(0, os.SEEK_END); err != nil {
			return nil, err
		}
	}

	t.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = t.watcher.Add(filename)
	if err != nil {
		return nil, err
	}

	t.readBuff = make([]byte, readBuffSize)
	t.filepath = t.file.Name()
	t.filename = filepath.Base(t.filepath)
	t.tags = t.buildTags(opt.GlobalTags)

	return t, nil
}

func (t *Single) Run() {
	defer t.Close()
	t.forwardMessage()
}

func (t *Single) Close() {
	t.stopCh <- struct{}{}
	t.closeWatcher()
	t.opt.log.Infof("closing %s", t.filepath)
}

func (t *Single) addWatcher(fn string) error {
	if t.watcher == nil {
		return nil
	}
	return t.watcher.Add(fn)
}

func (t *Single) removeWatcher(fn string) error {
	if t.watcher == nil {
		return nil
	}
	return t.watcher.Remove(fn)
}

func (t *Single) closeWatcher() {
	if t.watcher == nil {
		return
	}
	if err := t.watcher.Close(); err != nil {
		t.opt.log.Warnf("close watcher err: %s, ignored", err)
	}
}

func (t *Single) closeFile() {
	if t.file != nil {
		return
	}
	if err := t.file.Close(); err != nil {
		t.opt.log.Warnf("close file err: %s, ignored", err)
	}
	t.file = nil
}

func (t *Single) reopen() error {
	t.closeFile()
	t.opt.log.Debugf("reopen file %s", t.filepath)

	for {
		var err error
		t.file, err = os.Open(t.filepath)
		if err != nil {
			if os.IsNotExist(err) {
				t.opt.log.Debugf("waiting for %s to appear..", t.filepath)
				time.Sleep(time.Second)
				continue
			}
			return fmt.Errorf("unable to open file %s: %w", t.filepath, err)
		}
		break
	}

	return nil
}

func (t *Single) tellEvent(event fsnotify.Event) (err error) {
	switch {
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		t.opt.log.Debugf("receive remove event from file %s", t.filepath)

		if err = t.reopen(); err != nil {
			t.opt.log.Warnf("failed to reopen file %s, err: %s", t.filepath, err)
			return
		}
		if err = t.removeWatcher(t.filepath); err != nil {
			t.opt.log.Warnf("unable remove watcher %s, err: %s", t.filepath, err)
			return
		}
		if err = t.addWatcher(t.filepath); err != nil {
			t.opt.log.Warnf("unable add watcher %s, err: %s", t.filepath, err)
			return
		}

	case event.Op&fsnotify.Rename == fsnotify.Rename:
		t.opt.log.Debugf("receive rename event from file %s", t.filepath)

		if err = t.reopen(); err != nil {
			t.opt.log.Warnf("failed to reopen file %s, err: %s", t.filepath, err)
			return
		}

	default:
		t.opt.log.Debugf("receive %s event from file %s, ignored", event, t.filepath)
	}

	return
}

//nolint:cyclop
func (t *Single) forwardMessage() {
	var (
		b       = &buffer{}
		timeout = time.NewTicker(timeoutDuration)
		lines   []string
		readNum int
		err     error
	)
	defer timeout.Stop()

	for {
		select {
		case event, ok := <-t.watcher.Events:
			if !ok {
				t.opt.log.Warnf("receive events error, file %s", t.filepath)
				return
			}
			if err := t.tellEvent(event); err != nil {
				t.opt.log.Warn(err)
				return
			}

		case err, ok := <-t.watcher.Errors:
			t.opt.log.Warnf("receive error event from file %s, err: %s", t.filepath, err)
			if !ok {
				return
			}

		case <-t.stopCh:
			t.closeFile()
			t.opt.log.Infof("stop reading data from file %s", t.filename)
			return

		case <-timeout.C:
			if str := t.mult.FlushString(); str != "" {
				t.send(str)
			}

		default:
			// nil
		}

		b.buf, readNum, err = t.read()
		if err != nil {
			t.opt.log.Warnf("failed to read data from file %s, error: %s", t.filename, err)
			return
		}

		if readNum == 0 {
			t.wait()
			continue
		}

		// 如果接收到数据，则重置 ticker
		timeout.Reset(timeoutDuration)

		lines = b.split()

		if t.opt.DockerMode {
			t.dockerHandler(lines)
			continue
		}
		t.defaultHandler(lines)
	}
}

type dockerMessage struct {
	Log    string `json:"log"`
	Stream string `json:"stream"`
	// Time string `json:"time"`
}

func (t *Single) dockerHandler(lines []string) {
	var err error
	pending := []string{}

	tags := make(map[string]string)
	for k, v := range t.tags {
		tags[k] = v
	}
	for _, line := range lines {
		if line == "" {
			continue
		}

		var msg dockerMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			t.opt.log.Warnf("unmarshal err: %s, data: %s, ignored", err, line)
			msg = dockerMessage{
				Log:    line,
				Stream: "stdout",
			}
		}

		tags["stream"] = msg.Stream

		var text string
		text, err = t.decode(msg.Log)
		if err != nil {
			t.opt.log.Debugf("decode '%s' error: %s", t.opt.CharacterEncoding, err)
		}

		if len(text) > 0 && text[len(text)-1] == '\n' {
			text = text[:len(text)-1]
		}

		text = t.multiline(text)
		if text == "" {
			continue
		}

		logstr := removeAnsiEscapeCodes(text, t.opt.RemoveAnsiEscapeCodes)
		pending = append(pending, logstr)
	}
	if len(pending) == 0 {
		return
	}

	task := &worker.TaskTemplate{
		TaskName:              "logging/" + t.opt.Source,
		ScriptName:            t.opt.Pipeline,
		Source:                t.opt.Source,
		ContentDataType:       worker.ContentString,
		Content:               pending,
		IgnoreStatus:          t.opt.IgnoreStatus,
		DisableAddStatusField: t.opt.DisableAddStatusField,
		TS:                    time.Now(),
		MaxMessageLen:         maxFieldsLength,
		Tags:                  tags,
	}

	if err := worker.FeedPipelineTaskBlock(task); err != nil {
		t.opt.log.Warnf("pipline feed err = %v", err)
	}
}

func (t *Single) defaultHandler(lines []string) {
	var err error
	pending := []string{}
	for _, line := range lines {
		if line == "" {
			continue
		}

		var text string
		text, err = t.decode(line)
		if err != nil {
			t.opt.log.Debugf("decode '%s' error: %s", t.opt.CharacterEncoding, err)
		}

		text = t.multiline(text)
		if text == "" {
			continue
		}

		if t.opt.ForwardFunc != nil {
			t.sendToForwardCallback(text)
			continue
		}
		logstr := removeAnsiEscapeCodes(text, t.opt.RemoveAnsiEscapeCodes)
		pending = append(pending, logstr)
	}
	if len(pending) == 0 {
		return
	}
	t.sendToPipeline(pending)
}

func (t *Single) send(text string) {
	if t.opt.ForwardFunc != nil {
		t.sendToForwardCallback(text)
		return
	}

	t.sendToPipeline([]string{text})
}

func (t *Single) sendToForwardCallback(text string) {
	err := t.opt.ForwardFunc(t.filename, text)
	if err != nil {
		t.opt.log.Warnf("failed to forward text from file %s, error: %s", t.filename, err)
	}
}

func (t *Single) sendToPipeline(pending []string) {
	task := &worker.TaskTemplate{
		TaskName:              "logging/" + t.opt.Source,
		ScriptName:            t.opt.Pipeline,
		Source:                t.opt.Source,
		ContentDataType:       worker.ContentString,
		Content:               pending,
		IgnoreStatus:          t.opt.IgnoreStatus,
		DisableAddStatusField: t.opt.DisableAddStatusField,
		TS:                    time.Now(),
		MaxMessageLen:         maxFieldsLength,
		Tags:                  t.tags,
	}

	err := worker.FeedPipelineTaskBlock(task)
	if err != nil {
		t.opt.log.Warnf("pipline feed err = %v", err)
		return
	}
}

func (t *Single) currentOffset() int64 {
	if t.file == nil {
		return -1
	}
	offset, err := t.file.Seek(0, os.SEEK_CUR)
	if err != nil {
		return -1
	}
	return offset
}

func (t *Single) read() ([]byte, int, error) {
	n, err := t.file.Read(t.readBuff)
	if err != nil && err != io.EOF {
		// an unexpected error occurred, stop the tailor
		t.opt.log.Warnf("Unexpected error occurred while reading file: %s", err)
		return nil, 0, err
	}
	return t.readBuff[:n], n, nil
}

func (t *Single) wait() {
	time.Sleep(defaultSleepDuration)
}

func (t *Single) buildTags(globalTags map[string]string) map[string]string {
	tags := make(map[string]string)
	for k, v := range globalTags {
		tags[k] = v
	}
	if _, ok := tags["filepath"]; !ok {
		tags["filepath"] = t.filepath
	}
	if _, ok := tags["filename"]; !ok {
		tags["filename"] = t.filename
	}
	return tags
}

func (t *Single) decode(text string) (str string, err error) {
	if t.decoder == nil {
		return text, nil
	}
	return t.decoder.String(text)
}

func (t *Single) multiline(text string) string {
	if t.mult == nil {
		return text
	}
	return t.mult.ProcessLineString(text)
}

type buffer struct {
	buf           []byte
	previousBlock []byte
}

func (b *buffer) split() []string {
	// 以换行符 split
	lines := bytes.Split(b.buf, []byte{'\n'})
	if len(lines) == 0 {
		return nil
	}

	var res []string

	// block 不为空时，将其内容添加到 lines 首元素前端
	// block 置空
	if len(b.previousBlock) != 0 {
		lines[0] = append(b.previousBlock, lines[0]...)
		b.previousBlock = nil
	}

	// 当 lines 最后一个元素不为空时，说明这段内容并不包含换行符，将其暂存到 previousBlock
	if len(lines[len(lines)-1]) != 0 {
		// 将 lines 尾元素 append previousBlock，避免占用此 slice 造成内存泄漏
		b.previousBlock = append(b.previousBlock, lines[len(lines)-1]...)
		lines = lines[:len(lines)-1]
	}

	for _, line := range lines {
		res = append(res, string(line))
	}

	return res
}

func removeAnsiEscapeCodes(oldtext string, run bool) string {
	if !run {
		return oldtext
	}

	newtext, err := ansi.Strip([]byte(oldtext))
	if err != nil {
		l.Debugf("remove ansi code error: %w", err)
		return oldtext
	}

	return string(newtext)
}
