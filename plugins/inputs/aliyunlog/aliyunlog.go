package aliyunlog

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	consumerLibrary "github.com/aliyun/aliyun-log-go-sdk/consumer"
)

type AliyunLog struct {
	Consumer []*ConsumerInstance

	runningInstances []*runningInstance

	ctx       context.Context
	cancelFun context.CancelFunc

	accumulator telegraf.Accumulator

	logger *models.Logger
}

type runningInstance struct {
	cfg *ConsumerInstance

	agent *AliyunLog

	logger *models.Logger

	runningProjects []*runningProject
}

type runningProject struct {
	inst *runningInstance
	cfg  *LogProject

	logger *models.Logger

	runningStores []*runningStore
}

type tagReplace struct {
	name      string
	onlyAsTag bool
}

type runningStore struct {
	proj       *runningProject
	cfg        *LogStoreCfg
	metricName string

	fieldsInfo map[string]string

	tagsInfo map[string]*tagReplace

	logger *models.Logger
}

func (_ *AliyunLog) Catalog() string {
	return "aliyun"
}

func (_ *AliyunLog) SampleConfig() string {
	return aliyunlogConfigSample
}

func (_ *AliyunLog) Description() string {
	return "Collect logs from aliyun SLS"
}

func (_ *AliyunLog) Gather(telegraf.Accumulator) error {
	return nil
}

func (al *AliyunLog) Start(acc telegraf.Accumulator) error {

	al.logger = &models.Logger{
		Name: `aliyunlog`,
	}

	if len(al.Consumer) == 0 {
		al.logger.Warnf("no configuration found")
		return nil
	}

	al.logger.Info("starting...")

	al.accumulator = acc

	for _, instCfg := range al.Consumer {
		r := &runningInstance{
			cfg:    instCfg,
			agent:  al,
			logger: al.logger,
		}
		al.runningInstances = append(al.runningInstances, r)

		go func(ctx context.Context) {
			if err := r.run(al.ctx); err != nil && err != context.Canceled {
				al.logger.Errorf("%s", err)
			}
		}(al.ctx)
	}

	return nil
}

func (al *AliyunLog) Stop() {
	al.cancelFun()
}

func (r *runningInstance) run(ctx context.Context) error {

	for _, c := range r.cfg.Projects {

		p := &runningProject{
			cfg:    c,
			inst:   r,
			logger: r.logger,
		}
		r.runningProjects = append(r.runningProjects, p)

		go func(ctx context.Context) {
			if err := p.run(ctx); err != nil && err != context.Canceled {
				r.logger.Errorf("%s", err)
			}
		}(ctx)
	}

	return nil
}

func (r *runningProject) run(ctx context.Context) error {

	for _, c := range r.cfg.Stores {

		s := &runningStore{
			cfg:        c,
			proj:       r,
			logger:     r.logger,
			tagsInfo:   map[string]*tagReplace{},
			fieldsInfo: map[string]string{},
		}
		s.metricName = c.MetricName
		if s.metricName == "" {
			s.metricName = `aliyunlog_` + c.Name
		}
		r.runningStores = append(r.runningStores, s)

		go func(ctx context.Context) {
			if err := s.run(ctx); err != nil && err != context.Canceled {
				r.logger.Errorf("%s", err)
			}
		}(ctx)
	}

	return nil
}

func (r *runningStore) run(ctx context.Context) error {

	for _, titem := range r.cfg.Tags {
		parts := strings.Split(titem, ":")
		switch len(parts) {
		case 1:
			r.tagsInfo[parts[0]] = &tagReplace{
				name:      parts[0],
				onlyAsTag: true,
			}
		case 2:
			r.tagsInfo[parts[0]] = &tagReplace{
				name:      parts[0],
				onlyAsTag: parts[1] == "*",
			}
		case 3:
			aliasName := parts[2]
			if aliasName == "" {
				aliasName = parts[0]
			}
			r.tagsInfo[parts[0]] = &tagReplace{
				name:      aliasName,
				onlyAsTag: parts[1] != "*",
			}

		}
	}

	for _, fitem := range r.cfg.Fields {
		parts := strings.Split(fitem, ":")
		if len(parts) != 2 {
			r.logger.Warnf("invalid field type specification")
			continue
		}
		fieldType := parts[0]
		fieldNames := strings.Split(parts[1], ",")
		for _, f := range fieldNames {
			r.fieldsInfo[f] = fieldType
		}
	}

	option := consumerLibrary.LogHubConfig{
		Endpoint:          r.proj.inst.cfg.Endpoint,
		AccessKeyID:       r.proj.inst.cfg.AccessKeyID,
		AccessKeySecret:   r.proj.inst.cfg.AccessKeySecret,
		Project:           r.proj.cfg.Name,
		Logstore:          r.cfg.Name,
		ConsumerGroupName: r.cfg.ConsumerGroupName,
		ConsumerName:      r.cfg.ConsumerName,
		// This options is used for initialization, will be ignored once consumer group is created and each shard has been started to be consumed.
		// Could be "begin", "end", "specific time format in time stamp", it's log receiving time.
		CursorPosition: consumerLibrary.BEGIN_CURSOR,
	}

	consumerWorker := consumerLibrary.InitConsumerWorker(option, r.logProcess)
	consumerWorker.Start()

	<-ctx.Done()
	consumerWorker.StopAndWait()

	r.logger.Infof("%s done", r.cfg.Name)

	return nil

}

func (r *runningStore) checkAsTag(key string) *tagReplace {

	if v, ok := r.tagsInfo[key]; ok {
		return v
	}
	return nil
}

func (r *runningStore) checkFieldType(field string) string {
	if ftype, ok := r.fieldsInfo[field]; ok {
		return ftype
	}
	return "string"
}

func (r *runningStore) logProcess(shardId int, logGroupList *sls.LogGroupList) string {

	//r.logger.Debugf("shardId:%d, grouplist:%s", shardId, logGroupList.String())
	for _, lg := range logGroupList.LogGroups {

		for _, l := range lg.GetLogs() {

			fields := map[string]interface{}{}

			tags := map[string]string{}
			tags["store"] = r.cfg.Name
			tags["project"] = r.proj.cfg.Name
			tags["__topic__"] = lg.GetTopic()

			for _, lt := range lg.GetLogTags() {
				k := lt.GetKey()
				if k == "" || lt.GetValue() == "" {
					continue
				}
				tagInfo := r.checkAsTag(k)
				if tagInfo != nil {
					tags[tagInfo.name] = lt.GetValue()
					if !tagInfo.onlyAsTag {
						fields[k] = lt.GetValue()
					}
				} else {
					fields[k] = lt.GetValue()
				}
			}

			if lg.GetSource() != "" {
				tagInfo := r.checkAsTag("__source__")
				if tagInfo != nil {
					tags[tagInfo.name] = lg.GetSource()
					if !tagInfo.onlyAsTag {
						fields["__source__"] = lg.GetSource()
					}
				} else {
					fields["__source__"] = lg.GetSource()
				}
			}

			// if lg.GetCategory() != "" {
			// 	tags["__category__"] = lg.GetCategory()
			// }

			contentMap := map[string]string{}

			for _, lc := range l.Contents {
				k := lc.GetKey()
				if k == "" {
					continue
				}

				contentMap[k] = lc.GetValue()

				tagInfo := r.checkAsTag(k)
				if tagInfo != nil {
					tags[tagInfo.name] = lc.GetValue()
				}

				if tagInfo != nil && tagInfo.onlyAsTag {
					continue
				}

				strval := lc.GetValue()
				fieldType := r.checkFieldType(k)
				if fieldType != "string" {
					switch fieldType {
					case "int":
						nval, err := strconv.ParseInt(strval, 10, 64)
						if err != nil {
							if fval, err := strconv.ParseFloat(strval, 64); err == nil {
								nval = int64(math.Floor(fval))
							} else {
								//r.logger.Warnf("you specify '%s' as int, but fail to convert '%s' to int", k, strval)
							}
						} else {
							fields[k] = nval
						}
					case "float":
						fval, err := strconv.ParseFloat(strval, 64)
						if err != nil {
							//r.logger.Warnf("you specify '%s' as float, but fail to convert '%s' to float", k, strval)
						} else {
							fields[k] = fval
						}
					}
				} else {
					fields[k] = strval
				}
			}

			uid, _ := uuid.NewV4()
			uuidKey := "DF_LOG_ID"
			tagInfo := r.checkAsTag(uuidKey)
			if tagInfo != nil {
				tags[tagInfo.name] = uid.String()
				if !tagInfo.onlyAsTag {
					fields[uuidKey] = uid.String()
				}
			} else {
				fields[uuidKey] = uid.String()
			}

			contentStr, err := json.Marshal(&contentMap)
			if err == nil {
				fields["__content"] = string(contentStr)
			} else {
				r.logger.Warnf("fail to marshal content, %s", err)
			}

			tm := time.Unix(int64(l.GetTime()), 0)
			m, err := metric.New(r.metricName, tags, fields, tm)
			if err == nil {
				if r.proj.inst.agent.accumulator != nil {
					r.proj.inst.agent.accumulator.AddMetric(m)
				} else {
					fmt.Printf("%s", internal.Metric2InfluxLine(m))
				}
			} else {
				r.logger.Warnf("fail to generate metric, %s", err)
			}
		}
	}
	return ""
}

func NewAgent() *AliyunLog {
	ac := &AliyunLog{}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	return ac
}

func init() {
	inputs.Add("aliyunlog", func() inputs.Input {
		return NewAgent()
	})
}
