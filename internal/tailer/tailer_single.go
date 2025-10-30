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
	"strings"
	"sync/atomic"
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
	maxRetryAttempts     = 3
)

type Single struct {
	config   *config
	filepath string
	feedName string

	file           *os.File
	inode          string
	recordKey      string
	insideFilepath string
	reader         reader.Reader

	offset        int64
	readLines     int64
	partialBuffer bytes.Buffer

	decoder   *encoding.Decoder
	multiline *multiline.Multiline

	extraTags map[string]string

	updateChan chan []Option
	cancelFunc context.CancelFunc

	log *logger.Logger

	// 状态管理
	isRunning  atomic.Bool
	startTime  time.Time
	retryCount int32
}

func NewTailerSingle(filepath string, opts ...Option) (*Single, error) {
	if filepath == "" {
		return nil, fmt.Errorf("filepath cannot be empty")
	}

	t := &Single{
		filepath:   filepath,
		updateChan: make(chan []Option, defaultUpdateChannelSize),
		log:        logger.SLogger("logging/unknown"),
	}

	if err := t.applyOptions(opts); err != nil {
		return nil, fmt.Errorf("failed to apply options: %w", err)
	}
	if err := t.setupFile(); err != nil {
		return nil, fmt.Errorf("failed to setup file: %w", err)
	}

	return t, nil
}

func (t *Single) applyOptions(opts []Option) error {
	// 初始化配置
	t.config = defaultConfig()
	for _, opt := range opts {
		opt(t.config)
	}

	t.log = logger.SLogger("logging/" + t.config.source)

	t.feedName = dkio.FeedSource("logging", t.config.source)
	if t.config.storageIndex != "" {
		t.feedName = dkio.FeedSource(t.feedName, t.config.storageIndex)
	}

	if t.config.insideFilepathFunc != nil {
		t.insideFilepath = t.config.insideFilepathFunc(t.filepath)
	}

	if t.config.characterEncoding != "" {
		t.decoder, _ = encoding.NewDecoder(t.config.characterEncoding)
	}
	t.multiline, _ = multiline.New(t.config.multilinePatterns, multiline.WithMaxLength(int(t.config.maxMultilineLength)))

	t.extraTags = make(map[string]string)
	for k, v := range t.config.extraTags {
		if t.shouldAddField(k) {
			t.extraTags[k] = v
		}
	}

	return nil
}

func (t *Single) setupFile() error {
	t.log.Debugf("setting up file: %s", t.filepath)

	var err error
	t.file, err = openfile.OpenFile(t.filepath)
	if err != nil {
		t.log.Warnf("failed to open file %s: %s", t.filepath, err)
		return fmt.Errorf("failed to open file %s: %w", t.filepath, err)
	}

	t.reader = reader.NewReader(t.file)
	t.inode = openfile.Inode(t.filepath)
	t.recordKey = openfile.FileKey(t.filepath)

	t.log.Debugf("file opened successfully: %s, inode: %s", t.filepath, t.inode)

	if err := t.seekOffset(); err != nil {
		t.log.Warnf("failed to set position for file %s: %s", t.filepath, err)
		// 关闭文件句柄，避免资源泄漏
		if t.file != nil {
			if closeErr := t.file.Close(); closeErr != nil {
				t.log.Warnf("failed to close file %s: %s", t.filepath, closeErr)
			}
			t.file = nil
		}
		return fmt.Errorf("failed to set position for file %s: %w", t.filepath, err)
	}

	t.log.Debugf("file setup completed: %s, offset: %d", t.filepath, t.offset)
	return nil
}

func (t *Single) Run(ctx context.Context) {
	if !t.isRunning.CompareAndSwap(false, true) {
		t.log.Warn("single tailer is already running")
		return
	}

	t.startTime = time.Now()
	t.log.Infof("starting tailer for file: %s, source: %s", t.filepath, t.config.source)

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		t.isRunning.Store(false)
		t.cleanup()
	}()

	t.cancelFunc = cancel

	t.forwardMessage(ctx)
	t.log.Infof("tailer for file %s has stopped, source: %s", t.filepath, t.config.source)
}

func (t *Single) Close() {
	t.log.Infof("closing tailer for file: %s, source: %s", t.filepath, t.config.source)
	if t.cancelFunc != nil {
		t.cancelFunc()
	}
}

func (t *Single) UpdateOptions(newOpts []Option) {
	select {
	case t.updateChan <- newOpts:
		// 配置更新已发送
	default:
		t.log.Warnf("update channel full, dropping options update")
	}
}

func (t *Single) cleanup() {
	t.log.Debugf("cleaning up file: %s", t.filepath)

	t.recordPosition()
	t.closeFile()

	t.log.Infof("cleanup completed for file: %s", t.filepath)
}

func (t *Single) seekOffset() error {
	if t.file == nil {
		return fmt.Errorf("unexpected file pointer")
	}

	var err error

	// 使用记录的位置
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
			t.log.Infof("position %d larger than file size %d, file may have been truncated", offset, size)
		} else {
			t.offset, err = t.file.Seek(offset, io.SeekStart)
			if err != nil {
				t.log.Warnf("seek to offset %d failed: %s", offset, err)
				return err
			}

			t.log.Infof("set position %d for file %s", offset, t.filepath)
			return nil
		}
	}
	t.log.Infof("not found position for recorder key %s", t.recordKey)

	// 使用 fromBeginning 配置
	if t.config.fromBeginning {
		t.offset, err = t.file.Seek(0, io.SeekStart)
		t.log.Infof("set start position for file %s", t.filepath)
		return err
	}

	// 使用文件大小阈值
	if stat, _err := os.Stat(t.filepath); _err == nil {
		size := stat.Size()
		if size < t.config.fileSizeThreshold {
			t.offset, err = t.file.Seek(0, io.SeekStart)
			t.log.Infof("set start position for file %s, because file size %d < %d",
				t.filepath, size, t.config.fileSizeThreshold)
			return err
		}
	}

	// 使用文件末尾
	t.offset, err = t.file.Seek(0, io.SeekEnd)
	t.log.Infof("set end position for file %s and offset=%d", t.filepath, t.offset)

	return err
}

func (t *Single) recordPosition() {
	if t.offset <= 0 {
		return
	}

	c := &recorder.MetaData{Source: t.config.source, Offset: t.offset}

	if err := recorder.SetAndFlush(t.recordKey, c); err != nil {
		t.log.Debugf("recording cache %s err: %s", c, err)
	}
}

func (t *Single) closeFile() {
	if t.file == nil {
		t.log.Debugf("file already closed: %s", t.filepath)
		return
	}

	if err := t.file.Close(); err != nil {
		t.log.Warnf("failed to close file %s: %s", t.filepath, err)
	} else {
		t.log.Debugf("file closed successfully: %s", t.filepath)
	}

	t.file = nil
}

func (t *Single) reopen() error {
	t.closeFile()

	var err error
	t.file, err = openfile.OpenFile(t.filepath)
	if err != nil {
		return fmt.Errorf("unable to reopen file %s: %w", t.filepath, err)
	}

	ret, err := t.file.Seek(0, io.SeekStart)
	if err != nil {
		t.log.Warnf("seek to start failed after reopen: %s", err)
		return err
	}

	t.reader.SetReader(t.file)
	t.offset = ret
	t.inode = openfile.Inode(t.filepath)
	t.recordKey = openfile.FileKey(t.filepath)

	t.log.Infof("reopened file %s, offset reset to %d", t.filepath, t.offset)
	rotateCounter.WithLabelValues(t.config.source, t.filepath).Inc()
	return nil
}

func (t *Single) forwardMessage(ctx context.Context) {
	checkTicker := time.NewTicker(checkInterval)
	defer checkTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.handleContextCancellation()
			return
		case newOpts := <-t.updateChan:
			t.handleConfigUpdate(newOpts)
		case <-checkTicker.C:
			if t.handleFileCheck() {
				return
			}
		default:
			if t.handleFileRead() {
				return
			}
		}
	}
}

func (t *Single) handleContextCancellation() {
	t.flushCache()
	t.log.Infof("context canceled, exiting forwardMessage for file: %s, source: %s", t.filepath, t.config.source)
}

func (t *Single) handleConfigUpdate(newOpts []Option) {
	t.log.Debugf("received options update for file: %s", t.filepath)
	if err := t.applyOptions(newOpts); err != nil {
		t.log.Warnf("failed to apply new options for file %s: %s", t.filepath, err)
	} else {
		t.log.Infof("new options applied successfully for file: %s, source: %s", t.filepath, t.config.source)
	}
}

func (t *Single) handleFileCheck() bool {
	did, _ := openfile.DidRotate(t.file, t.offset)
	exist := openfile.FileExists(t.filepath)

	if did || !exist {
		t.log.Debugf("file %s rotated or removed, reading to EOF", t.filepath)
		if shouldExit := t.readToEOF(context.Background()); shouldExit {
			t.log.Debugf("readToEOF indicated exit, stopping forwardMessage for file: %s", t.filepath)
			return true
		}
	}

	if did {
		t.log.Infof("file rotated, reopening: %s", t.filepath)
		if err := t.reopen(); err != nil {
			t.log.Warnf("failed to reopen the file %s, err: %s", t.filepath, err)
			return true
		}
	}

	if !openfile.FileIsActive(t.filepath, t.config.ignoreDeadLog) {
		t.log.Infof("file %s has been inactive for longer than %s or has been removed, exiting", t.filepath, t.config.ignoreDeadLog)
		return true
	}

	return false
}

func (t *Single) handleFileRead() bool {
	if err := t.readOnce(); err != nil {
		if !errors.Is(err, reader.ErrReadEmpty) {
			t.log.Warnf("failed to read data from file %s, error: %s", t.filepath, err)
			// 如果是文件读取错误，尝试重新打开文件
			if t.shouldRetryRead(err) {
				t.log.Debugf("attempting to recover from read error for file: %s", t.filepath)
				if reopenErr := t.reopen(); reopenErr != nil {
					t.log.Errorf("failed to reopen file %s after read error: %s", t.filepath, reopenErr)
					return true
				}
			}
		}
		t.flushCache()
		time.Sleep(defaultSleepDuration)
	}
	return false
}

func (t *Single) readToEOF(ctx context.Context) (shouldExit bool) {
	t.log.Infof("file %s has been rotated or removed, current offset %d, reading to EOF", t.filepath, t.offset)
	for {
		select {
		case <-ctx.Done():
			t.log.Debugf("context canceled during readToEOF for file %s", t.filepath)
			return true

		case <-datakit.Exit.Wait():
			t.log.Debugf("global exit signal received during readToEOF for file %s", t.filepath)
			return true

		default:
			if err := t.readOnce(); err != nil {
				if !errors.Is(err, reader.ErrReadEmpty) {
					t.log.Warnf("read to EOF error: %s", err)
				}
				t.log.Infof("finished reading to EOF for file %s", t.filepath)
				return false
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
	t.process(t.config.mode, lines)

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
			parseFailCounter.WithLabelValues(t.config.source, t.filepath, t.config.mode.String()).Inc()
			continue
		}

		if msg.IsPartial {
			t.partialBuffer.Write(msg.Log)
			continue
		}

		var originalText []byte
		if t.partialBuffer.Len() != 0 {
			t.partialBuffer.Write(msg.Log)
			originalText = t.partialBuffer.Bytes()
			t.partialBuffer.Reset()
		} else {
			originalText = msg.Log
		}

		text, err := t.decode(originalText)
		if err != nil {
			decodeErrorCounter.WithLabelValues(t.config.source, t.config.characterEncoding, err.Error()).Inc()
		}

		text = removeAnsiEscapeCodes(text, t.config.removeAnsiEscapeCodes)

		finalText := t.processMultiline(multiline.TrimRightSpace(text))
		if len(finalText) == 0 {
			continue
		}

		pending = append(pending, finalText)
	}

	t.feed(pending)
}

func (t *Single) feed(pending [][]byte) {
	// 转发到远程
	if t.config.forwardFunc != nil {
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

		if t.config.enableDebugFields {
			fields["log_read_offset"] = t.offset
			fields["log_file_inode"] = t.inode
		}

		err := t.config.forwardFunc(t.filepath, string(text), fields)
		if err != nil {
			t.log.Warnf("failed to forward text from file %s, error: %s", t.filepath, err)
		}
	}
}

const (
	// LogTimeStep 日志时间步长，用于确保日志顺序.
	LogTimeStep = time.Microsecond // 1us
)

func (t *Single) feedToIO(pending [][]byte) {
	var (
		points  = []*point.Point{}
		opts    = append(point.DefaultLoggingOptions(), point.WithPrecheck(false))
		timeNow = ntp.Now().Add(-time.Duration(len(pending)) * LogTimeStep)
	)

	for i, cnt := range pending {
		t.readLines++

		kvs := make(point.KVs, 0, len(t.extraTags)+4)
		kvs = kvs.Add(constants.FieldMessage, string(cnt))

		if t.shouldAddField("filepath") {
			kvs = kvs.Add("filepath", t.filepath)
		}
		if t.shouldAddField("log_read_lines") {
			kvs = kvs.Add("log_read_lines", t.readLines)
		}
		if t.shouldAddField(constants.FieldStatus) {
			kvs = kvs.AddTag(constants.FieldStatus, constants.DefaultStatus)
		}

		if t.shouldAddField("inside_filepath") && t.insideFilepath != "" {
			kvs = kvs.Add("inside_filepath", t.insideFilepath)
		}

		for key, value := range t.extraTags {
			kvs = kvs.AddTag(key, value)
		}

		if t.config.enableDebugFields {
			kvs = kvs.Add("log_read_offset", t.offset)
			kvs = kvs.Add("log_file_inode", t.inode)
		}

		// only the message field is present, with no match in the whitelist
		// discard this data
		if len(kvs) == 1 {
			discardCounter.WithLabelValues(t.config.source, t.filepath).Inc()
			continue
		}

		pt := point.NewPoint(t.config.source, kvs, opts...)

		// 此处设置每条日志的时间差为 1us, 这样日志查看器中的日志显示顺序跟当前的采集顺序就保持一致了.
		// 此处如果这批日志的时间戳都一样, 在查看器中看到日志将可能随机展示(因为查看器默认按照 point 的
		// 时间戳来倒排显示)
		// 注意, 此处这个时间还是可以在后续的 pipeline 被改写.
		pt.SetTime(timeNow.Add(time.Duration(i) * LogTimeStep))

		points = append(points, pt)
	}

	if len(points) == 0 {
		return
	}

	if err := t.config.feeder.Feed(
		point.Logging,
		points,
		dkio.WithStorageIndex(t.config.storageIndex),
		dkio.WithSource(t.feedName),
		dkio.WithPipelineOption(&lang.LogOption{
			ScriptMap: map[string]string{t.config.source: t.config.pipeline},
		}),
	); err != nil {
		t.log.Errorf("feed %d points failed: %s, logging block-mode off, ignored", len(points), err)
	}
}

func (t *Single) flushCache() {
	if t.multiline != nil && t.multiline.BuffLength() > 0 {
		text := t.multiline.Flush()
		logstr := removeAnsiEscapeCodes(text, t.config.removeAnsiEscapeCodes)
		t.feed([][]byte{logstr})
	}
}

func (t *Single) shouldAddField(name string) bool {
	if len(t.config.fieldWhitelist) == 0 {
		return true
	}

	for _, key := range t.config.fieldWhitelist {
		if key == name {
			return true
		}
	}
	return false
}

func (t *Single) decode(text []byte) ([]byte, error) {
	if t.decoder == nil {
		return text, nil
	}
	return t.decoder.Bytes(text)
}

func (t *Single) processMultiline(text []byte) []byte {
	if !t.config.enableMultiline || t.multiline == nil {
		return text
	}
	res, state := t.multiline.ProcessLine(text)
	if state == multiline.FlushPartial {
		multilineCounter.WithLabelValues(t.config.source, state.String()).Inc()
	}
	return res
}

func (t *Single) shouldRetryRead(err error) bool {
	if err == nil {
		return false
	}

	// 检查重试次数
	if atomic.LoadInt32(&t.retryCount) >= maxRetryAttempts {
		t.log.Warnf("max retry attempts reached for file %s", t.filepath)
		return false
	}

	// 检查文件是否仍然存在
	if !openfile.FileExists(t.filepath) {
		t.log.Debugf("file %s no longer exists, not retrying", t.filepath)
		return false
	}

	// 检查是否是临时性错误（如权限问题、文件被锁定等）
	if strings.Contains(err.Error(), "permission denied") ||
		strings.Contains(err.Error(), "access denied") ||
		strings.Contains(err.Error(), "file is locked") ||
		strings.Contains(err.Error(), "resource temporarily unavailable") {
		atomic.AddInt32(&t.retryCount, 1)
		return true
	}

	return false
}

func removeAnsiEscapeCodes(text []byte, run bool) []byte {
	if !run {
		return text
	}
	return ansi.Strip(text)
}
