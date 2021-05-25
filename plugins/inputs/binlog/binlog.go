// +build !386,!arm

package binlog

import (
	"context"
	"io"
	"sync"
	"time"

	blog "github.com/siddontang/go-log/log"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	moduleLogger *logger.Logger
	inputName    = `binlog`
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

	mode string

	testError error
}

func (_ *Binlog) Catalog() string {
	return "db"
}

func (_ *Binlog) SampleConfig() string {
	return binlogConfigSample
}

type adapterLogWriter struct {
	io.Writer
}

func (al *adapterLogWriter) Write(p []byte) (n int, err error) {
	moduleLogger.Debugf("%s", string(p))
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

	moduleLogger = logger.SLogger(inputName)

	if len(b.Instances) == 0 {
		moduleLogger.Warnf("no config found")
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

				select {
				case <-datakit.Exit.Wait():
					return
				default:
				}

				if err := bl.run(b.ctx); err != nil && err != context.Canceled {
					moduleLogger.Errorf("%s", err.Error())
					datakit.SleepContext(b.ctx, time.Second*3)
				} else if err == context.Canceled {
					break
				}
			}

		}(inst)
	}

	b.wg.Wait()
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		b := &Binlog{}
		b.ctx, b.cancelfun = context.WithCancel(context.Background())
		return b
	})
}
