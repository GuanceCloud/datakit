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
#       # require
# 	regexs = ["*.log"]
#
# 	# Cannot be set to datakit.log
# 	# Directory and file paths
# 	paths = ["/tmp/tmux-1000"]
#
#       # require
#	source = "tailf"
# 	
# 	# auto update the directory files
# 	update_files = false
# 	
# 	update_files_cycle = 10s
#
# 	# Read the file from to beginning
# 	from_beginning = false
# 	
# 	# [inputs.tailf.tags]
# 	# tags1 = "tags1"
# 
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
	UpdateFilesCycle time.Duration     `toml:"update_files_cycle"`
	FormBeginning    bool              `toml:"from_beginning"`
	Tags             map[string]string `toml:"tags"`

	seek       *tail.SeekInfo
	updateTick *time.Ticker

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

	t.initcfg()

	t.buildRegexp()

	t.updateTailers()

	t.foreachLines()
}

func (t *Tailf) initcfg() {
	for {
		if t.Source == "" {
			l.Errorf("invalid source")
			time.Sleep(time.Second)
		} else {
			break

		}

	}

	if t.UpdateFiles {
		for {
			if t.UpdateFilesCycle > 0 {
				t.updateTick = time.NewTicker(t.UpdateFilesCycle)
				break
			} else {
				l.Errorf("invalid cycle duration")
				time.Sleep(time.Second)
			}
		}
	}

	if t.Tags == nil {
		t.Tags = make(map[string]string)
	}

	t.tailers = make(map[string]*tail.Tail)

	if t.FormBeginning {
		t.seek = &tail.SeekInfo{
			Whence: 0,
			Offset: 0,
		}
	} else {
		t.seek = &tail.SeekInfo{
			Whence: 2,
			Offset: 0,
		}
	}
}

func (t *Tailf) buildRegexp() {

	for _, regexStr := range t.Regexs {
		for {
			select {
			case <-datakit.Exit.Wait():
				l.Info("exit")
				return
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

	if testAssert {
		for fn := range t.fileList {
			l.Debugf("file list: %s", fn)
		}
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

func (t *Tailf) foreachLines() {

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

		// update
		case <-t.updateTick.C:
			if t.UpdateFiles {
				t.updateTailers()
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
			if err := io.Feed(data, io.Logging); err != nil {
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
	tags["__source"] = t.Source
	tags["filename"] = filename

	text := strings.TrimRight(line.Text, "\r")
	fields["__content"] = text

	return io.MakeMetric(defaultMeasurement, tags, fields, time.Now())
}
