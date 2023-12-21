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

	"github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/encoding"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/ansi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/diskcache"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/register"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"google.golang.org/protobuf/proto"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

const (
	defaultSleepDuration = time.Second
	readBuffSize         = 1024 * 4   // 4 KiB
	maxReadSize          = 1024 * 128 // 128 KiB

	checkInterval = time.Second * 1
)

type Single struct {
	opt                *Option
	file               *os.File
	filepath, filename string

	decoder *encoding.Decoder
	mult    *multiline.Multiline

	flushScore int
	readBuff   []byte
	readLines  int64

	offset   int64 // 必然只在同一个 goroutine 操作，不必使用 atomic
	readTime time.Time

	partialContentBuff bytes.Buffer

	enableDiskCache bool

	tags map[string]string
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

	_ = logtail.InitDefault()

	var err error
	if opt.CharacterEncoding != "" {
		t.decoder, err = encoding.NewDecoder(opt.CharacterEncoding)
		if err != nil {
			return nil, err
		}
	}
	t.mult, err = multiline.New(
		opt.MultilinePatterns,
		&multiline.Option{MaxLifeDuration: opt.MaxMultilineLifeDuration},
	)
	if err != nil {
		return nil, err
	}

	t.file, err = openFile(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}
	t.filepath = t.file.Name()
	t.filename = filepath.Base(t.filepath)

	if err := t.seekOffset(); err != nil {
		return nil, err
	}

	t.readBuff = make([]byte, readBuffSize)
	t.tags = t.buildTags(opt.GlobalTags)

	openfileVec.WithLabelValues(t.opt.Mode.String()).Inc()
	return t, nil
}

func (t *Single) Run() {
	t.forwardMessage()
	t.Close()
}

func (t *Single) Close() {
	t.recordingLastCache()
	t.closeFile()

	openfileVec.WithLabelValues(t.opt.Mode.String()).Dec()
	t.opt.log.Infof("closing: file %s", t.filepath)
}

func (t *Single) seekOffset() error {
	if t.file == nil {
		return fmt.Errorf("unexpected file pointer")
	}

	var ret int64
	var err error

	func() {
		if pos := t.getPosition(); pos != -1 {
			ret, err = t.file.Seek(pos, io.SeekStart)
			t.opt.log.Infof("set position %d for filename %s", pos, t.filepath)
			return
		}

		if t.opt.FromBeginning {
			ret, err = t.file.Seek(0, io.SeekStart)
			t.opt.log.Infof("set start position for filename %s", t.filepath)
			return
		}

		if stat, _err := os.Stat(t.filepath); _err == nil {
			if stat.Size() < t.opt.FileFromBeginningThresholdSize {
				ret, err = t.file.Seek(0, io.SeekStart)
				t.opt.log.Infof("set start position for filename %s, because file size < %sKiB",
					t.opt.FileFromBeginningThresholdSize/1024)
			}
			return
		}

		ret, err = t.file.Seek(0, io.SeekEnd)
		t.opt.log.Infof("set end position for filename %s", t.filepath)
	}()

	t.offset = ret
	return err
}

func (t *Single) getPosition() (pos int64) {
	// default -1
	pos = -1

	data := register.Get(getFileKey(t.filepath))
	if data == nil {
		return
	}

	stat, err := t.file.Stat()
	if err != nil {
		t.opt.log.Warnf("open file %s err %s, ignored", t.filepath, err)
		return
	}

	if size := stat.Size(); data.Offset > size {
		t.opt.log.Infof("position %d larger than the file size %d, may be truncated", data.Offset, size)
		return
	}

	return data.Offset
}

func (t *Single) recordingCache() {
	if t.offset <= 0 {
		return
	}

	c := &register.MetaData{Source: t.opt.Source, Offset: t.offset}

	if err := register.Set(getFileKey(t.filepath), c); err != nil {
		t.opt.log.Debugf("recording cache %s err: %s", c, err)
		return
	}

	t.opt.log.Debugf("recording cache %s success", c)
}

func (t *Single) recordingLastCache() {
	if t.offset <= 0 {
		return
	}

	c := &register.MetaData{Source: t.opt.Source, Offset: t.offset}

	if err := register.SetAndFlush(getFileKey(t.filepath), c); err != nil {
		t.opt.log.Debugf("recording last cache %s err: %s", c, err)
		return
	}

	t.opt.log.Debugf("recording last cache %s success", c)
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
	t.file, err = openFile(t.filepath)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %w", t.filepath, err)
	}

	ret, err := t.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	t.offset = ret
	t.opt.log.Infof("reopen file %s, offset %d", t.filepath, t.offset)

	rotateVec.WithLabelValues(t.opt.Source, t.filepath).Inc()
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

	for {
		if t.shoudFlush() {
			if t.mult != nil && t.mult.BuffLength() > 0 {
				t.feed([]string{t.mult.FlushString()})
				forceFlushVec.WithLabelValues(t.opt.Source, t.filepath).Inc()
			}
			t.resetFlushScore()
		}

		select {
		case <-datakit.Exit.Wait():
			t.opt.log.Infof("exiting: file %s", t.filepath)
			return
		case <-t.opt.Done:
			t.opt.log.Infof("exiting: file %s", t.filepath)
			return

		case <-checkTicker.C:
			did, _ := DidRotate(t.file, t.offset)
			exist := FileExists(t.filepath)

			if did || !exist {
				t.opt.log.Infof("file %s has been rotated or removed, current offset %d, try to read EOF", t.filepath, t.offset)
				for {
					b.buf, readNum, err = t.read()
					if err != nil {
						t.opt.log.Warnf("failed to read data from file %s, error: %s", t.filename, err)
						break
					}

					t.opt.log.Debugf("read %d bytes from file %s, offset %d", readNum, t.filepath, t.offset)

					if readNum == 0 {
						t.opt.log.Infof("read EOF from rotate file %s, offset %d", t.filepath, t.offset)
						break
					}
					t.readTime = time.Now()

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
					// 数据处理完成，再记录 offset
					t.offset += int64(readNum)
				}
			}

			if did { // 只有文件 rotated 才会 reopen
				t.opt.log.Infof("reopen the file %s", t.filepath)
				if err = t.reopen(); err != nil {
					t.opt.log.Warnf("failed to reopen the file %s, err: %s", t.filepath, err)
					return
				}
			}

			if !FileIsActive(t.filepath, t.opt.IgnoreDeadLog) {
				t.opt.log.Infof("file %s has been inactive for larger than %s or has been removed, exit", t.filepath, t.opt.IgnoreDeadLog)
				return
			}

		default: // nil
		}

		b.buf, readNum, err = t.read()
		if err != nil {
			t.opt.log.Warnf("failed to read data from file %s, error: %s", t.filename, err)
			t.wait()
			continue
		}

		t.opt.log.Debugf("read %d bytes from file %s, offset %d", readNum, t.filepath, t.offset)

		if readNum == 0 {
			t.wait()
			continue
		}
		t.readTime = time.Now()

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

		// 数据处理完成，再记录 offset
		t.offset += int64(readNum)
		t.resetFlushScore()
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
	pending := []string{}
	for _, line := range lines {
		if line == "" {
			continue
		}

		var msg dockerMessage
		err := json.Unmarshal([]byte(line), &msg)
		if err != nil {
			parseFailVec.WithLabelValues(t.opt.Source, t.filepath, t.opt.Mode.String()).Inc()
			t.opt.log.Warnf("json-data parsed err: %s, data: %s, ignored", err, trim(line))
			continue
		}

		if isJSONLogPartialContent(msg.Log) {
			t.partialContentBuff.WriteString(msg.Log)
			continue
		}

		var originalText string

		if t.partialContentBuff.Len() != 0 {
			t.partialContentBuff.WriteString(msg.Log)
			originalText = t.partialContentBuff.String()
			t.partialContentBuff.Reset()
		} else {
			originalText = msg.Log
		}

		var text string
		text, err = t.decode(originalText)
		if err != nil {
			t.opt.log.Debugf("decode '%s' error: %s", t.opt.CharacterEncoding, err)
		}

		text = t.multiline(multiline.TrimRightSpace(text))
		if text == "" {
			continue
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

		err := parseCRILog([]byte(line), &criMsg)
		if err != nil {
			parseFailVec.WithLabelValues(t.opt.Source, t.filepath, t.opt.Mode.String()).Inc()
			t.opt.log.Warnf("cri-log parsed err: %s, data: %s, ignored", err, trim(line))
			continue
		}

		if criMsg.isPartial {
			t.partialContentBuff.WriteString(criMsg.log)
			continue
		}

		var contents []string

		if t.partialContentBuff.Len() != 0 {
			contents = append(contents, t.partialContentBuff.String())
			t.partialContentBuff.Reset()
		}
		contents = append(contents, criMsg.log)

		for _, content := range contents {
			var text string

			text, err = t.decode(content)
			if err != nil {
				t.opt.log.Debugf("decode '%s' error: %s", t.opt.CharacterEncoding, err)
			}

			text = t.multiline(multiline.TrimRightSpace(text))
			if text == "" {
				continue
			}

			logstr := removeAnsiEscapeCodes(text, t.opt.RemoveAnsiEscapeCodes)
			pending = append(pending, logstr)
		}
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

		logstr := removeAnsiEscapeCodes(text, t.opt.RemoveAnsiEscapeCodes)
		pending = append(pending, logstr)
	}
	if len(pending) == 0 {
		return
	}
	t.feed(pending)
}

func (t *Single) feed(pending []string) {
	// feed to remote
	if t.opt.ForwardFunc != nil {
		t.feedToRemote(pending)
		return
	}
	defer t.recordingCache()

	if t.enableDiskCache {
		err := t.feedToCache(pending)
		if err == nil {
			return
		}
		t.opt.log.Warnf("failed of save cache, err: %s, retry feed to io", err)
	}

	t.feedToIO(pending)
}

func (t *Single) feedToRemote(pending []string) {
	for _, text := range pending {
		err := t.opt.ForwardFunc(t.filename, text)
		if err != nil {
			t.opt.log.Warnf("failed to forward text from file %s, error: %s", t.filename, err)
		}
	}
}

func (t *Single) feedToCache(pending []string) error {
	res := []*point.Point{}
	// -1ns
	timeNow := time.Now().Add(-time.Duration(len(pending)))

	for i, cnt := range pending {
		t.readLines++

		fields := map[string]interface{}{
			"log_read_lines":      t.readLines,
			"log_read_offset":     t.offset,
			"log_read_time":       t.readTime.UnixNano(),
			"message_length":      len(cnt),
			pipeline.FieldMessage: cnt,
			pipeline.FieldStatus:  pipeline.DefaultStatus,
		}

		pt := point.NewPointV2(t.opt.Source,
			append(point.NewTags(t.tags), point.NewKVs(fields)...),
			point.WithTime(timeNow.Add(time.Duration(i))),
		)
		res = append(res, pt)
	}

	if len(res) == 0 {
		return nil
	}

	encoder := point.GetEncoder(point.WithEncBatchSize(0), point.WithEncEncoding(point.Protobuf))

	ptsDatas, err := encoder.Encode(res)
	if err != nil {
		return err
	}
	if len(ptsDatas) != 1 {
		err := fmt.Errorf("invalid pbpoints")
		return err
	}

	pbdata := &diskcache.PBData{
		Points: ptsDatas[0],
		Config: &diskcache.PBConfig{
			Source:                t.opt.Source,
			Pipeline:              t.opt.Pipeline,
			Blocking:              t.opt.BlockingMode,
			DisableAddStatusField: t.opt.DisableAddStatusField,
			IgnoreStatus:          t.opt.IgnoreStatus,
		},
	}

	b, err := proto.Marshal(pbdata)
	if err != nil {
		return err
	}

	return diskcache.Put(b)
}

func (t *Single) feedToIO(pending []string) {
	pts := []*point.Point{}
	// -1ns
	timeNow := time.Now().Add(-time.Duration(len(pending)))
	for i, cnt := range pending {
		t.readLines++

		fields := map[string]interface{}{
			"log_read_lines":      t.readLines,
			"log_read_offset":     t.offset,
			"log_read_time":       t.readTime.UnixNano(),
			"message_length":      len(cnt),
			pipeline.FieldMessage: cnt,
			pipeline.FieldStatus:  pipeline.DefaultStatus,
		}
		opts := append(point.DefaultLoggingOptions(), point.WithTime(timeNow.Add(time.Duration(i))))

		pt := point.NewPointV2(
			t.opt.Source,
			append(point.NewTags(t.tags), point.NewKVs(fields)...),
			opts...,
		)
		pts = append(pts, pt)
	}

	if len(pts) == 0 {
		return
	}

	if err := t.opt.Feeder.Feed(
		"logging/"+t.opt.Source,
		point.Logging,
		pts,
		&dkio.Option{
			PlOption: &manager.Option{
				DisableAddStatusField: t.opt.DisableAddStatusField,
				IgnoreStatus:          t.opt.IgnoreStatus,
				ScriptMap:             map[string]string{t.opt.Source: t.opt.Pipeline},
			},
			Blocking: t.opt.BlockingMode,
		},
	); err != nil {
		t.opt.log.Errorf("feed %d pts failed: %s, logging block-mode off, ignored", len(pts), err)
	}
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
	t.flushScore++
}

func (t *Single) shoudFlush() bool {
	if t.opt.MaxForceFlushLimit == -1 {
		return false
	}
	return t.flushScore >= t.opt.MaxForceFlushLimit
}

func (t *Single) resetFlushScore() {
	t.flushScore = 0
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

	res, state := t.mult.ProcessLineString(text)
	multilineVec.WithLabelValues(t.opt.Source, t.filepath, state.String()).Inc()
	return res
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

	if len(b.previousBlock) > maxReadSize {
		res = append(res, string(b.previousBlock))
		b.previousBlock = nil
	}

	for _, line := range lines {
		res = append(res, string(line))
	}

	return res
}

func removeAnsiEscapeCodes(text string, run bool) string {
	if !run {
		return text
	}
	return ansi.Strip(text)
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
//
//	2016-10-06T00:17:09.669794202Z stdout P log content 1
//	2016-10-06T00:17:09.669794203Z stderr F log content 2
//
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

func isJSONLogPartialContent(content string) bool {
	if len(content) < 1 {
		return false
	}
	if content[len(content)-1] != '\n' {
		return true
	}
	return false
}

func trim(s string) string {
	if len(s) > 64 {
		return s[:64]
	}
	return s
}
