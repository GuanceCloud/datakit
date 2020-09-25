package ucmon

import (
	"context"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func (r *ucInstance) run(ctx context.Context) error {

	uccfg := ucloud.NewConfig()
	credential := auth.NewCredential()
	credential.PrivateKey = r.PrivateKey
	credential.PublicKey = r.PublicKey
	r.ucCli = ucloud.NewClient(&uccfg, &credential)

	r.queryInfos = r.genQueryInfo()

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	limit := rate.Every(50 * time.Millisecond)
	rateLimiter := rate.NewLimiter(limit, 1)

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

			rateLimiter.Wait(ctx)
			if err := r.fetchMetric(ctx, req); err != nil {
				moduleLogger.Errorf(`fail to get metric "%s.%s", %s`, req.resourceID, req.metricname, err)
			}
		}
		useage := time.Now().Sub(t)
		if useage < 5*time.Minute {
			datakit.SleepContext(ctx, 5*time.Minute-useage)
		}
	}
}

func (r *ucInstance) fetchMetric(ctx context.Context, info *queryListInfo) error {

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

	moduleLogger.Debugf("query param: resourceID=%s, metric=%s, range=%v-%v, interval=%s", info.resourceID, info.metricname, begin, end, info.intervalTime)

	req := r.ucCli.NewGenericRequest()
	req.SetAction("GetMetric")
	req.SetRegion(r.Region)
	req.SetZone(r.Zone)
	req.SetProjectId(r.ProjectID)
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
							"Region":       r.Region,
							"Zone":         r.Zone,
							"ProjectID":    r.ProjectID,
							"ResourceID":   info.resourceID,
							"ResourceType": info.resourceType,
						}
						fields := metricItemMap
						metricTime := time.Unix(time.Now().Unix(), 0)
						if tm, ok := fields["Timestamp"].(float64); ok {
							metricTime = time.Unix(int64(tm), 0)
							delete(fields, "Timestamp")
						}

						io.NamedFeedEx(inputName, io.Metric, metricName, tags, fields, metricTime)

					}
				}
			}
		}
	}

	info.lastFetchTime = now

	return nil
}
