package tencentobject

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `tencentobject`
	moduleLogger *logger.Logger
)

type subModule interface {
	run(*objectAgent)
}

func (_ *objectAgent) SampleConfig() string {
	var buf bytes.Buffer
	buf.WriteString(sampleConfig)
	buf.WriteString(cvmSampleConfig)
	buf.WriteString(cosSampleConfig)
	buf.WriteString(clbSampleConfig)
	buf.WriteString(cdbSampleConfig)
	return buf.String()
}

func (_ *objectAgent) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"tencent_cdb": cdbPipelineConfig,
		"tencent_cos": cosPipelineConfig,
		"tencent_clb": clbPipelineConfig,
		"tencent_cvm": cvmPipelineConfig,
	}
	return pipelineMap
}

func (_ *objectAgent) Catalog() string {
	return `tencentcloud`
}

func (ag *objectAgent) getCredential() *common.Credential {
	credential := common.NewCredential(
		ag.AccessKeyID,
		ag.AccessKeySecret,
	)
	return credential
}

func newPipeline(pipelinePath string) (*pipeline.Pipeline, error) {
	scriptPath := filepath.Join(datakit.PipelineDir, pipelinePath)
	data, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}
	p, err := pipeline.NewPipeline(string(data))
	return p, err
}

func (ag *objectAgent) parseObject(obj interface{}, class, id string, pipeline *pipeline.Pipeline, blacklist, whitelist []string, tags map[string]string) {
	if datakit.CheckExcluded(id, blacklist, whitelist) {
		return
	}
	data, err := json.Marshal(obj)
	if err != nil {
		moduleLogger.Errorf("[error] json marshal err:%s", err.Error())
		return
	}
	fields, err := pipeline.Run(string(data)).Result()
	if err != nil {
		moduleLogger.Errorf("[error] pipeline run err:%s", err.Error())
		return
	}
	fields["message"] = string(data)

	io.NamedFeedEx(inputName, datakit.Object, class, tags, fields, time.Now().UTC())
}

// TODO
func (*objectAgent) RunPipeline() {
}

func (ag *objectAgent) Run() {

	moduleLogger = logger.SLogger(inputName)

	ag.ctx, ag.cancelFun = context.WithCancel(context.Background())

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Minute * 5
	}

	if ag.Cvm != nil {
		ag.addModule(ag.Cvm)
	}
	if ag.Cos != nil {
		ag.addModule(ag.Cos)
	}
	if ag.Clb != nil {
		ag.addModule(ag.Clb)
	}
	if ag.Cdb != nil {
		ag.addModule(ag.Cdb)
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
