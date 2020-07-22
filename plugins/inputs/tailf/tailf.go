// +build !solaris

package tailf

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hpcloud/tail"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "tailf"

	defaultMeasurement = "tailf"

	sampleCfg = `
# [[inputs.tailf]]
#	# use leftmost-first matching is the same semantics that Perl, Python
#	# hit basename
#	# require
#	regexs = [".log"]
#
#	# Cannot be set to datakit.log
#	# Directory and file paths
#	paths = ["/tmp/tailf_test"]
#
#	# require
#	source = "tailf"
#
#	# auto update the directory files
#	update_files = false
#
#	# valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
#	update_files_cycle = "10s"
#
#	# [inputs.tailf.tags]
#	# tags1 = "tags1"
`
)

var (
	l *logger.Logger

	testAssert = false
)

type Tailf struct {
	Regexs           []string          `toml:"regexs"`
	Paths            []string          `toml:"paths"`
	Source           string            `toml:"source"`
	UpdateFiles      bool              `toml:"update_files"`
	UpdateFilesCycle string            `toml:"update_files_cycle"`
	Tags             map[string]string `toml:"tags"`

	seek      *tail.SeekInfo
	regexList []*regexp.Regexp

	fileList map[string]interface{}
	tailers  map[string]*tail.Tail
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Tailf{}
	})
}

func (_ *Tailf) Catalog() string {
	return "log"
}

func (_ *Tailf) SampleConfig() string {
	return sampleCfg
}

func (t *Tailf) Run() {
	l = logger.SLogger(inputName)

	if t.initcfg() {
		return
	}

	if t.UpdateFiles {
		d, err := time.ParseDuration(t.UpdateFilesCycle)
		if err != nil || d <= 0 {
			l.Errorf("invalid duration of update_files_cycle")
			return
		}
		t.getLinesUpdate(d)
	} else {
		t.getLines()
	}

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

	// build regexp
	for _, regexStr := range t.Regexs {
		for {
			select {
			case <-datakit.Exit.Wait():
				l.Info("exit")
				return true
			default:
				// nil
			}

			if re, err := regexp.Compile(regexStr); err == nil {
				t.regexList = append(t.regexList, re)
				break
			} else {
				l.Errorf("invalid regex, err: %s", err.Error())
				time.Sleep(time.Second)
			}
		}
	}

	if t.Tags == nil {
		t.Tags = make(map[string]string)
	}

	t.tailers = make(map[string]*tail.Tail)

	t.seek = &tail.SeekInfo{
		Whence: 2,
		Offset: 0,
	}

	return false
}

func (t *Tailf) filterPath() {
	var fileList = getFileList(t.Paths)
	t.fileList = make(map[string]interface{})

	if testAssert {
		l.Debugf("file list: %v", fileList)
	}

	// if error not nil, datakitlog is ""
	datakitlog, _ := filepath.Abs(config.Cfg.MainCfg.Log)

	for _, fn := range fileList {
		if fn == datakitlog {
			continue
		}

		if info, err := os.Stat(fn); err != nil || info.IsDir() {
			continue
		}

		for _, re := range t.regexList {
			if re.MatchString(filepath.Base(fn)) {
				t.fileList[fn] = nil
				break
			}
		}
	}

	for fn := range t.fileList {
		contentType, err := getFileContentType(fn)
		if err != nil {
			continue
		}

		if _, ok := typeWhiteList[contentType]; !ok {
			l.Warnf("%s is not text file", fn)
		}
	}
}

func (t *Tailf) updateTailers() {
	t.filterPath()

	for fn := range t.fileList {
		l.Debugf("update new file list: %s", fn)
	}

	for fn := range t.fileList {

		if _, ok := t.tailers[fn]; !ok {

			tailer, err := tail.TailFile(fn,
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

			t.tailers[fn] = tailer
		}
	}
}

func (t *Tailf) getLinesUpdate(d time.Duration) {
	l.Infof("tailf input started...")
	t.updateTailers()

	loopTick := time.NewTicker(100 * time.Millisecond)
	updateTick := time.NewTicker(d)

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-loopTick.C:
			for _, tailer := range t.tailers {
				// return true is 'exit'
				if t.loopTailer(tailer) {
					return
				}
			}
		// update
		case <-updateTick.C:
			if t.UpdateFiles {
				t.updateTailers()
			}
		}
	}
}

func (t *Tailf) getLines() {
	l.Infof("tailf input started...")
	t.updateTailers()
	loopTick := time.NewTicker(100 * time.Millisecond)

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-loopTick.C:
			for _, tailer := range t.tailers {
				// return true is 'exit'
				if t.loopTailer(tailer) {
					return
				}
			}
		}
	}
}

// loopTailer return bool of is "exit"
func (t *Tailf) loopTailer(tailer *tail.Tail) bool {
	for {
		select {
		case line := <-tailer.Lines:
			data, err := t.parseLine(line, tailer.Filename)
			if err != nil {
				l.Error(err)
				continue
			}
			if testAssert {
				l.Debugf("io.Feed data: %s\n", string(data))
				continue
			}
			if err := io.NamedFeed(data, io.Logging, inputName); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))

		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true

		default:
			// clean
			if _, ok := t.fileList[tailer.Filename]; !ok {
				tailer.Cleanup()
				delete(t.tailers, tailer.Filename)
			}
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
