package ucmon

import (
	"context"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/limiter"
)

var (
	batchInterval = 5 * time.Minute
	rateLimit     = 10
)

func (r *runningInstance) run(ctx context.Context) error {

	uccfg := ucloud.NewConfig()
	credential := auth.NewCredential()
	credential.PrivateKey = r.cfg.PrivateKey
	credential.PublicKey = r.cfg.PublicKey
	r.ucCli = ucloud.NewClient(&uccfg, &credential)

	r.queryInfos = r.cfg.genQueryInfo()

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	lmtr := limiter.NewRateLimiter(rateLimit, time.Second)
	defer lmtr.Stop()

	for {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		t := time.Now()
		for _, req := range r.queryInfos {

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			<-lmtr.C
			if err := r.fetchMetric(ctx, req); err != nil {
				r.logger.Errorf(`fail to get metric "%s.%s", %s`, req.resourceID, req.metricname, err)
			}
		}
		useage := time.Now().Sub(t)
		if useage < 5*time.Minute {
			internal.SleepContext(ctx, 5*time.Minute-useage)
		}
	}
}

func (r *runningInstance) fetchMetric(ctx context.Context, info *queryListInfo) error {

	now := time.Now().Truncate(time.Minute).Add(-time.Minute)
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
		begin = info.lastFetchTime.Unix()
	}
	end = now.Unix()

	r.logger.Debugf("query param: resourceID=%s, metric=%s, range=%v-%v, interval=%s", info.resourceID, info.metricname, begin, end, info.intervalTime)

	req := r.ucCli.NewGenericRequest()
	req.SetAction("GetMetric")
	req.SetRegion(r.cfg.Region)
	req.SetZone(r.cfg.Zone)
	req.SetProjectId(r.cfg.ProjectID)
	reqPayload := map[string]interface{}{}
	reqPayload["ResourceType"] = info.resourceType
	reqPayload["ResourceId"] = info.resourceID
	reqPayload["MetricName"] = info.metricname
	reqPayload["BeginTime"] = begin
	reqPayload["EndTime"] = end
	req.SetPayload(reqPayload)

	resp, err := r.ucCli.GenericInvoke(req)
	if err != nil {
		return err
	}

	payload := resp.GetPayload()
	if mapData, ok := payload["DataSets"].(map[string]interface{}); ok {
		for _, v := range mapData {
			if series, ok := v.([]interface{}); ok {
				for _, metricItem := range series {
					if metricItemMap, ok := metricItem.(map[string]interface{}); ok {

						metricName := "ucmon_" + info.metricname
						tags := map[string]string{
							"Region":       r.cfg.Region,
							"Zone":         r.cfg.Zone,
							"ProjectID":    r.cfg.ProjectID,
							"ResourceID":   info.resourceID,
							"ResourceType": info.resourceType,
						}
						fields := metricItemMap
						metricTime := time.Unix(time.Now().Unix(), 0)
						if tm, ok := fields["Timestamp"].(float64); ok {
							metricTime = time.Unix(int64(tm), 0)
							delete(fields, "Timestamp")
						}

						if r.agent.accumulator != nil && len(fields) > 0 {
							r.agent.accumulator.AddFields(metricName, fields, tags, metricTime)
						}
					}
				}
			}
		}
	}

	info.lastFetchTime = now

	return nil
}
