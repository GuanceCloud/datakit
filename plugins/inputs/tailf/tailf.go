// +build !solaris

package tailf

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hpcloud/tail"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "tailf"

	sampleCfg = `
[[inputs.tailf]]
    # required
    logfiles = ["/usr/local/cloudcare/dataflux/*/*.txt"]

    # glob filteer
    ignore = [""]

    # required
    source = "tailf"

    # [inputs.tailf.tags]
    # tags1 = "value1"
`

	updateFileListInterval   = time.Second * 3
	checkFileIsExistInterval = time.Minute * 20
	metricFeedCount          = 10
)

var l = logger.DefaultSLogger(inputName)

type Tailf struct {
	LogFiles []string          `toml:"logfiles"`
	Ignore   []string          `toml:"ignore"`
	Source   string            `toml:"source"`
	Tags     map[string]string `toml:"tags"`

	tailerConf tail.Config

	runningFileList sync.Map
	wg              sync.WaitGroup
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Tailf{}
	})
}

func (*Tailf) Catalog() string {
	return "log"
}

func (*Tailf) SampleConfig() string {
	return sampleCfg
}

func (t *Tailf) Run() {
	l = logger.SLogger(inputName)

	if t.loadcfg() {
		return
	}

	l.Infof("tailf input started.")

	ticker := time.NewTicker(updateFileListInterval)
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
			for _, f := range fileList {
				if _, ok := t.runningFileList.Load(f); !ok {
					t.runningFileList.Store(f, nil)
					l.Debugf("start tail, %s", f)

					t.wg.Add(1)
					go t.startTail(f)
				} else {
					l.Debugf("file %s is tailing now", f)
				}
			}
		}
	}
}

func (t *Tailf) loadcfg() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if t.Source == "" {
			l.Errorf("invalid source")
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	t.tailerConf = tail.Config{
		ReOpen: true,
		Follow: true,
		Location: &tail.SeekInfo{
			Whence: 2, // seek is 2
			Offset: 0,
		},
		MustExist: true,
		Poll:      false, // default watch method is "inotify"
		Pipe:      false,
		Logger:    tail.DiscardingLogger,
	}
	t.runningFileList = sync.Map{}
	t.wg = sync.WaitGroup{}
	return false
}

func (t *Tailf) startTail(file string) {
	defer t.wg.Done()

	err := t.getLines(file)
	// file is not exist or datakit exit
	if err == nil {
		t.runningFileList.Delete(file)
		l.Debugf("file %s is ending", file)
	}
}

func (t *Tailf) getLines(file string) error {
	tailer, err := tail.TailFile(file, t.tailerConf)
	if err != nil {
		l.Error("build tailer, %s", err)
		return err
	}
	defer tailer.Cleanup()

	tags := make(map[string]string)
	for k, v := range t.Tags {
		tags[k] = v
	}
	tags["filename"] = file

	var buffer bytes.Buffer
	count := 0

	for {
		select {
		case <-datakit.Exit.Wait():
			return nil

		case line := <-tailer.Lines:
			if line.Err != nil {
				l.Error("tailer lines, %s", err)
			}

			text := strings.TrimRight(line.Text, "\r")
			fields := map[string]interface{}{"__content": text}

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
				count = 0
				// not use buffer.Reset()
				buffer = bytes.Buffer{}
			}

		case <-time.After(checkFileIsExistInterval):
			if count > 0 {
				if err := io.NamedFeed(buffer.Bytes(), io.Logging, inputName); err != nil {
					l.Error(err)
				}
				count = 0
				buffer = bytes.Buffer{}
			}

			_, statErr := os.Lstat(file)
			if os.IsNotExist(statErr) {
				l.Warnf("check file %s is not exist", file)
				return nil
			}
		}
	}
}
