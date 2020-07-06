// +build !solaris

package tailf

import (
	"strings"
	"time"

	"github.com/hpcloud/tail"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "tailf"

	sampleCfg = `
# [[tailf]]
#	## 文件路径名。不能设置为 datakit 日志文件。
#	filename = ""
#
#       ## 是否从文件首部读取
#	from_beginning = false
#
#       ## 是否是一个pipe
#	pipe = false
#
#	## 通知方式，默认是 inotify 即由操作系统进行变动通知
#       ## 可以设为 poll 改为轮询文件的方式
#	watch_method = "inotify"
#
#       ## 数据源名字，不可为空
#       source = ""
`
)

var (
	l *zap.SugaredLogger
)

type (
	Tailf struct {
		C []Impl `toml:"tailf"`
	}

	Impl struct {
		File          string `toml:"filename"`
		FormBeginning bool   `toml:"from_beginning"`
		Pipe          bool   `toml:"pipe"`
		WatchMethod   string `toml:"watch_method"`
		Measurement   string `toml:"source"`
		offset        int64
	}
)

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

	for _, i := range t.C {
		go i.start()
	}
}

func (i *Impl) start() {

	if i.File == config.Cfg.MainCfg.Log {
		l.Error("cannot collect datakit log!")
		return
	}

	if i.Measurement == "" {
		l.Error("invalid measurement")
		return
	}

	var poll bool
	// const defaultWatchMethod = "inotify"
	if i.WatchMethod == "poll" {
		poll = true
	}

	var seek *tail.SeekInfo

	if !i.Pipe && !i.FormBeginning {
		seek = &tail.SeekInfo{
			Whence: 0,
			Offset: i.offset,
		}
		l.Infof("using offset %d", i.offset)
	} else {
		seek = &tail.SeekInfo{
			Whence: 2,
			Offset: 0,
		}
	}

	tailer, err := tail.TailFile(i.File,
		tail.Config{
			ReOpen:    true,
			Follow:    true,
			Location:  seek,
			MustExist: true,
			Poll:      poll,
			Pipe:      i.Pipe,
			Logger:    tail.DiscardingLogger,
		})
	if err != nil {
		l.Errorf("failed to open file, err: %s", err.Error())
		return
	}

	go i.foreachLines(tailer)

	<-datakit.Exit.Wait()

	if !i.Pipe && !i.FormBeginning {
		offset, err := tailer.Tell()
		if err == nil {
			l.Infof("recording offset %d", offset)
		} else {
			l.Errorf("recording offset err:%s", err.Error())
		}
		i.offset = offset
	}

	if err := tailer.Stop(); err != nil {
		l.Errorf("stop err: %s", err.Error())
	} else {
		l.Info("exit")
	}
}

func (i *Impl) foreachLines(tailer *tail.Tail) {

	var fields = make(map[string]interface{})

	for line := range tailer.Lines {
		if line.Err != nil {
			l.Errorf("tailing error: %s", line.Err.Error())
			continue
		}

		text := strings.TrimRight(line.Text, "\r")
		fields["__content"] = text

		pt, err := influxdb.NewPoint(i.Measurement, nil, fields, time.Now())
		if err != nil {
			l.Error(err)
			continue
		}

		data := []byte(pt.String())
		if err := io.Feed(data, io.Logging); err != nil {
			l.Error(err)
		} else {
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}

	l.Info("tailer stop")
	if err := tailer.Err(); err != nil {
		l.Errorf("tailing error: %s", err.Error())
	}
}
