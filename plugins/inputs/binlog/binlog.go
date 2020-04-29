// +build !386

package binlog

import (
	"context"
	"io"
	"log"
	"time"

	blog "github.com/siddontang/go-log/log"

	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Binlog struct {
	NullInt   int               `toml:"null_int"`
	NullFloat float64           `toml:"null_float"`
	Instances []*InstanceConfig `toml:"sources,omitempty"`

	runningBinlogs []*RunningBinloger

	tags map[string]string

	ctx       context.Context
	cancelfun context.CancelFunc

	accumulator telegraf.Accumulator

	logger *models.Logger
}

func (_ *Binlog) SampleConfig() string {
	return binlogConfigSample
}

func (_ *Binlog) Description() string {
	return ""
}

func (b *Binlog) Gather(telegraf.Accumulator) error {
	return nil
}

type adapterLogWriter struct {
	io.Writer
}

func (al *adapterLogWriter) Write(p []byte) (n int, err error) {
	log.Printf("D! [binlog] %s", string(p))
	return len(p), nil
}

//将第三方库的日志打到datakit
func setupLogger() {
	loghandler, _ := blog.NewStreamHandler(&adapterLogWriter{})
	blogger := blog.New(loghandler, 0)
	blog.SetLevel(blog.LevelDebug)
	blog.SetDefaultLogger(blogger)
}

func (b *Binlog) Start(acc telegraf.Accumulator) error {

	if len(b.Instances) == 0 {
		b.logger.Warnf("no config found")
		return nil
	}

	setupLogger()

	b.accumulator = acc

	b.logger.Infof("start")

	for _, inst := range b.Instances {

		inst.applyDefault()

		bl := NewRunningBinloger(inst)
		bl.binlog = b

		b.runningBinlogs = append(b.runningBinlogs, bl)

		go func(rb *RunningBinloger) {

			for {
				if err := rb.run(b.ctx); err != nil && err != context.Canceled {
					b.logger.Errorf("%s", err.Error())
					internal.SleepContext(b.ctx, time.Second*3)
				} else if err == context.Canceled {
					break
				}
			}

			b.logger.Infof("done")

		}(bl)

	}

	return nil
}

func (b *Binlog) Stop() {
	b.cancelfun()
	for _, rb := range b.runningBinlogs {
		rb.stop()
	}
}

func init() {
	inputs.Add("binlog", func() telegraf.Input {
		b := &Binlog{
			logger: &models.Logger{
				Name: `binlog`,
			},
		}
		b.ctx, b.cancelfun = context.WithCancel(context.Background())
		return b
	})
}
