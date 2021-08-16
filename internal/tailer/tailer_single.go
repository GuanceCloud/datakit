package tailer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/encoding"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	defaultSleepDuration = time.Second
	readBuffSize         = 1024 * 4
	timeoutDuration      = time.Second * 3
)

type TailerSingle struct {
	opt      *Option
	file     *os.File
	filename string

	decoder  *encoding.Decoder
	mult     *Multiline
	pipeline *pipeline.Pipeline

	tags map[string]string

	stop chan interface{}
	err  error
}

func NewTailerSingle(filename string, opt *Option) (*TailerSingle, error) {
	if opt == nil {
		return nil, fmt.Errorf("option cannot be null pointer")
	}

	t := &TailerSingle{opt: opt}

	t.decoder, t.err = encoding.NewDecoder(opt.CharacterEncoding)
	t.mult, t.err = NewMultiline(opt.Match)
	t.file, t.err = os.Open(filename)
	if t.err != nil {
		return nil, t.err
	}

	if !opt.FromBeginning {
		if _, err := t.file.Seek(0, os.SEEK_END); err != nil {
			return nil, t.err
		}
	}

	if opt.Pipeline != "" {
		p, err := pipeline.NewPipelineFromFile(opt.Pipeline)
		if err == nil {
			t.pipeline = p
		}
	}

	t.stop = make(chan interface{})
	t.filename = t.file.Name()
	t.tags = t.buildTags(opt.GlobalTags)

	return t, nil
}

func (t *TailerSingle) Start() {
	go t.forwardMessage()
}

func (t *TailerSingle) Stop() {
	select {
	case <-t.stop:
		// nil
	default:
		close(t.stop)
	}

	t.file.Close()
	t.opt.done <- t.filename
	t.opt.log.Debugf("closing %s", t.filename)
}

func (t *TailerSingle) forwardMessage() {
	var (
		b       = &buffer{}
		timeout = time.NewTicker(timeoutDuration)
		lines   []string
		readNum int
		err     error
	)
	defer timeout.Stop()

	for {
		b.buf = b.buf[:0]

		select {
		case <-t.stop:
			t.opt.log.Debugf("stop reading data from file %s", t.filename)
			return
		case <-timeout.C:
			if err := t.processText(t.mult.Flush()); err != nil {
				t.opt.log.Debug(err)
			}
		default:
			// nil
		}

		b.buf, readNum, err = t.read()
		if err != nil {
			t.opt.log.Debugf("failed of read data from file %s", t.filename)
			return
		}
		if readNum == 0 {
			t.wait()
			continue
		}

		lines = b.split()

		for _, line := range lines {
			if line == "" {
				continue
			}

			text, err := t.decode(line)
			if err != nil {
				t.opt.log.Debugf("decode '%s' error: %s", t.opt.CharacterEncoding, err)
				err = feed(t.opt.InputName, t.opt.Source, t.tags, line)
				if err != nil {
					t.opt.log.Debug(err)
				}
			}

			text = t.multiline(text)
			if text == "" {
				continue
			}

			err = t.processText(text)
			if err != nil {
				t.opt.log.Debug(err)
			}
		}
	}
}

func (t *TailerSingle) processText(text string) error {
	if text == "" {
		return nil
	}

	err := NewLogs(text).
		Pipeline(t.pipeline).
		CheckFieldsLength().
		AddStatus(t.opt.DisableAddStatusField).
		IgnoreStatus(t.opt.IgnoreStatus).
		TakeTime().
		Point(t.opt.Source, t.tags).
		Feed(t.opt.InputName).
		Error()

	return err
}

func (t *TailerSingle) read() ([]byte, int, error) {
	buf := make([]byte, readBuffSize)
	n, err := t.file.Read(buf)
	if err != nil && err != io.EOF {
		t.opt.log.Debug(err)
		return nil, 0, err
	}
	return buf[:n], n, nil
}

func (t *TailerSingle) wait() {
	time.Sleep(defaultSleepDuration)
}

func (t *TailerSingle) buildTags(globalTags map[string]string) map[string]string {
	var tags = make(map[string]string)
	for k, v := range globalTags {
		tags[k] = v
	}
	if _, ok := tags["filename"]; !ok {
		tags["filename"] = filepath.Base(t.filename)
	}
	return tags
}

func (t *TailerSingle) decode(text string) (str string, err error) {
	if t.decoder == nil {
		return text, nil
	}
	return t.decoder.String(text)
}

func (t *TailerSingle) multiline(text string) string {
	if t.mult == nil {
		return text
	}
	return t.mult.ProcessLine(text)
}

func (t *TailerSingle) multilineFlush() string {
	if t.mult == nil {
		return ""
	}
	return t.mult.Flush()
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
		b.previousBlock = b.previousBlock[:0]
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
