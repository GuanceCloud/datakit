package ucmon

import (
	"context"
	"fmt"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `ucloud_monitor`
	moduleLogger *logger.Logger
)

func (_ *ucInstance) SampleConfig() string {
	return sampleConfig
}

func (_ *ucInstance) Catalog() string {
	return "ucloud"
}

func (ag *ucInstance) Run() {

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	uccfg := ucloud.NewConfig()
	credential := auth.NewCredential()
	credential.PrivateKey = ag.PrivateKey
	credential.PublicKey = ag.PublicKey
	ag.ucCli = ucloud.NewClient(&uccfg, &credential)

	ag.queryInfos = ag.genQueryInfo()

	select {
	case <-ag.ctx.Done():
		return
	default:
	}

	limit := rate.Every(50 * time.Millisecond)
	rateLimiter := rate.NewLimiter(limit, 1)

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		t := time.Now()
		for _, req := range ag.queryInfos {

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			rateLimiter.Wait(ag.ctx)
			ag.fetchMetric(ag.ctx, req)
		}

		if ag.isTest() {
			return
		}
		useage := time.Now().Sub(t)
		if useage < 5*time.Minute {
			datakit.SleepContext(ag.ctx, 5*time.Minute-useage)
		}
	}
}

func (ag *ucInstance) fetchMetric(ctx context.Context, info *queryListInfo) error {

	now := time.Now().Truncate(time.Minute)
	if now.Sub(info.lastFetchTime) < info.intervalTime {
		return nil
	}

	var begin, end int64
	if info.lastFetchTime.IsZero() {
		if info.intervalTime < time.Minute*5 {
			begin = now.Add(-5 * time.Minute).Unix()
		} else {
			begin = now.Add(-info.intervalTime).Unix()
		}
	} else {
		begin = info.lastFetchTime.Add(-time.Minute).Unix()
	}
	end = now.Unix()

	req := ag.ucCli.NewGenericRequest()
	req.SetAction("GetMetric")
	req.SetRegion(ag.Region)
	req.SetZone(ag.Zone)
	req.SetProjectId(ag.ProjectID)
	reqPayload := map[string]interface{}{}
	reqPayload["ResourceType"] = info.resourceType
	reqPayload["ResourceId"] = info.resourceID
	reqPayload["MetricName.0"] = info.metricname
	reqPayload["BeginTime"] = begin
	reqPayload["EndTime"] = end
	req.SetPayload(reqPayload)

	resp, err := ag.ucCli.GenericInvoke(req)
	if err == nil && resp.GetRetCode() != 0 {
		err = fmt.Errorf("%s", resp.GetMessage())
	}

	if err != nil {
		moduleLogger.Errorf(`fail to get metric "%s.%s", %s`, info.resourceID, info.metricname, err)
		if ag.isTest() {
			ag.testError = err
		}
		return err
	}

	measurement := "ucloud_monitor_" + info.resourceType

	payload := resp.GetPayload()
	if mapData, ok := payload["DataSets"].(map[string]interface{}); ok {
		for name, datapoints := range mapData {
			if dps, ok := datapoints.([]interface{}); ok {

				moduleLogger.Debugf("%d datapoints, %s.%s(%s), %v - %v", len(dps), info.resourceType, info.metricname, info.resourceID, begin, end)

				for _, dp := range dps {
					if datapoint, ok := dp.(map[string]interface{}); ok {

						tags := map[string]string{
							"Region":       ag.Region,
							"Zone":         ag.Zone,
							"ProjectID":    ag.ProjectID,
							"ResourceID":   info.resourceID,
							"ResourceType": info.resourceType,
						}

						fields := map[string]interface{}{}

						if val, ok := datapoint["Value"]; ok {
							fields[name] = val
						} else {
							moduleLogger.Warnf("Value not found, %s", datapoint)
							continue
						}

						metricTime := time.Now().UTC()
						if tm, ok := fields["Timestamp"].(float64); ok {
							metricTime = time.Unix(int64(tm), 0)
						}

						if ag.isTest() {
							// pass
						} else if ag.isDebug() {
							data, _ := io.MakeMetric(measurement, tags, fields, metricTime)
							fmt.Printf("-----%s\n", data)
						} else {
							io.NamedFeedEx(inputName, datakit.Metric, measurement, tags, fields, metricTime)
						}

					}
				}
			}
		}
	}

	info.lastFetchTime = now

	return nil
}

func newAgent() *ucInstance {
	ac := &ucInstance{}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent()
	})
}
