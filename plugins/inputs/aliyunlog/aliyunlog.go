package aliyunlog

import (
	"encoding/json"
	"fmt"
	sysio "io"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	consumerLibrary "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/aliyunlog/consumer"
)

var (
	inputName    = `aliyunlog`
	moduleLogger *logger.Logger
)

type runningProject struct {
	inst *ConsumerInstance
	cfg  *LogProject

	wg sync.WaitGroup
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
}

func (_ *ConsumerInstance) Catalog() string {
	return "aliyun"
}

func (_ *ConsumerInstance) SampleConfig() string {
	return aliyunlogConfigSample
}

func (al *ConsumerInstance) Run() {

	moduleLogger = logger.SLogger(inputName)

	for _, c := range al.Projects {
		al.wg.Add(1)

		go func(lp *LogProject) {
			defer al.wg.Done()

			p := &runningProject{
				cfg:  lp,
				inst: al,
			}
			p.run()
		}(c)
	}

	al.wg.Wait()

}

func (r *runningProject) run() {

	for _, c := range r.cfg.Stores {
		r.wg.Add(1)

		if c.ConsumerGroupName == "" {
			c.ConsumerGroupName = "datakit-" + config.Cfg.UUID
		}

		if c.ConsumerName == "" {
			c.ConsumerName = "datakit-" + config.Cfg.UUID
		}

		go func(ls *LogStoreCfg) {
			defer r.wg.Done()

			s := &runningStore{
				cfg:        ls,
				proj:       r,
				tagsInfo:   map[string]*tagReplace{},
				fieldsInfo: map[string]string{},
			}
			s.metricName = ls.MetricName
			if s.metricName == "" {
				s.metricName = `aliyunlog_` + ls.Name
			}

			s.run()
		}(c)

	}

	r.wg.Wait()
}

type adapterLogWriter struct {
	sysio.Writer
}

func (al *adapterLogWriter) Write(p []byte) (n int, err error) {
	moduleLogger.Debugf("%s", string(p))
	return len(p), nil
}

func sdkLogger() log.Logger {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(&adapterLogWriter{}))
	switch config.Cfg.LogLevel {
	case "debug":
		logger = level.NewFilter(logger, level.AllowDebug())
	default:
		logger = level.NewFilter(logger, level.AllowInfo())
	}
	logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	return logger
}

func (r *runningStore) run() error {

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
			moduleLogger.Warnf("invalid field type specification")
			continue
		}
		fieldType := parts[0]
		fieldNames := strings.Split(parts[1], ",")
		for _, f := range fieldNames {
			r.fieldsInfo[f] = fieldType
		}
	}

	option := consumerLibrary.LogHubConfig{
		Endpoint:          r.proj.inst.Endpoint,
		AccessKeyID:       r.proj.inst.AccessKeyID,
		AccessKeySecret:   r.proj.inst.AccessKeySecret,
		Project:           r.proj.cfg.Name,
		Logstore:          r.cfg.Name,
		ConsumerGroupName: r.cfg.ConsumerGroupName,
		ConsumerName:      r.cfg.ConsumerName,
		CursorPosition:    consumerLibrary.END_CURSOR,
	}

	consumerWorker := consumerLibrary.InitConsumerWorker(option, sdkLogger(), r.logProcess)
	consumerWorker.Start()

	<-datakit.Exit.Wait()
	consumerWorker.StopAndWait()

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
			tags["*source"] = r.cfg.Name

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

			// if lg.GetSource() != "" {
			// 	tagInfo := r.checkAsTag("__source__")
			// 	if tagInfo != nil {
			// 		tags[tagInfo.name] = lg.GetSource()
			// 		if !tagInfo.onlyAsTag {
			// 			fields["__source__"] = lg.GetSource()
			// 		}
			// 	} else {
			// 		fields["__source__"] = lg.GetSource()
			// 	}
			// }

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
								if strval != "-" {
									moduleLogger.Debugf("you specify '%s' as int, but fail to convert '%s' to int", k, strval)
								}
							}
						} else {
							fields[k] = nval
						}
					case "float":
						fval, err := strconv.ParseFloat(strval, 64)
						if err != nil {
							if strval != "-" {
								moduleLogger.Debugf("you specify '%s' as float, but fail to convert '%s' to float", k, strval)
							}
						} else {
							fields[k] = fval
						}
					}
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
				fields["message"] = string(contentStr)
			} else {
				moduleLogger.Debugf("fail to marshal content, %s", err)
			}

			tm := time.Unix(int64(l.GetTime()), 0)

			if r.proj.inst.mode == "debug" {
				mdata, _ := io.MakeMetric(r.metricName, tags, fields, tm)
				fmt.Printf("%s\n", string(mdata))
			} else {
				io.NamedFeedEx(inputName, datakit.Logging, r.metricName, tags, fields, tm)
			}

		}
	}
	return ""
}

func NewAgent() *ConsumerInstance {
	ac := &ConsumerInstance{}
	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return NewAgent()
	})
}
