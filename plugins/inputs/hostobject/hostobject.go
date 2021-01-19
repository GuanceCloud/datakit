package hostobject

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var moduleLogger *logger.Logger

type objCollector struct {
	Name  string            //deprecated
	Class string            //deprecated
	Tags  map[string]string `toml:"tags,omitempty"`        //deprecated
	Desc  string            `toml:"description,omitempty"` //deprecated

	Interval datakit.Duration
	Pipeline string `toml:"pipeline"`

	ctx       context.Context
	cancelFun context.CancelFunc

	mode string

	testResult *inputs.TestResult
	testError  error
}

func (c *objCollector) isTestOnce() bool {
	return c.mode == "test"
}

func (c *objCollector) isDebug() bool {
	return c.mode == "debug"
}

func (_ *objCollector) Catalog() string {
	return InputCat
}

func (_ *objCollector) SampleConfig() string {
	return SampleConfig
}

func (r *objCollector) PipelineConfig() map[string]string {
	return map[string]string{
		InputName: pipelineSample,
	}
}

func (c *objCollector) Test() (*inputs.TestResult, error) {
	c.mode = "test"
	c.testResult = &inputs.TestResult{}
	c.Run()
	return c.testResult, c.testError
}

func (c *objCollector) Run() {

	moduleLogger = logger.SLogger(InputName)

	if c.Interval.Duration == 0 {
		c.Interval.Duration = 5 * time.Minute
	}

	defer func() {
		if e := recover(); e != nil {
			if err := recover(); err != nil {
				buf := make([]byte, 2048)
				n := runtime.Stack(buf, false)
				moduleLogger.Errorf("panic: %s", err)
				moduleLogger.Errorf("%s", string(buf[:n]))
			}
		}
	}()

	go func() {
		<-datakit.Exit.Wait()
		c.cancelFun()
	}()

	var thePipeline *pipeline.Pipeline

	script := c.Pipeline
	if script == "" {
		scriptPath := filepath.Join(datakit.PipelineDir, InputName+".p")
		data, err := ioutil.ReadFile(scriptPath)
		if err == nil {
			script = string(data)
		}
	}

	if script != "" {
		p, err := pipeline.NewPipeline(script)
		if err != nil {
			moduleLogger.Errorf("%s", err)
		} else {
			thePipeline = p
		}
	}

	for {

		select {
		case <-c.ctx.Done():
			return
		default:
		}

		message := getHostObjectMessage()

		messageData, err := json.Marshal(message)
		if err != nil {
			moduleLogger.Errorf("json marshal err:%s", err.Error())
			datakit.SleepContext(c.ctx, c.Interval.Duration)
			continue
		}

		moduleLogger.Debugf("%s", string(messageData))

		fields := map[string]interface{}{
			"message":     string(messageData),
			"os":          message.Host.HostMeta.OS,
			"start_time":  message.Host.HostMeta.BootTime,
			"datakit_ver": git.Version,
			"cpu":         message.Host.cpuPercent,
			"mem":         message.Host.Mem.usedPercent,
			"load":        message.Host.load15,
		}
		if thePipeline != nil {
			if result, err := thePipeline.Run(string(messageData)).Result(); err == nil {
				for k, v := range result {
					fields[k] = v
				}
			} else {
				moduleLogger.Errorf("%s", err)
			}
		}

		tags := map[string]string{
			"name": message.Host.HostMeta.HostName,
		}

		tm := time.Now().UTC()

		if c.isTestOnce() {
			data, err := io.MakeMetric(InputName, tags, fields, tm)
			if err != nil {
				moduleLogger.Errorf("%s", err)
				c.testError = err
			} else {
				c.testResult = &inputs.TestResult{
					Result: data,
					Desc:   "",
				}
				moduleLogger.Debugf("%s\n", string(data))
			}
			return
		} else if c.isDebug() {
			data, _ := io.MakeMetric("HOST", tags, fields, tm)
			fmt.Printf("%s\n", string(data))
		} else {
			io.NamedFeedEx("HOST", io.Object, InputName, tags, fields, tm)
		}

		datakit.SleepContext(c.ctx, c.Interval.Duration)
	}
}

func newInput() *objCollector {
	o := &objCollector{}
	o.ctx, o.cancelFun = context.WithCancel(context.Background())
	return o
}

func init() {
	inputs.Add(InputName, func() inputs.Input {
		return newInput()
	})
}
