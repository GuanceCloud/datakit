// +build !solaris

package tailf

import (
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/hpcloud/tail"
	"github.com/mattn/go-zglob"

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
	loopReadFileTimeMillisecond = 100
)

var l = logger.DefaultSLogger(inputName)

type Tailf struct {
	LogFiles []string          `toml:"logfiles"`
	Ignore   []string          `toml:"ignore"`
	Source   string            `toml:"source"`
	Tags     map[string]string `toml:"tags"`

	seek    *tail.SeekInfo
	tailers map[string]*tailer
}

type tailer struct {
	tl    *tail.Tail
	exist bool
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

	if t.initcfg() {
		return
	}

	t.getLines()
}

func (t *Tailf) initcfg() bool {
	// check source
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

	if t.Tags == nil {
		t.Tags = make(map[string]string)
	}

	t.tailers = make(map[string]*tailer)

	t.seek = &tail.SeekInfo{
		Whence: 2,
		Offset: 0,
	}

	return false
}

func (t *Tailf) updateTailers() error {

	fileList, err := getFileList(t.LogFiles, t.Ignore)
	if err != nil {
		return err
	}

	// set false
	for _, tailer := range t.tailers {
		tailer.exist = false
	}

	for _, fn := range fileList {
		if _, ok := t.tailers[fn]; !ok {
			tl, err := tail.TailFile(fn,
				tail.Config{
					ReOpen:    true,
					Follow:    true,
					Location:  t.seek,
					MustExist: true,
					// defaultWatchMethod is "inotify"
					Poll:   false,
					Pipe:   false,
					Logger: tail.DiscardingLogger,
				})
			if err != nil {
				continue
			}

			t.tailers[fn] = &tailer{tl: tl}
		}
		t.tailers[fn].exist = true
	}

	// filter the not exist file
	for key, tailer := range t.tailers {
		if !tailer.exist {
			tailer.tl.Cleanup()
			delete(t.tailers, key)
		}
	}

	return nil
}

func (t *Tailf) getLines() {
	l.Infof("tailf input started.")

	loopTick := time.NewTicker(loopReadFileTimeMillisecond * time.Millisecond)
	defer loopTick.Stop()
	updateTick := time.NewTicker(time.Second)
	defer updateTick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-loopTick.C:
			for _, tailer := range t.tailers {
				// return true is 'exit'
				if t.loopTailer(tailer.tl) {
					return
				}
			}

		case <-updateTick.C:
			if err := t.updateTailers(); err != nil {
				l.Error("update file list error: %s", err.Error())
			}
		}
	}
}

// loopTailer return bool of is "exit"
func (t *Tailf) loopTailer(tl *tail.Tail) bool {
	for {
		select {
		case line := <-tl.Lines:
			data, err := t.parseLine(line, tl.Filename)
			if err != nil {
				l.Error(err)
				continue
			}
			if err := io.NamedFeed(data, io.Logging, inputName); err != nil {
				l.Error(err)
				continue
			}

		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			return false
		}
	}
}

func (t *Tailf) parseLine(line *tail.Line, filename string) ([]byte, error) {
	// only '__content' kv

	if line.Err != nil {
		return nil, line.Err
	}

	var tags = make(map[string]string)
	var fields = make(map[string]interface{}, 1)
	for k, v := range t.Tags {
		tags[k] = v
	}
	tags["filename"] = filename

	text := strings.TrimRight(line.Text, "\r")
	fields["__content"] = text

	return io.MakeMetric(t.Source, tags, fields, time.Now())
}

func getFileList(filesGlob, ignoreGlob []string) ([]string, error) {

	var matches, passlist []string
	for _, f := range filesGlob {
		matche, err := zglob.Glob(f)
		if err != nil {
			return nil, err
		}
		matches = append(matches, matche...)
	}

	var globs []glob.Glob
	for _, ig := range ignoreGlob {
		g, err := glob.Compile(ig)
		if err != nil {
			return nil, err
		}
		globs = append(globs, g)
	}

	for _, match := range matches {
		for _, g := range globs {
			if g.Match(match) {
				goto __NEXT
			}
		}
		passlist = append(passlist, match)
	__NEXT:
	}

	return passlist, nil
}
