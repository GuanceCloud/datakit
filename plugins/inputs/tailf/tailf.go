// +build !solaris

package tailf

import (
	"fmt"
	"strings"
	"time"

	"github.com/hpcloud/tail"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "tailf"

	defaultMeasurement = "tailf"

	sampleCfg = `
# [inputs.tailf]
# 	# Cannot be set to datakit.log
# 	# Directory and file paths
# 	paths = [""]
# 	
# 	# auto update the directory files
# 	update_files = false
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
	Paths         []string          `toml:"paths"`
	UpdateFiles   bool              `toml:"update_files"`
	FormBeginning bool              `toml:"from_beginning"`
	Tags          map[string]string `toml:"tags"`

	seek *tail.SeekInfo

	fileList []string
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

	t.updateTailers()

	t.foreachLines()
}

func (t *Tailf) updateTailers() {
	t.fileList = filterPath(t.Paths)
	l.Debugf("update file list: %v", t.fileList)

	for _, file := range t.fileList {
		tailer, err := tail.TailFile(file,
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
		t.tailers[file] = tailer
	}
}

func (t *Tailf) foreachLines() {

	count := 0
	for {
		time.Sleep(500 * time.Millisecond)
		for _, tailer := range t.tailers {
			if t.loopTailer(tailer) {
				return
			}
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		default:
			// nil
		}

		// update
		if t.UpdateFiles && count == 64 {
			t.updateTailers()
			count = 0
		}
		count++
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
				fmt.Printf("io.Feed data: %s\n", string(data))
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

	return io.MakeMetric(defaultMeasurement, tags, fields, time.Now())
}
