package huaweiyunces

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"
	"unicode"

	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	moduleLogger *logger.Logger
	inputName    = "huaweiyunces"
)

func (*agent) SampleConfig() string {
	return sampleConfig
}

func (*agent) Catalog() string {
	return `huaweiyun`
}

func (ag *agent) Test() (*inputs.TestResult, error) {
	ag.mode = "test"
	ag.testResult = &inputs.TestResult{}
	ag.Run()
	return ag.testResult, ag.testError
}

func (ag *agent) Run() {

	moduleLogger = logger.SLogger(inputName)

	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			moduleLogger.Errorf("panic: %s", err)
			moduleLogger.Errorf("%s", string(buf[:n]))
		}
	}()

	if ag.Delay.Duration == 0 {
		ag.Delay.Duration = time.Minute * 1
	}

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Minute * 5
	}

	ag.genHWClient()

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	//每分钟最多100个请求
	limit := rate.Every(600 * time.Millisecond)
	ag.limiter = rate.NewLimiter(limit, 1)

	if err := ag.genReqs(ag.ctx); err != nil {
		if ag.isTestOnce() {
			ag.testError = err
		}
		return
	}

	if len(ag.reqs) == 0 {
		moduleLogger.Warnf("no metric found")
		if ag.isTestOnce() {
			ag.testError = fmt.Errorf("no metric found")
		}
		return
	}
	moduleLogger.Debugf("%d reqs", len(ag.reqs))

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		for _, req := range ag.reqs {

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			resp, err := ag.showMetricData(req)
			if err != nil {
				if ag.isTestOnce() {
					ag.testError = err
					return
				}
				continue
			}

			if resp == nil || resp.Datapoints == nil {
				continue
			}

			metricSetName := req.metricSetName
			if metricSetName == "" {
				metricSetName = formatMeasurement(req.namespace)
			}

			for _, dp := range *resp.Datapoints {

				select {
				case <-ag.ctx.Done():
					return
				default:
				}

				tags := map[string]string{}
				extendTags(tags, req.tags, false)
				extendTags(tags, ag.Tags, false)
				for _, k := range req.dimensoions {
					tags[k.Name] = k.Value
				}
				tags["unit"] = *dp.Unit

				fields := map[string]interface{}{}

				var val float64
				switch req.filter {
				case "max":
					if dp.Max != nil {
						val = *dp.Max
					}
				case "min":
					if dp.Min != nil {
						val = *dp.Min
					}
				case "sum":
					if dp.Sum != nil {
						val = *dp.Sum
					}
				case "variance":
					if dp.Variance != nil {
						val = *dp.Variance
					}
				default:
					if dp.Average != nil {
						val = *dp.Average
					}
				}

				fields[fmt.Sprintf("%s_%s", req.metricname, req.filter)] = val

				tm := time.Unix(dp.Timestamp/1000, 0)

				if len(fields) == 0 {
					moduleLogger.Warnf("skip %s.%s datapoint for no fields, %s", req.namespace, req.metricname, dp.String())
					continue
				}

				if ag.isTestOnce() {
					data, _ := io.MakeMetric(metricSetName, tags, fields, tm)
					ag.testResult.Result = append(ag.testResult.Result, data...)
				} else if ag.isDebug() {
					data, _ := io.MakeMetric(metricSetName, tags, fields, tm)
					fmt.Printf("%s\n", string(data))

				} else {
					io.NamedFeedEx(inputName, io.Metric, metricSetName, tags, fields, tm)
				}
			}
		}

		if ag.isTestOnce() {
			break
		}

		datakit.SleepContext(ag.ctx, time.Second*3)
	}
}

func extendTags(to map[string]string, from map[string]string, override bool) {
	if from == nil {
		return
	}
	for k, v := range from {
		if !override {
			if _, exist := to[k]; exist {
				continue
			}
		}
		to[k] = v
	}
}

func formatMeasurement(project string) string {
	project = strings.Replace(project, "/", "_", -1)
	project = snakeCase(project)
	return fmt.Sprintf("%s_%s", inputName, project)
}

func snakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	s := strings.Replace(string(out), "__", "_", -1)

	return s
}

func (ag *agent) genReqs(ctx context.Context) error {

	ag.ecsInstanceIDs = ag.listServersDetails()
	moduleLogger.Debugf("%d ecs instances", len(ag.ecsInstanceIDs))

	//生成所有请求
	for _, proj := range ag.Namespace {

		if err := proj.checkProperties(); err != nil {
			return err
		}

		for _, metricName := range proj.MetricNames {

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			req := proj.genMetricReq(metricName)
			if req.interval == 0 {
				req.interval = ag.Interval.Duration
			}

			adds := proj.applyProperty(req, ag.ecsInstanceIDs)

			ag.reqs = append(ag.reqs, req)
			if len(adds) > 0 {
				ag.reqs = append(ag.reqs, adds...)
			}
		}
	}

	return nil
}

func newAgent(mode string) *agent {
	ac := &agent{}
	ac.mode = mode
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent("")
	})
}
