package huaweiyunobject

import (
	"bytes"
	"context"
	"encoding/json"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	moduleLogger *logger.Logger
)

type subModule interface {
	run(*objectAgent)
}

func (_ *objectAgent) SampleConfig() string {
	var buf bytes.Buffer
	buf.WriteString(sampleConfig)
	buf.WriteString(ecsSampleConfig)
	buf.WriteString(elbSampleConfig)
	buf.WriteString(obsSampleConfig)
	buf.WriteString(rdsSampleConfig)
	buf.WriteString(vpcSampleConfig)
	return buf.String()
}

func (_ *objectAgent) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName + "_ecs": ecsPipelineConifg,
		inputName + "_elb": elbPipelineConfig,
		inputName + "_obs": obsPipelineConifg,
		inputName + "_rds": rdsPipelineConfig,
		inputName + "_vpc": vpcPipelineConifg,
	}
	return pipelineMap
}

func (_ *objectAgent) Catalog() string {
	return `huaweiyun`
}

func (ag *objectAgent) Test() (result *inputs.TestResult, err error) {
	return
}

func (ag *objectAgent) Run() {

	moduleLogger = logger.SLogger(inputName)

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
		ag.cancelFun()
	}()

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Hour * 6
	} else if ag.Interval.Duration < time.Hour*1 {
		ag.Interval.Duration = time.Hour * 1
	} else if ag.Interval.Duration > time.Hour*24 {
		ag.Interval.Duration = time.Hour * 24
	}

	if ag.Ecs != nil {
		ag.addModule(ag.Ecs)
	}
	if ag.Elb != nil {
		ag.addModule(ag.Elb)
	}
	if ag.Obs != nil {
		ag.addModule(ag.Obs)
	}
	if ag.Rds != nil {
		ag.addModule(ag.Rds)
	}
	if ag.Vpc != nil {
		ag.addModule(ag.Vpc)
	}

	for _, s := range ag.subModules {
		ag.wg.Add(1)
		go func(s subModule) {
			defer ag.wg.Done()
			s.run(ag)
		}(s)
	}

	ag.wg.Wait()
}

func getPipeline(name string) *pipeline.Pipeline {

	p, err := pipeline.NewPipelineByScriptPath(name)
	if err != nil {
		moduleLogger.Warnf("%s", err)
		return nil
	}

	return p
}

func (ag *objectAgent) parseObject(obj interface{}, name, class, id string, pipeline *pipeline.Pipeline, blacklist, whitelist []string) error {
	if datakit.CheckExcluded(id, blacklist, whitelist) {
		return nil
	}
	data, err := json.Marshal(obj)
	if err != nil {
		moduleLogger.Errorf("json marshal err:%s", err.Error())
		return err
	}

	fields := map[string]interface{}{}
	if pipeline != nil {
		fields, err = pipeline.Run(string(data)).Result()
		if err != nil {
			moduleLogger.Errorf("pipeline run err:%s", err.Error())
			return err
		}
	}

	fields["message"] = string(data)

	tags := map[string]string{
		"name": name,
	}

	if ag.IsDebug() {
		data, err := io.MakeMetric(class, tags, fields, time.Now().UTC())
		if err != nil {
			moduleLogger.Errorf("%s", err)
		} else {
			moduleLogger.Infof("%s", string(data))
		}
		return nil
	} else {
		return io.NamedFeedEx(inputName, io.Object, class, tags, fields, time.Now().UTC())
	}
}

func newAgent(mode string) *objectAgent {
	ag := &objectAgent{}
	ag.mode = mode
	ag.ctx, ag.cancelFun = context.WithCancel(context.Background())
	return ag
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent("")
	})
}
