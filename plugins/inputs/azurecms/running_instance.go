package azurecms

import (
	"context"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/limiter"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

var (
	batchInterval = 5 * time.Minute
	rateLimit     = 10
)

func (r *runningInstance) run(ctx context.Context) error {

	if r.cfg.EndPoint == "" {
		r.cfg.EndPoint = `https://management.chinacloudapi.cn`
	}

	r.metricDefinitionClient = insights.NewMetricDefinitionsClientWithBaseURI(r.cfg.EndPoint, r.cfg.SubscriptionID)
	r.metricClient = insights.NewMetricsClientWithBaseURI(r.cfg.EndPoint, r.cfg.SubscriptionID)

	settings := auth.EnvironmentSettings{
		Values: map[string]string{},
	}
	settings.Values[auth.SubscriptionID] = r.cfg.SubscriptionID
	settings.Values[auth.TenantID] = r.cfg.TenantID
	settings.Values[auth.ClientID] = r.cfg.ClientID
	settings.Values[auth.ClientSecret] = r.cfg.ClientSecret
	settings.Environment = azure.ChinaCloud
	settings.Values[auth.Resource] = settings.Environment.ResourceManagerEndpoint

	r.metricDefinitionClient.Authorizer, _ = settings.GetAuthorizer()
	r.metricClient.Authorizer, _ = settings.GetAuthorizer()

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
				r.logger.Errorf(`fail to get metric "%s.%s", %s`, req.resourceID, req.metricname)
			}
		}
		useage := time.Now().Sub(t)
		internal.SleepContext(ctx, 5*time.Minute-useage)
	}
}

func (r *runningInstance) fetchMetric(ctx context.Context, info *queryListInfo) error {

	now := time.Now().Truncate(time.Minute).Add(-time.Minute)
	if now.Sub(info.lastFetchTime) < info.intervalTime {
		return nil
	}

	start := ""
	if info.lastFetchTime.IsZero() {
		if info.intervalTime < time.Minute*5 {
			start = unixTimeStrISO8601(now.Add(-5 * time.Minute))
		} else {
			start = unixTimeStrISO8601(now.Add(-info.intervalTime))
		}
	} else {
		start = unixTimeStrISO8601(info.lastFetchTime)
	}
	end := unixTimeStrISO8601(now)

	r.logger.Debugf("query param: resourceID=%s, metric=%s, span=%s, interval=%s", info.resourceID, info.metricname, start+"/"+end, info.interval)

	res, err := r.metricClient.List(ctx, info.resourceID, start+"/"+end, &info.interval, info.metricname, info.aggregation, nil, info.orderby, info.filter, info.resultType, info.metricnamespace)

	if err != nil {
		return err
	}

	region := ""
	namespace := ""

	if res.Resourceregion != nil {
		region = *res.Resourceregion
	}

	if res.Namespace != nil {
		namespace = *res.Namespace
	}

	for _, m := range *res.Value {
		metricName := *(*m.Name).Value
		metricUnit := string(m.Unit)

		tms := *m.Timeseries

		r.logger.Debugf("Timeseries(%s) length: %v", metricName, len(tms))

		for _, tm := range tms {

			// if tm.Metadatavalues != nil {
			// 	for _, metaitem := range *tm.Metadatavalues {
			// 		k := *(*metaitem.Name).Value
			// 		v := *(metaitem.Value)
			// 	}
			// }

			if *tm.Data == nil {
				continue
			}

			for _, mv := range *tm.Data {

				tags := map[string]string{
					"Unit": metricUnit,
				}
				if region != "" {
					tags["Resourceregion"] = region
				}
				if namespace != "" {
					tags["Namespace"] = namespace
				}

				fields := map[string]interface{}{}

				if mv.Average != nil {
					fields["Average"] = *mv.Average
				}
				if mv.Count != nil {
					fields["Count"] = *mv.Count
				}
				if mv.Total != nil {
					fields["Total"] = *mv.Total
				}
				if mv.Maximum != nil {
					fields["Maximum"] = *mv.Maximum
				}
				if mv.Minimum != nil {
					fields["Minimum"] = *mv.Minimum
				}

				if len(fields) == 0 {
					continue
				}
				metricTime := time.Now()
				if mv.TimeStamp != nil {
					metricTime = (*mv.TimeStamp).Time
				}
				if r.agent.accumulator != nil {
					r.agent.accumulator.AddFields(metricName, fields, tags, metricTime)
				}
			}
		}
	}

	info.lastFetchTime = now

	return nil
}
