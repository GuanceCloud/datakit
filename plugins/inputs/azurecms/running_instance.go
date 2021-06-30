package azurecms

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"golang.org/x/time/rate"
)

func (r *azureInstance) run(ctx context.Context) error {

	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			moduleLogger.Errorf("panic: %s", err)
			moduleLogger.Errorf("%s", string(buf[:n]))
		}
	}()

	if r.EndPoint == "" {
		r.EndPoint = `https://management.chinacloudapi.cn/`
	}

	ev := azure.PublicCloud

	endPoints := []azure.Environment{
		azure.ChinaCloud,
		azure.USGovernmentCloud,
		azure.GermanCloud,
	}

	for _, e := range endPoints {
		if e.ResourceManagerEndpoint == r.EndPoint {
			ev = e
		}
	}

	settings := auth.EnvironmentSettings{
		Values: map[string]string{},
	}
	settings.Values[auth.SubscriptionID] = r.SubscriptionID
	settings.Values[auth.TenantID] = r.TenantID
	settings.Values[auth.ClientID] = r.ClientID
	settings.Values[auth.ClientSecret] = r.ClientSecret
	settings.Environment = ev
	settings.Values[auth.Resource] = settings.Environment.ResourceManagerEndpoint

	moduleLogger.Debugf("ev: %s", ev.ResourceManagerEndpoint)

	r.metricDefinitionClient = insights.NewMetricDefinitionsClientWithBaseURI(ev.ResourceManagerEndpoint, r.SubscriptionID)
	r.metricClient = insights.NewMetricsClientWithBaseURI(ev.ResourceManagerEndpoint, r.SubscriptionID)

	r.metricDefinitionClient.Authorizer, _ = settings.GetAuthorizer()
	r.metricClient.Authorizer, _ = settings.GetAuthorizer()

	r.queryInfos = r.genQueryInfo()

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	limit := rate.Every(50 * time.Millisecond)
	r.rateLimiter = rate.NewLimiter(limit, 1)

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

			r.rateLimiter.Wait(ctx)
			if err := r.fetchMetric(ctx, req); err != nil {
				moduleLogger.Errorf("fail to get metric '%s.%s', %s", req.resourceID, req.metricname, err)
				if r.isTest() {
					r.testError = err
					return err
				}
			}
		}

		if r.isTest() {
			return nil
		}
		useage := time.Now().Sub(t)
		datakit.SleepContext(ctx, 5*time.Minute-useage)
	}
}

func (r *azureInstance) fetchMetric(ctx context.Context, info *queryListInfo) error {

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

	moduleLogger.Debugf("query param: resourceID=%s, metric=%s, span=%s, interval=%s", info.resourceID, info.metricname, start+"/"+end, info.interval)

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

		moduleLogger.Debugf("Timeseries(%s) length: %v", metricName, len(tms))

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

				if r.isTest() {
					// pass
				} else if r.isDebug() {
					data, _ := io.MakeMetric(metricName, tags, fields, metricTime)
					fmt.Printf("%s\n", string(data))
				} else {
					io.NamedFeedEx(inputName, datakit.Metric, metricName, tags, fields, metricTime)
				}

			}
		}
	}

	info.lastFetchTime = now

	return nil
}
