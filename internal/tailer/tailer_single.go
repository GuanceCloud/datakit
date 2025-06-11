// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/lang"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/encoding"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/ansi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/openfile"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/reader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/recorder"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/textparser"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
)

const (
	defaultSleepDuration = time.Second
	checkInterval        = time.Second * 3
)

type Single struct {
	opt *option

	file                     *os.File
	inode                    string
	recordKey                string
	filepath, insideFilepath string

	decoder *encoding.Decoder
	mult    *multiline.Multiline
	reader  reader.Reader

	offset    int64
	readLines int64

	partialContentBuff bytes.Buffer

	tags map[string]string
	log  *logger.Logger
}

func NewTailerSingle(filepath string, opts ...Option) (*Single, error) {
	t := &Single{
		opt:      getOption(opts...),
		filepath: filepath,
	}

	if t.opt.insideFilepathFunc != nil {
		t.insideFilepath = t.opt.insideFilepathFunc(filepath)
	}

	t.buildTags(t.opt.extraTags)
	t.log = logger.SLogger("logging/" + t.opt.source)

	if err := t.setup(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Single) setup() error {
	var err error

	if t.opt.characterEncoding != "" {
		t.decoder, err = encoding.NewDecoder(t.opt.characterEncoding)
		if err != nil {
			t.log.Warnf("newdecoder err: %s", err)
			return err
		}
	}

	t.mult, err = multiline.New(t.opt.multilinePatterns, multiline.WithMaxLength(int(t.opt.maxMultilineLength)))
	if err != nil {
		return err
	}

	t.file, err = openfile.OpenFile(t.filepath)
	if err != nil {
		t.log.Warnf("openfile err: %s", err)
		return err
	}

	t.reader = reader.NewReader(t.file)
	t.inode = openfile.Inode(t.filepath)
	t.recordKey = openfile.FileKey(t.filepath)

	if err := t.seekOffset(); err != nil {
		t.log.Warnf("set position err: %s", err)
		return err
	}
	return nil
}

func (t *Single) Run(ctx context.Context) {
	t.forwardMessage(ctx)
	t.Close()
}

func (t *Single) Close() {
	t.recordPosition()
	t.closeFile()
	t.log.Infof("closing file %s", t.filepath)
}

func (t *Single) seekOffset() error {
	if t.file == nil {
		return fmt.Errorf("unexpected file pointer")
	}

	var err error

	// use position
	data := recorder.Get(t.recordKey)
	if data != nil {
		offset := data.Offset
		var size int64

		if stat, err := t.file.Stat(); err != nil {
			t.log.Warnf("open file %s err %s, ignored", t.filepath, err)
		} else {
			size = stat.Size()
		}

		if offset > size {
			t.log.Infof("position %d larger than the file size %d, may be truncated", offset, size)
		} else {
			t.offset, err = t.file.Seek(offset, io.SeekStart)
			t.log.Infof("set position %d for file %s", offset, t.filepath)
			return err
		}
	}
	t.log.Infof("not found position for recorder key %s", t.recordKey)

	// use fromBeginning
	if t.opt.fromBeginning {
		t.offset, err = t.file.Seek(0, io.SeekStart)
		t.log.Infof("set start position for file %s", t.filepath)
		return err
	}

	// use fileFromBeginningThresholdSize
	if stat, _err := os.Stat(t.filepath); _err == nil {
		size := stat.Size()
		if size < t.opt.fileFromBeginningThresholdSize {
			t.offset, err = t.file.Seek(0, io.SeekStart)
			t.log.Infof("set start position for file %s, because file size %d < %d",
				t.filepath, size, t.opt.fileFromBeginningThresholdSize)
			return err
		}
	}

	// use tail
	t.offset, err = t.file.Seek(0, io.SeekEnd)
	t.log.Infof("set end position for file %s and offset=%d", t.filepath, t.offset)

	return err
}

func (t *Single) recordPosition() {
	if t.offset <= 0 {
		return
	}

	c := &recorder.MetaData{Source: t.opt.source, Offset: t.offset}

	if err := recorder.SetAndFlush(t.recordKey, c); err != nil {
		t.log.Debugf("recording cache %s err: %s", c, err)
	}
}

func (t *Single) closeFile() {
	if t.file == nil {
		return
	}
	if err := t.file.Close(); err != nil {
		t.log.Warnf("close file err: %s, ignored", err)
	}
	t.file = nil
}

func (t *Single) reopen() error {
	t.closeFile()

	var err error
	t.file, err = openfile.OpenFile(t.filepath)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %w", t.filepath, err)
	}

	ret, err := t.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	t.reader.SetReader(t.file)
	t.offset = ret
	t.inode = openfile.Inode(t.filepath)
	t.recordKey = openfile.FileKey(t.filepath)

	t.log.Infof("reopen file %s, offset %d", t.filepath, t.offset)
	rotateVec.WithLabelValues(t.opt.source, t.filepath).Inc()
	return nil
}

//nolint:cyclop
func (t *Single) forwardMessage(ctx context.Context) {
	checkTicker := time.NewTicker(checkInterval)
	defer checkTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.flushCache()
			t.log.Infof("exiting: file %s", t.filepath)
			return

		case <-checkTicker.C:
			did, _ := openfile.DidRotate(t.file, t.offset)
			exist := openfile.FileExists(t.filepath)

			if did || !exist {
				t.readToEOF()
			}

			if did {
				t.log.Infof("reopen the file %s", t.filepath)
				if err := t.reopen(); err != nil {
					t.log.Warnf("failed to reopen the file %s, err: %s", t.filepath, err)
					return
				}
			}

			if !openfile.FileIsActive(t.filepath, t.opt.ignoreDeadLog) {
				t.log.Infof("file %s has been inactive for larger than %s or has been removed, exit", t.filepath, t.opt.ignoreDeadLog)
				return
			}

		default: // nil
		}

		if err := t.readOnce(); err != nil {
			if !errors.Is(err, reader.ErrReadEmpty) {
				t.log.Warnf("failed to read data from file %s, error: %s", t.filepath, err)
			}
			t.flushCache()
			time.Sleep(defaultSleepDuration)
			continue
		}
	}
}

func (t *Single) readToEOF() {
	t.log.Infof("file %s has been rotated or removed, current offset %d, try to read EOF", t.filepath, t.offset)
	for {
		select {
		case <-datakit.Exit.Wait():
			return
		default:
			if err := t.readOnce(); err != nil {
				if !errors.Is(err, reader.ErrReadEmpty) {
					t.log.Warnf("read to EOF err: %s", err)
				}
				return
			}
		}
	}
}

func (t *Single) readOnce() error {
	block, readNum, err := t.reader.ReadLineBlock()
	if err != nil {
		return err
	}

	lines := reader.SplitLines(block)
	t.process(t.opt.mode, lines)

	t.offset += int64(readNum)
	t.log.Debugf("read %d bytes from file %s, offset %d", readNum, t.filepath, t.offset)
	return nil
}

func (t *Single) process(mode Mode, lines [][]byte) {
	var pending [][]byte

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var msg textparser.LogMessage
		var err error

		switch mode {
		case FileMode:
			err = textparser.ParseFileText(line, &msg)
		case DockerJSONLogMode:
			err = textparser.ParseDockerJSONLog(line, &msg)
		case CriLogdMode:
			err = textparser.ParseCRILog(line, &msg)
		default:
			err = textparser.ParseFileText(line, &msg)
		}
		if err != nil {
			t.log.Warnf("parse log failed %s", err)
			parseFailVec.WithLabelValues(t.opt.source, t.filepath, t.opt.mode.String()).Inc()
			continue
		}

		if msg.IsPartial {
			t.partialContentBuff.Write(msg.Log)
			continue
		}

		var originalText []byte
		if t.partialContentBuff.Len() != 0 {
			t.partialContentBuff.Write(msg.Log)
			originalText = t.partialContentBuff.Bytes()
			t.partialContentBuff.Reset()
		} else {
			originalText = msg.Log
		}

		text, err := t.decode(originalText)
		if err != nil {
			t.log.Debugf("decode '%s' error: %s", t.opt.characterEncoding, err)
		}

		text = removeAnsiEscapeCodes(text, t.opt.removeAnsiEscapeCodes)

		finalText := t.multiline(multiline.TrimRightSpace(text))
		if len(finalText) == 0 {
			continue
		}

		pending = append(pending, finalText)
	}

	t.feed(pending)
}

func (t *Single) feed(pending [][]byte) {
	// feed to remote
	if t.opt.forwardFunc != nil {
		t.feedToRemote(pending)
		return
	}
	t.feedToIO(pending)
}

func (t *Single) feedToRemote(pending [][]byte) {
	for _, text := range pending {
		t.readLines++
		fields := map[string]interface{}{
			"filepath":            t.filepath,
			"log_read_lines":      t.readLines,
			constants.FieldStatus: pipeline.DefaultStatus,
		}

		if t.opt.enableDebugFields {
			fields["log_read_offset"] = t.offset
			fields["log_file_inode"] = t.inode
		}

		err := t.opt.forwardFunc(t.filepath, string(text), fields)
		if err != nil {
			t.log.Warnf("failed to forward text from file %s, error: %s", t.filepath, err)
		}
	}
}

const (
	logTimeStep = time.Microsecond // 1us
)

func (t *Single) feedToIO(pending [][]byte) {
	var (
		pts     = []*point.Point{}
		opts    = append(point.DefaultLoggingOptions(), point.WithPrecheck(false))
		timeNow = ntp.Now().Add(-time.Duration(len(pending)) * logTimeStep)
	)

	for i, cnt := range pending {
		t.readLines++

		kvs := make(point.KVs, 0, len(t.tags)+4)
		kvs = kvs.Add(constants.FieldMessage, string(cnt), false, false)

		if t.shouldAddField("filepath") {
			kvs = kvs.Add("filepath", t.filepath, false, false)
		}
		if t.shouldAddField("log_read_lines") {
			kvs = kvs.Add("log_read_lines", t.readLines, false, false)
		}
		if t.shouldAddField(constants.FieldStatus) {
			kvs = kvs.AddTag(constants.FieldStatus, constants.DefaultStatus)
		}

		if t.shouldAddField("inside_filepath") && t.insideFilepath != "" {
			kvs = kvs.Add("inside_filepath", t.insideFilepath, false, false)
		}

		for key, value := range t.tags {
			kvs = kvs.AddTag(key, value)
		}

		if t.opt.enableDebugFields {
			kvs = kvs.Add("log_read_offset", t.offset, false, false)
			kvs = kvs.Add("log_file_inode", t.inode, false, false)
		}

		// only the message field is present, with no match in the whitelist
		// discard this data
		if len(kvs) == 1 {
			discardVec.WithLabelValues(t.opt.source, t.filepath).Inc()
			continue
		}

		pt := point.NewPointV2(t.opt.source, kvs, opts...)

		// 此处设置每条日志的时间差为 1us, 这样日志查看器中的日志显示顺序跟当前的采集顺序就保持一致了.
		// 此处如果这批日志的时间戳都一样, 在查看器中看到日志将可能随机展示(因为查看器默认按照 point 的
		// 时间戳来倒排显示)
		// 注意, 此处这个时间还是可以在后续的 pipeline 被改写.
		pt.SetTime(timeNow.Add(time.Duration(i) * logTimeStep))

		pts = append(pts, pt)
	}

	if len(pts) == 0 {
		return
	}

	if err := t.opt.feeder.FeedV2(
		point.Logging,
		pts,
		dkio.WithInputName("logging/"+t.opt.source),
		dkio.WithPipelineOption(&lang.LogOption{
			DisableAddStatusField: t.opt.disableAddStatusField,
			IgnoreStatus:          t.opt.ignoreStatus,
			ScriptMap:             map[string]string{t.opt.source: t.opt.pipeline},
		}),
	); err != nil {
		t.log.Errorf("feed %d pts failed: %s, logging block-mode off, ignored", len(pts), err)
	}
}

func (t *Single) flushCache() {
	if t.mult != nil && t.mult.BuffLength() > 0 {
		text := t.mult.Flush()
		logstr := removeAnsiEscapeCodes(text, t.opt.removeAnsiEscapeCodes)
		t.feed([][]byte{logstr})
	}
}

func (t *Single) buildTags(extraTags map[string]string) {
	t.tags = make(map[string]string)

	for k, v := range extraTags {
		if t.shouldAddField(k) {
			t.tags[k] = v
		}
	}
}

func (t *Single) shouldAddField(key string) bool {
	if len(t.opt.fieldWhiteList) == 0 {
		return true
	}
	_, exist := t.opt.fieldWhiteList[key]
	return exist
}

func (t *Single) decode(text []byte) ([]byte, error) {
	if t.decoder == nil {
		return text, nil
	}
	return t.decoder.Bytes(text)
}

func (t *Single) multiline(text []byte) []byte {
	if !t.opt.enableMultiline || t.mult == nil {
		return text
	}
	res, state := t.mult.ProcessLine(text)
	if state == multiline.FlushPartial {
		multilineStateVec.WithLabelValues(t.opt.source, state.String()).Inc()
	}
	return res
}

func removeAnsiEscapeCodes(text []byte, run bool) []byte {
	if !run {
		return text
	}
	return ansi.Strip(text)
}
