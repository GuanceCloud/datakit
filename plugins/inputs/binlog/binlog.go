// +build !386,!arm

package binlog

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	blog "github.com/siddontang/go-log/log"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Binlog struct {
	NullInt   int               `toml:"null_int"`
	NullFloat float64           `toml:"null_float"`
	Instances []*InstanceConfig `toml:"sources,omitempty"`

	runningBinlogs []*RunningBinloger

	wg sync.WaitGroup

	tags map[string]string

	ctx       context.Context
	cancelfun context.CancelFunc

	logger *models.Logger
}

func (_ *Binlog) Catalog() string {
	return "mysql"
}

func (_ *Binlog) SampleConfig() string {
	return binlogConfigSample
}

// func (_ *Binlog) Description() string {
// 	return ""
// }

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

func (b *Binlog) Run() {

	if len(b.Instances) == 0 {
		b.logger.Warnf("no config found")
		return
	}

	setupLogger()

	go func() {
		<-datakit.Exit.Wait()
		b.cancelfun()
		for _, rb := range b.runningBinlogs {
			rb.stop()
		}
	}()

	for _, inst := range b.Instances {

		b.wg.Add(1)

		go func(inst *InstanceConfig) {
			defer b.wg.Done()

			inst.applyDefault()

			bl := NewRunningBinloger(inst)
			bl.binlog = b

			b.runningBinlogs = append(b.runningBinlogs, bl)

			for {
				if err := bl.run(b.ctx); err != nil && err != context.Canceled {
					b.logger.Errorf("%s", err.Error())
					internal.SleepContext(b.ctx, time.Second*3)
				} else if err == context.Canceled {
					break
				}
			}

			b.logger.Infof("done")

		}(inst)
	}

	b.wg.Wait()
}

func init() {
	inputs.Add("binlog", func() inputs.Input {
		b := &Binlog{
			logger: &models.Logger{
				Name: `binlog`,
			},
		}
		b.ctx, b.cancelfun = context.WithCancel(context.Background())
		return b
	})
}
