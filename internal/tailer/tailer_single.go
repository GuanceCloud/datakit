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
	readChanSize         = 32
)

type TailerSingle struct {
	opt      *Option
	file     *os.File
	filename string

	decoder  *encoding.Decoder
	mult     *Multiline
	pipeline *pipeline.Pipeline

	tags map[string]string

	outputChan chan []byte
	stop       chan interface{}

	err error
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

	t.outputChan = make(chan []byte, readChanSize)
	t.stop = make(chan interface{}, 1)
	t.filename = t.file.Name()
	t.tags = t.buildTags(opt.GlobalTags)

	return t, nil
}

func (t *TailerSingle) Start() {
	go t.forwardMessage()
	go t.readFroever()
}

func (t *TailerSingle) Stop() {
	t.stop <- nil
	t.file.Close()
	t.opt.done <- t.filename
	t.opt.log.Debugf("closing %s", t.filename)
	select {
	case <-t.outputChan:
		// nil
	default:
		close(t.outputChan)
	}
}

func (t *TailerSingle) forwardMessage() {
	var textBlock []byte
	for output := range t.outputChan {
		lines := bytes.Split(output, []byte{'\n'})
		if len(lines) == 0 {
			continue
		}

		if len(textBlock) != 0 {
			lines[0] = append(textBlock, lines[0]...)
			textBlock = textBlock[:0]
		}

		if len(lines[len(lines)-1]) != 0 {
			textBlock = lines[len(lines)-1]
			lines = lines[:len(lines)-1]
		}

		for _, line := range lines {
			if len(line) == 0 {
				continue
			}

			text, err := t.decode(string(line))
			if err != nil {
				t.opt.log.Debug(err)
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
	err := newLogs(text).
		pipeline(t.pipeline).
		checkFieldsLength().
		addStatus(t.opt.DisableAddStatusField).
		ignoreStatus(t.opt.IgnoreStatus).
		takeTime().
		point(t.opt.Source, t.tags).
		feed(t.opt.InputName).
		error()

	return err
}

func (t *TailerSingle) read() (int, error) {
	buf := make([]byte, readBuffSize)
	n, err := t.file.Read(buf)
	if err != nil && err != io.EOF {
		t.opt.log.Debug(err)
		return 0, err
	}

	if n == 0 {
		return 0, nil
	}
	t.outputChan <- buf[:n]
	return n, nil

}

func (t *TailerSingle) readFroever() {
	defer t.Stop()
	for {
		n, err := t.read()
		if err != nil {
			t.opt.log.Debugf("failed of read data from file %s", t.filename)
			return
		}

		select {
		case <-t.stop:
			t.opt.log.Debugf("stop reading data from file %s", t.filename)
			return
		default:
			if n == 0 {
				t.wait()
			}
		}
	}
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
