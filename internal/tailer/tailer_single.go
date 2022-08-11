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

	"github.com/pborman/ansi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/encoding"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/multiline"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

const (
	defaultSleepDuration = time.Second
	readBuffSize         = 1024 * 4

	checkInterval = time.Second * 1
)

type Single struct {
	opt                *Option
	file               *os.File
	filepath, filename string

	decoder *encoding.Decoder
	mult    *multiline.Multiline

	readBuff  []byte
	readLines int64

	offset int64 // 必然只在同一个 goroutine 操作，不必使用 atomic

	tags            map[string]string
	expectMultiLine bool // only for docker log, relation to log size (16K)
}

func NewTailerSingle(filename string, opt *Option) (*Single, error) {
	if opt == nil {
		return nil, fmt.Errorf("option cannot be null pointer")
	}

	t := &Single{opt: opt}

	if opt.Mode != FileMode && !FileExists(filename) {
		filename2 := filepath.Join("/rootfs", filename)
		if !FileExists(filename2) {
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

	// check if from begine
	if !opt.FromBeginning {
		ret, err := t.file.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}
		t.offset = ret
	}

	checkpointData, err := getLogCheckpoint(getFileKey(filename))
	if err == nil {
		stat, err := t.file.Stat()
		if err != nil {
			return nil, err
		}

		if checkpointData.Offset <= stat.Size() {
			ret, err := t.file.Seek(checkpointData.Offset, io.SeekStart)
			if err != nil {
				return nil, err
			}
			t.offset = ret
		}
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

func (t *Single) Close() {
	if t.offset > 0 {
		err := updateLogCheckpoint(getFileKey(t.filepath), &logCheckpointData{Offset: t.offset})
		if err != nil {
			t.opt.log.Warnf("update checkpoint %s, offset %d, err: %s", t.filepath, t.offset, err)
		}
	}
	t.closeFile()
	t.opt.log.Infof("closing %s", t.filepath)
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

	var err error
	t.file, err = os.Open(t.filepath)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %w", t.filepath, err)
	}

	ret, err := t.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	t.offset = ret
	t.opt.log.Infof("reopen file %s, offset %d", t.filepath, t.offset)
	return nil
}

//nolint:cyclop
func (t *Single) forwardMessage() {
	var (
		b       = &buffer{}
		lines   []string
		readNum int
		err     error

		checkTicker = time.NewTicker(checkInterval)
	)
	defer checkTicker.Stop()

	handle := func(read func() ([]byte, int, error)) {
		b.buf, readNum, err = read()
		if err != nil {
			t.opt.log.Warnf("failed to read data from file %s, error: %s", t.filename, err)
			return
		}

		t.opt.log.Debugf("read %d bytes from file %s, offset %d", readNum, t.filepath, t.offset)

		if readNum == 0 {
			t.wait()
			return
		}

		lines = b.split()

		switch t.opt.Mode {
		case FileMode:
			t.defaultHandler(lines)
		case DockerMode:
			t.dockerHandler(lines)
		case ContainerdMode:
			t.containerdHandler(lines)
		default:
			t.defaultHandler(lines)
		}
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-checkTicker.C:
			if !FileIsActive(t.filepath, t.opt.IgnoreDeadLog) {
				t.opt.log.Infof("file %s is not active, larger than %s, exit", t.filepath, t.opt.IgnoreDeadLog)
				return
			}
			if t.offset > 0 {
				err := updateLogCheckpoint(getFileKey(t.filepath), &logCheckpointData{Offset: t.offset})
				if err != nil {
					t.opt.log.Warnf("update checkpoint %s, offset %d, err: %s", t.filepath, t.offset, err)
				}
			}

			did, err := DidRotate(t.file, t.offset)
			if err != nil {
				t.opt.log.Warnf("didrotate err: %s", err)
			}
			if did {
				t.opt.log.Infof("file %s has rotated, try to reopen file", t.filepath)

				handle(t.readAll)

				if err = t.reopen(); err != nil {
					t.opt.log.Warnf("failed to reopen file %s, err: %s", t.filepath, err)
					return
				}
			}

		default: // nil
		}

		handle(t.read)
	}
}

type dockerMessage struct {
	Log    string `json:"log"`
	Stream string `json:"stream"`
	// Time string `json:"time"`
}

func (t *Single) dockerHandler(lines []string) {
	logs := t.generateJSONLogs(lines)
	t.feed(logs)
}

func (t *Single) generateJSONLogs(lines []string) []string {
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
					text = t.multiline(text)
				}
				t.expectMultiLine = textLen/1000 == 16 // almost to 16 K
			} else {
				if t.expectMultiLine {
					text = t.multilineWithFlag(text, true)
					t.expectMultiLine = false
				} else {
					text = t.multiline(text)
				}
			}
		}

		if text == "" {
			continue
		}

		// text 意外的不匹配多行规则
		if !t.mult.MatchString(text) {
			t.opt.log.Warnf("unexpected multiline text: %s, next: %s, file %s, multiline rule '%s'",
				text, t.mult.BuffString(), t.filename, t.opt.MultilineMatch)
		}

		logstr := removeAnsiEscapeCodes(text, t.opt.RemoveAnsiEscapeCodes)
		pending = append(pending, logstr)
	}

	return pending
}

func (t *Single) containerdHandler(lines []string) {
	logs := t.generateCRILogs(lines)
	t.feed(logs)
}

func (t *Single) generateCRILogs(lines []string) []string {
	pending := []string{}
	for _, line := range lines {
		if line == "" {
			continue
		}

		var criMsg logMessage
		var text string

		if err := parseCRILog([]byte(line), &criMsg); err != nil {
			l.Warnf("parse cri-o log err: %s, data: %s", err, line)
			continue
		}

		if t.expectMultiLine {
			text = t.multilineWithFlag(criMsg.log, true)
		} else {
			text = t.multiline(criMsg.log)
		}

		t.expectMultiLine = criMsg.isPartial

		if text == "" {
			continue
		}

		// text 意外的不匹配多行规则
		if !t.mult.MatchString(text) {
			t.opt.log.Warnf("unexpected multiline text: %s, next: %s, file %s, multiline rule '%s'",
				text, t.mult.BuffString(), t.filename, t.opt.MultilineMatch)
		}

		logstr := removeAnsiEscapeCodes(text, t.opt.RemoveAnsiEscapeCodes)
		pending = append(pending, logstr)
	}

	return pending
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

		// TODO
		// text 意外的不匹配多行规则
		if !t.mult.MatchString(text) {
			t.opt.log.Warnf("unexpected multiline text: %s, next: %s, file %s, multiline rule '%s'",
				text, t.mult.BuffString(), t.filename, t.opt.MultilineMatch)
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
	t.feed(pending)
}

func (t *Single) sendToForwardCallback(text string) {
	err := t.opt.ForwardFunc(t.filename, text)
	if err != nil {
		t.opt.log.Warnf("failed to forward text from file %s, error: %s", t.filename, err)
	}
}

func (t *Single) feed(pending []string) {
	res := []*point.Point{}
	// -1ns
	timeNow := time.Now().Add(-time.Duration(len(pending)))
	for i, cnt := range pending {
		t.readLines++
		pt, err := point.NewPoint(t.opt.Source, t.tags,
			map[string]interface{}{
				"log_read_lines":      t.readLines,
				"log_read_offset":     t.offset,
				pipeline.FieldMessage: cnt,
				pipeline.FieldStatus:  pipeline.DefaultStatus,
			},
			&point.PointOption{Time: timeNow.Add(time.Duration(i)), Category: datakit.Logging})
		if err != nil {
			t.opt.log.Error(err)
			continue
		}
		res = append(res, pt)
	}

	if len(res) == 0 {
		return
	}

	retry := 0
__retry:
	if err := iod.Feed("logging/"+t.opt.Source, datakit.Logging, res, &iod.Option{
		PlScript: map[string]string{t.opt.Source: t.opt.Pipeline},
		PlOption: &script.Option{
			MaxFieldValLen:        maxFieldsLength,
			DisableAddStatusField: t.opt.DisableAddStatusField,
			IgnoreStatus:          t.opt.IgnoreStatus,
		},
	}); err != nil {
		if t.opt.BlockingMode {
			t.opt.log.Warnf("feed %d pts failed: %s, logging block-mode on, retry/%d", len(res), err, retry)
			time.Sleep(time.Millisecond * 300)
			retry++
			goto __retry
		} else {
			t.opt.log.Errorf("feed %d pts failed: %s, logging block-mode off, ignored", len(res), err)
		}
	}
}

func (t *Single) read() ([]byte, int, error) {
	n, err := t.file.Read(t.readBuff)
	if err != nil && err != io.EOF {
		// an unexpected error occurred, stop the tailor
		t.opt.log.Warnf("Unexpected error occurred while reading file: %s", err)
		return nil, 0, err
	}
	t.offset += int64(n)

	return t.readBuff[:n], n, nil
}

func (t *Single) readAll() ([]byte, int, error) {
	var res []byte
	var num int

	temp := make([]byte, 1024)

	for {
		n, err := t.file.Read(temp)
		if err != nil && err != io.EOF {
			// an unexpected error occurred, stop the tailor
			t.opt.log.Warnf("Unexpected error occurred while reading file: %s", err)
			return nil, 0, err
		}
		if n == 0 {
			break
		}
		res = append(res, temp[:n]...)
		num += n
	}

	t.offset += int64(num)
	return res, num, nil
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
		// l.Debugf("remove ansi code error: %w", err)
		return oldtext
	}

	return string(newtext)
}

var (
	// timeFormatIn is the format for parsing timestamps from other logs.
	timeFormatIn = "2006-01-02T15:04:05.999999999Z07:00"

	// delimiter is the delimiter for timestamp and stream type in log line.
	delimiter = []byte{' '}
	// tagDelimiter is the delimiter for log tags.
	tagDelimiter = []byte(runtimeapi.LogTagDelimiter)
)

// logMessage is the CRI internal log type.
type logMessage struct {
	timestamp time.Time
	stream    runtimeapi.LogStreamType
	log       string
	isPartial bool
}

// parseCRILog parses logs in CRI log format. CRI Log format example:
//   2016-10-06T00:17:09.669794202Z stdout P log content 1
//   2016-10-06T00:17:09.669794203Z stderr F log content 2
// refer to https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/kuberuntime/logs/logs.go#L128
func parseCRILog(log []byte, msg *logMessage) error {
	var err error
	// Parse timestamp
	idx := bytes.Index(log, delimiter)
	if idx < 0 {
		return fmt.Errorf("timestamp is not found")
	}
	msg.timestamp, err = time.Parse(timeFormatIn, string(log[:idx]))
	if err != nil {
		return fmt.Errorf("unexpected timestamp format %q: %w", timeFormatIn, err)
	}

	// Parse stream type
	log = log[idx+1:]
	idx = bytes.Index(log, delimiter)
	if idx < 0 {
		return fmt.Errorf("stream type is not found")
	}
	msg.stream = runtimeapi.LogStreamType(log[:idx])
	if msg.stream != runtimeapi.Stdout && msg.stream != runtimeapi.Stderr {
		return fmt.Errorf("unexpected stream type %q", msg.stream)
	}

	// Parse log tag
	log = log[idx+1:]
	idx = bytes.Index(log, delimiter)
	if idx < 0 {
		return fmt.Errorf("log tag is not found")
	}
	// Keep this forward compatible.
	tags := bytes.Split(log[:idx], tagDelimiter)
	partial := runtimeapi.LogTag(tags[0]) == runtimeapi.LogTagPartial
	// Trim the tailing new line if this is a partial line.
	if partial && len(log) > 0 && log[len(log)-1] == '\n' {
		log = log[:len(log)-1]
	}
	msg.isPartial = partial

	// Get log content
	msg.log = string(log[idx+1:])

	return nil
}
