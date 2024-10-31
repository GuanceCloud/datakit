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
	"github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/encoding"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/ansi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/bytechannel"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/openfile"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/reader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/recorder"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/textparser"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
)

const (
	defaultSleepDuration = time.Second
	checkInterval        = time.Second * 3
)

type Single struct {
	opt *option

	file      *os.File
	inode     string
	recordKey string
	filepath  string

	decoder *encoding.Decoder
	mult    *multiline.Multiline
	reader  reader.Reader

	offset    int64
	readLines int64
	bc        bytechannel.ByteChannel

	partialContentBuff bytes.Buffer

	tags map[string]string
	log  *logger.Logger
}

func NewTailerSingle(filepath string, opts ...Option) (*Single, error) {
	_ = logtail.InitDefault()

	c := defaultOption()
	for _, opt := range opts {
		opt(c)
	}

	t := &Single{
		opt:      c,
		filepath: filepath,
		bc:       bytechannel.NewByteChannel(),
	}
	t.buildTags(t.opt.extraTags)
	t.log = logger.SLogger("logging/" + t.opt.source)

	if err := t.setup(); err != nil {
		return nil, err
	}

	openfileVec.WithLabelValues(t.opt.mode.String()).Inc()
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

	t.mult, err = multiline.New(t.opt.multilinePatterns,
		multiline.WithMaxLength(int(t.opt.maxMultilineLength)),
		multiline.WithMaxLifeDuration(t.opt.maxMultilineLifeDuration))
	if err != nil {
		return err
	}

	t.file, err = openfile.OpenFile(t.filepath)
	if err != nil {
		t.log.Warnf("openfile err: %s", err)
		return err
	}

	t.reader = reader.NewReader(t.file)
	t.inode = openfile.FileInode(t.filepath)
	t.recordKey = openfile.FileKey(t.filepath)

	if err := t.seekOffset(); err != nil {
		t.log.Warnf("set position err: %s", err)
		return err
	}
	return nil
}

var processorGo = datakit.G("tailer_processor")

func (t *Single) Run() {
	ctx, cancel := context.WithCancel(context.Background())

	processorGo.Go(func(_ context.Context) error {
		for {
			select {
			case <-datakit.Exit.Wait():
				cancel()
				t.log.Infof("%s processor exit", t.opt.source)
				return nil
			case <-t.opt.done:
				cancel()
				t.log.Infof("%s processor exit", t.opt.source)
				return nil
			case <-ctx.Done():
				t.log.Infof("%s processor exit", t.opt.source)
				return nil
			default:
				// nil
			}

			block := t.bc.Receive()
			if len(block) == 0 {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			lines := reader.SplitLines(block)
			t.process(t.opt.mode, lines)

			pendingBlockLength.WithLabelValues(t.opt.source, t.filepath).Set(float64(t.bc.CurrentChannelSize()))
			pendingByteSize.WithLabelValues(t.opt.source, t.filepath).Set(float64(t.bc.CurrentByteSize()))
		}
	})

	t.forwardMessage(ctx)
	t.Close()
	cancel()
}

func (t *Single) Close() {
	t.recordPosition()
	t.closeFile()

	openfileVec.WithLabelValues(t.opt.mode.String()).Dec()
	t.log.Infof("closing: file %s", t.filepath)
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
	t.log.Infof("set end position for file %s", t.filepath)

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
	t.inode = openfile.FileInode(t.filepath)
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
			t.log.Infof("exiting: file %s", t.filepath)
			return

		case <-checkTicker.C:
			did, _ := openfile.DidRotate(t.file, t.offset)
			exist := openfile.FileExists(t.filepath)

			if did || !exist {
				t.readToEOF(ctx)
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

		if err := t.readOnce(ctx); err != nil {
			if !errors.Is(err, reader.ErrReadEmpty) {
				t.log.Warnf("failed to read data from file %s, error: %s", t.filepath, err)
			}
			t.wait()
			continue
		}
	}
}

func (t *Single) readToEOF(ctx context.Context) {
	t.log.Infof("file %s has been rotated or removed, current offset %d, try to read EOF", t.filepath, t.offset)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := t.readOnce(ctx); err != nil {
				if !errors.Is(err, reader.ErrReadEmpty) {
					t.log.Warnf("read to EOF err: %s", err)
				}
				return
			}
		}
	}
}

func (t *Single) readOnce(ctx context.Context) error {
	block, readNum, err := t.reader.ReadLineBlock()
	if err != nil {
		return err
	}

	t.bc.SendSync(ctx, block)
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
			"filepath":           t.filepath,
			"log_read_lines":     t.readLines,
			pipeline.FieldStatus: pipeline.DefaultStatus,
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

func (t *Single) feedToIO(pending [][]byte) {
	pts := []*point.Point{}
	// -1us
	timeNow := time.Now().Add(-time.Duration(len(pending)) * time.Microsecond)
	for i, cnt := range pending {
		t.readLines++

		fields := map[string]interface{}{
			"filepath":            t.filepath,
			"log_read_lines":      t.readLines,
			pipeline.FieldMessage: string(cnt),
			pipeline.FieldStatus:  pipeline.DefaultStatus,
		}
		if t.opt.enableDebugFields {
			fields["log_read_offset"] = t.offset
			fields["log_file_inode"] = t.inode
		}

		opts := append(point.DefaultLoggingOptions(),
			point.WithTime(timeNow.Add(time.Duration(i)*time.Microsecond)),
			point.WithPrecheck(false),
		)

		pt := point.NewPointV2(
			t.opt.source,
			append(point.NewTags(t.tags), point.NewKVs(fields)...),
			opts...,
		)
		pts = append(pts, pt)
	}

	if len(pts) == 0 {
		return
	}

	if err := t.opt.feeder.FeedV2(point.Logging, pts,
		dkio.WithInputName("logging/"+t.opt.source),
		dkio.WithPipelineOption(&manager.Option{
			DisableAddStatusField: t.opt.disableAddStatusField,
			IgnoreStatus:          t.opt.ignoreStatus,
			ScriptMap:             map[string]string{t.opt.source: t.opt.pipeline},
		}),
	); err != nil {
		t.log.Errorf("feed %d pts failed: %s, logging block-mode off, ignored", len(pts), err)
	}
}

func (t *Single) wait() {
	time.Sleep(defaultSleepDuration)
}

func (t *Single) buildTags(extraTags map[string]string) {
	t.tags = make(map[string]string)
	for k, v := range extraTags {
		t.tags[k] = v
	}
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
	res, _ := t.mult.ProcessLine(text)
	return res
}

func removeAnsiEscapeCodes(text []byte, run bool) []byte {
	if !run {
		return text
	}
	return ansi.Strip(text)
}
