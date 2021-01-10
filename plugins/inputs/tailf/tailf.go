// +build !solaris

package tailf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hpcloud/tail"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

type Tailf struct {
	LogFiles          []string          `toml:"logfiles"`
	Ignore            []string          `toml:"ignore"`
	Source            string            `toml:"source"`
	PipelinePath      string            `toml:"pipeline_path"`
	FromBeginning     bool              `toml:"from_beginning"`
	CharacterEncoding string            `toml:"character_encoding"`
	Tags              map[string]string `toml:"tags"`

	MultilineConfig MultilineConfig `toml:"multiline"`
	multiline       *Multiline

	decoder decoder

	tailerConf tail.Config

	runningFileList sync.Map
	wg              sync.WaitGroup
}

func (*Tailf) Catalog() string {
	return "log"
}

func (*Tailf) SampleConfig() string {
	return sampleCfg
}

func (*Tailf) Test() (result *inputs.TestResult, err error) {
	// 监听文件变更，无法进行测试
	result.Desc = "success"
	return
}

func (t *Tailf) Run() {
	l = logger.SLogger(inputName)

	if t.loadcfg() {
		return
	}

	l.Infof("tailf input started.")

	ticker := time.NewTicker(defaultDruation)
	defer ticker.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("waiting for all tailers to exit")
			t.wg.Wait()
			l.Info("exit")
			return

		case <-ticker.C:
			fileList := getFileList(t.LogFiles, t.Ignore)

			for _, file := range fileList {
				t.tailNewFiles(file)
			}

			if t.FromBeginning {
				// disable auto-discovery, ticker was unreachable
				ticker.Stop()
			}
		}
	}
}

func (t *Tailf) loadcfg() bool {
	var err error

	if t.PipelinePath == "" {
		t.PipelinePath = filepath.Join(datakit.InstallDir, t.Source+".p")
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if t.Source == "" {
			err = fmt.Errorf("tailf source cannot be empty")
			goto label
		}

		if t.decoder, err = NewDecoder(t.CharacterEncoding); err != nil {
			goto label
		}

		if t.multiline, err = t.MultilineConfig.NewMultiline(); err != nil {
			goto label
		}

		if _, err = pipeline.NewPipelineFromFile(t.PipelinePath); err != nil {
			goto label
		} else {
			break
		}

	label:
		l.Error(err)
		time.Sleep(time.Second)
	}

	var seek *tail.SeekInfo
	if !t.FromBeginning {
		seek = &tail.SeekInfo{
			Whence: 2, // seek is 2
			Offset: 0,
		}
	}

	t.tailerConf = tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  seek,
		MustExist: true,
		Poll:      false, // default watch method is "inotify"
		Pipe:      false,
		Logger:    tail.DiscardingLogger,
	}

	return false
}

func (t *Tailf) tailNewFiles(file string) {
	if _, ok := t.runningFileList.Load(file); ok {
		l.Debugf("file %s already tailing now", file)
		return
	}

	t.runningFileList.Store(file, nil)

	l.Debugf("start tail, %s", file)

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

		t.tailStart(file)
		t.runningFileList.Delete(file)
		l.Debugf("remove file %s from the list", file)
	}()
}

func (t *Tailf) tailStart(file string) {
	tailer, err := tail.TailFile(file, t.tailerConf)
	if err != nil {
		l.Error("build tailer, %s", err)
		return
	}
	defer tailer.Cleanup()

	tags := make(map[string]string)
	for k, v := range t.Tags {
		tags[k] = v
	}
	tags["filename"] = file

	t.receiver(tailer, tags)
}

func (t *Tailf) receiver(tailer *tail.Tail, tags map[string]string) {
	p, _ := pipeline.NewPipelineFromFile(t.PipelinePath)

	ticker := time.NewTicker(defaultDruation)
	defer ticker.Stop()

	var (
		buffer   bytes.Buffer
		textLine bytes.Buffer

		tailerOpen  = true
		channelOpen = true

		line  *tail.Line
		count int64
	)

	for {
		line = nil

		select {
		case <-datakit.Exit.Wait():
			l.Debugf("Tailing file %s is ending", tailer.Filename)
			return

		case line, tailerOpen = <-tailer.Lines:
			if !tailerOpen {
				channelOpen = false
			}

		case <-ticker.C:
			if count > 0 {
				if err := io.NamedFeed(buffer.Bytes(), io.Logging, inputName); err != nil {
					l.Error(err)
				}
				buffer = bytes.Buffer{}
				count = 0
			}

			_, statErr := os.Lstat(tailer.Filename)
			if os.IsNotExist(statErr) {
				l.Warnf("check file %s is not exist", tailer.Filename)
				return
			}
		}

		var text string

		if line != nil {
			text = strings.TrimRight(line.Text, "\r")

			if t.multiline.IsEnabled() {
				if text = t.multiline.ProcessLine(text, &textLine); text == "" {
					continue
				}
			}
		}

		if line == nil || !channelOpen || !tailerOpen {
			if text += t.multiline.Flush(&textLine); text == "" {
				if !channelOpen {
					l.Warnf("Tailing %s data channel is closed", tailer.Filename)
					return
				}
				continue
			}
		}

		if line != nil && line.Err != nil {
			l.Errorf("Tailing %q: %s", tailer.Filename, line.Err.Error())
			continue
		}

		decodeText, err := t.decoder.String(text)
		if err != nil {
			l.Errorf("decode error, %s", err)
			continue
		}

		var fields = make(map[string]interface{})
		if p != nil {
			fields, err = p.Run(decodeText).Result()
			if err != nil {
				l.Errorf("run pipeline error, %s", err)
				continue
			}
		} else {
			fields["message"] = decodeText
		}

		data, err := io.MakeMetric(t.Source, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
			continue
		}

		buffer.Write(data)
		buffer.WriteString("\n")
		count++

		if count >= metricFeedCount {
			if err := io.NamedFeed(buffer.Bytes(), io.Logging, inputName); err != nil {
				l.Error(err)
			}
			buffer = bytes.Buffer{}
			count = 0
		}

	}
}
