// +build !solaris

package tailf

import (
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
		time.Sleep(time.Second)
	__out:
		for _, tailer := range t.tailers {
			for {
				select {
				case line := <-tailer.Lines:
					data, err := t.parse(line)
					if err != nil {
						l.Error(err)
						continue
					}
					if err := io.Feed(data, io.Logging); err != nil {
						l.Error(err)
						continue
					}
					l.Debugf("feed %d bytes to io ok", len(data))

				case <-datakit.Exit.Wait():
					l.Info("exit")
					return

				default:
					goto __out
				}
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
		if count == 64 {
			t.updateTailers()
			count = 0
		}
		count++
	}
}

func (t *Tailf) parse(line *tail.Line) ([]byte, error) {
	// only '__content' kv
	var fields = make(map[string]interface{}, 1)

	if line.Err != nil {
		return nil, line.Err
	}

	text := strings.TrimRight(line.Text, "\r")
	fields["__content"] = text

	return io.MakeMetric(defaultMeasurement, t.Tags, fields, time.Now())
}
