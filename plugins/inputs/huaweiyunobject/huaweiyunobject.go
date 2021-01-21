package huaweiyunobject

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `huaweiyunobject`
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
	buf.WriteString(mysqlSampleConfig)
	return buf.String()
}

func (_ *objectAgent) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"huaweiyun_ecs":   ecsPipelineConifg,
		"huaweiyun_elb":   elbPipelineConfig,
		"huaweiyun_obs":   obsPipelineConifg,
		"huaweiyun_mysql": mysqlPipelineConfig,
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

	ag.ctx, ag.cancelFun = context.WithCancel(context.Background())

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
	if ag.Mysql != nil {
		ag.addModule(ag.Mysql)
	}

	for _, s := range ag.subModules {
		ag.wg.Add(1)
		go func(s subModule) {
			defer ag.wg.Done()
			s.run(ag)
		}(s)
	}

	ag.wg.Wait()

	moduleLogger.Debugf("done")
}

func newAgent() *objectAgent {
	ag := &objectAgent{}
	return ag
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent()
	})
}

func (ag *objectAgent) parseObject(obj interface{}, name, class, id string, pipeline *pipeline.Pipeline, blacklist, whitelist []string) error {
	if datakit.CheckExcluded(id, blacklist, whitelist) {
		return nil
	}
	data, err := json.Marshal(obj)
	if err != nil {
		moduleLogger.Errorf("[error] json marshal err:%s", err.Error())
		return err
	}

	fields := map[string]interface{}{}
	if pipeline != nil {
		fields, err = pipeline.Run(string(data)).Result()
		if err != nil {
			moduleLogger.Errorf("[error] pipeline run err:%s", err.Error())
			return err
		}
	}

	fields["message"] = string(data)

	tags := map[string]string{
		"name": name,
	}

	return io.NamedFeedEx(inputName, io.Object, class, tags, fields, time.Now().UTC())
}
