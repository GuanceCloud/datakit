package huaweiyunobject

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	rmsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	moduleLogger *logger.Logger
)

func (_ *agent) SampleConfig() string {
	return sampleConfig
}

func (_ *agent) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName + "_ecs": ecsPipelineConifg,
		inputName + "_elb": elbPipelineConfig,
		inputName + "_rds": rdsPipelineConfig,
		inputName + "_vpc": vpcPipelineConifg,
		inputName + "_evs": evsPipelineConifg,
		inputName + "_ims": imsPipelineConifg,
	}
	return pipelineMap
}

func (_ *agent) Catalog() string {
	return `huaweiyun`
}

// TODO
func (*agent) RunPipeline() {
}

func (ag *agent) Run() {

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

	//每分钟最多100个请求
	limit := rate.Every(600 * time.Millisecond)
	ag.limiter = rate.NewLimiter(limit, 1)

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Minute * 15
	}

	ag.run()
}

func getPipeline(name string) *pipeline.Pipeline {

	scriptPath := filepath.Join(datakit.PipelineDir, name)
	if _, e := os.Stat(scriptPath); e != nil && os.IsNotExist(e) {
		return nil
	}
	p, err := pipeline.NewPipelineByScriptPath(name)
	if err != nil {
		moduleLogger.Warnf("%s", err)
		return nil
	}

	return p
}

func (ag *agent) parseObject(res *rmsmodel.ResourceEntity, pipeline *pipeline.Pipeline) error {

	class := fmt.Sprintf("huaweiyun_%s", *res.Provider)

	resName := ""
	resID := ""
	if res.Name != nil {
		resName = *res.Name
	}
	if res.Id != nil {
		resID = *res.Id
	}

	if res.Properties["id"] == nil {
		res.Properties["id"] = resID
	}

	if res.Properties["name"] == nil {
		res.Properties["name"] = resName
	}

	data, err := json.Marshal(res.Properties)
	if err != nil {
		moduleLogger.Errorf("json marshal err:%s", err.Error())
		return err
	}

	fields := map[string]interface{}{}

	if pipeline != nil {
		fields, err = pipeline.Run(string(data)).Result()
		if err != nil {
			moduleLogger.Errorf("pipeline run err:%s", err.Error())
		}
	}

	if fields["id"] == nil {
		fields["id"] = resID //默认加上
	}

	fields["resource_type"] = *res.Type

	if res.RegionId != nil {
		fields["region"] = *res.RegionId
	}
	if res.ProjectId != nil && *res.ProjectId != "" {
		fields["project_id"] = *res.ProjectId
	}
	if res.ProjectName != nil && *res.ProjectName != "" {
		fields["project_name"] = *res.ProjectName
	}

	fields["message"] = string(data)

	tags := map[string]string{
		"name": resName,
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
		return io.NamedFeedEx(inputName, datakit.Object, class, tags, fields, time.Now().UTC())
	}
}

func newAgent(mode string) *agent {
	ag := &agent{}
	ag.mode = mode
	ag.ctx, ag.cancelFun = context.WithCancel(context.Background())
	return ag
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent("")
	})
}
