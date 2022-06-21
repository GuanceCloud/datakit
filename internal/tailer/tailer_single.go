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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/encoding"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/multiline"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
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

	tags            map[string]string
	stopCh          chan struct{}
	expectMultiLine bool // only for docker log, relation to log size (16K)
}

func NewTailerSingle(filename string, opt *Option) (*Single, error) {
	if opt == nil {
		return nil, fmt.Errorf("option cannot be null pointer")
	}

	t := &Single{
		stopCh: make(chan struct{}, 1),
		opt:    opt,
	}

	if opt.DockerMode && !fileExists(filename) {
		filename2 := filepath.Join("/rootfs", filename)
		if !fileExists(filename2) {
			return nil, fmt.Errorf("file %s does not exist", filename)
		}
		filename = filename2
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
		return nil, err
	}

	checkpointData, err := getLogCheckpoint(getFileKey(filename))
	if err != nil {
		if !opt.FromBeginning {
			if _, err := t.file.Seek(0, io.SeekEnd); err != nil {
				return nil, err
			}
		}
	} else if checkpointData != nil && checkpointData.Offset > 0 {
		if _, err := t.file.Seek(checkpointData.Offset, io.SeekStart); err != nil {
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
	t.forwardMessage()
	t.Close()
}

func (t *Single) Stop() {
	t.stopCh <- struct{}{}
}

func (t *Single) Close() {
	t.removeWatcher(t.filepath)
	t.closeWatcher()
	if offset := t.currentOffset(); offset > 0 {
		err := updateLogCheckpoint(getFileKey(t.filepath), &logCheckpointData{Offset: offset})
		if err != nil {
			t.opt.log.Warnf("update checkpoint %s, offset %d, err: %s", t.filepath, offset, err)
		}
	}
	t.closeFile()
	t.opt.log.Infof("closing %s", t.filepath)
}

func (t *Single) removeWatcher(fn string) {
	if t.watcher == nil {
		return
	}
	if err := t.watcher.Remove(fn); err != nil {
		t.opt.log.Warnf("remove watcher err: %s, ignored", err)
	}
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
	if t.file == nil {
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
				return
			}
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				t.opt.log.Debugf("receive rename event from file %s", t.filepath)
				if err = t.reopen(); err != nil {
					t.opt.log.Warnf("failed to reopen file %s, err: %s", t.filepath, err)
				}
			}

		case err, ok := <-t.watcher.Errors:
			if !ok {
				return
			}
			t.opt.log.Warnf("receive error event from file %s, err: %s", t.filepath, err)

		case <-t.stopCh:
			return

		case <-timeout.C:
			if str := t.mult.FlushString(); str != "" {
				t.send(str)
			}
			if !fileIsHealthy(t.filepath, t.opt.IgnoreDeadLog) {
				t.opt.log.Debugf("file %s does not exist, exit", t.filepath)
				return
			}
			if offset := t.currentOffset(); offset > 0 {
				err := updateLogCheckpoint(getFileKey(t.filepath), &logCheckpointData{Offset: offset})
				if err != nil {
					t.opt.log.Warnf("update checkpoint %s, offset %d, err: %s", t.filepath, offset, err)
				}
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

		if len(text) > 0 {
			// deal with docker log size exceed 16 K
			if text[len(text)-1] != '\n' {
				textLen := len(text)
				if t.expectMultiLine {
					text = t.multilineWithFlag(text, true)
				} else {
					text = t.multiline(multiline.TrimRightSpace(text))
				}
				t.expectMultiLine = textLen/1000 == 16 // almost to 16 K
			} else {
				if t.expectMultiLine {
					text = t.multilineWithFlag(text, true)
					t.expectMultiLine = false
				} else {
					text = t.multiline(multiline.TrimRightSpace(text))
				}
			}
		}

		if text == "" {
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

		text = t.multiline(multiline.TrimRightSpace(text))
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
	res := []*iod.Point{}
	// -1ns
	timeNow := time.Now().Add(-time.Duration(len(pending)))
	for i, cnt := range pending {
		pt, err := iod.MakePoint(t.opt.Source, t.tags,
			map[string]interface{}{
				pipeline.FieldMessage: cnt,
				pipeline.FieldStatus:  pipeline.DefaultStatus,
			},
			timeNow.Add(time.Duration(i)),
		)
		if err != nil {
			l.Error(err)
			continue
		}
		res = append(res, pt)
	}
	if len(res) > 0 {
		if err := iod.Feed("logging/"+t.opt.Source, datakit.Logging, res, &iod.Option{
			PlScript: map[string]string{t.opt.Source: t.opt.Pipeline},
			PlOption: &script.Option{
				MaxFieldValLen:        maxFieldsLength,
				DisableAddStatusField: t.opt.DisableAddStatusField,
				IgnoreStatus:          t.opt.IgnoreStatus,
			},
		}); err != nil {
			l.Error(err)
		}
	}
}

func (t *Single) currentOffset() int64 {
	if t.file == nil {
		return -1
	}
	offset, err := t.file.Seek(0, io.SeekCurrent)
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

func (t *Single) multilineWithFlag(text string, flag bool) string {
	if t.mult == nil {
		return text
	}
	return t.mult.ProcessLineStringWithFlag(text, flag)
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func fileIsHealthy(fn string, timeout time.Duration) bool {
	info, err := os.Stat(fn)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	if timeout > 0 && time.Since(info.ModTime()) > timeout {
		return false
	}
	return true
}
