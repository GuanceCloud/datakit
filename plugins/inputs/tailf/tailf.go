// +build !solaris

package tailf

import (
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
    logfiles = ["/tmp/tailf_test/**/*.log"]

    # glob filteer
    ignore = [""]

    # required
    source = "tailf"

    # [inputs.tailf.tags]
    # tags1 = "value1"
`
)

var l = logger.DefaultSLogger(inputName)

type Tailf struct {
	LogFiles []string          `toml:"logfiles"`
	Ignore   []string          `toml:"ignore"`
	Source   string            `toml:"source"`
	Tags     map[string]string `toml:"tags"`

	tailerConf tail.Config
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

	var fileList = getFileList(t.LogFiles, t.Ignore)
	var wg sync.WaitGroup

	for _, f := range fileList {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			for {
				if err := t.getLines(file); err != nil {
					time.Sleep(time.Second)
				} else {
					l.Infof("file %s is ending", f)
					break
				}
			}
		}(f)
	}
	wg.Wait()
	l.Info("exit")
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
	return false
}

func (t *Tailf) getLines(file string) error {
	tailer, err := tail.TailFile(file, t.tailerConf)
	if err != nil {
		l.Error("build tailer, %s", err)
		return err
	}
	defer tailer.Cleanup()

	var tags = make(map[string]string)
	for k, v := range t.Tags {
		tags[k] = v
	}
	tags["filename"] = file

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
			if err := io.NamedFeed(data, io.Logging, inputName); err != nil {
				l.Error(err)
				continue
			}
		}
	}
}
