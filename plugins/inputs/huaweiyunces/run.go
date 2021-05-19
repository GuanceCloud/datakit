package huaweiyunces

import (
	"fmt"
	"strconv"
	"time"

	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func (ag *agent) run() {

	ag.parseConfig()

	ag.reloadCloud()
	ag.reloadCloudTime = time.Now()

	moduleLogger.Debugf("start getting metris....")

	for {
		select {
		case <-ag.ctx.Done():
			return
		default:
		}
		for _, cli := range ag.clients {
			cli.loop(ag)
		}
		datakit.SleepContext(ag.ctx, time.Second*3)

		if time.Now().Sub(ag.reloadCloudTime) > time.Minute*15 {
			ag.reloadCloud()
			ag.reloadCloudTime = time.Now()
		}
	}
}

func (ag *agent) reloadCloud() {
	//获取所有项目
	projects, err := ag.keystoneListAuthProjects()
	if err != nil {
		return
	}

	ag.clients = []*cesCli{}

	allProjects := []string{}
	for _, p := range projects {
		moduleLogger.Debugf("%s(%s)", p.Name, p.Id)
		if ag.checkProjectIgnore(p.Id) {
			continue
		}
		regID := p.Name
		if p.ParentId != p.DomainId {
			if pp, ok := projects[p.ParentId]; ok {
				regID = pp.Name
			} else {
				continue
			}
		}
		cli, endpoint := ag.genCesClient(p.Id, regID)
		if cli != nil {
			moduleLogger.Debugf("client: %s(%s)", p.Name, endpoint)
			allProjects = append(allProjects, p.Name)
			ag.clients = append(ag.clients, &cesCli{
				proj: p,
				cli:  cli,
			})
		}
	}

	if len(ag.clients) == 0 {
		moduleLogger.Warnf("no project found")
		return
	}

	moduleLogger.Debugf("generating all metrics...")

	for _, cli := range ag.clients {
		select {
		case <-ag.ctx.Done():
			return
		default:
		}
		cli.genRequests(ag)
	}

}

func (cli *cesCli) loop(ag *agent) {
	//按命名空间
	for _, namespaceReqs := range cli.requests {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		for _, req := range namespaceReqs.requests {

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			response, err := cli.batchListMetricData(ag, req)
			if err != nil {
				continue
			}

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			cli.handleResponse(ag, response, req)
		}
	}
}

func (cli *cesCli) genRequests(ag *agent) {

	cli.requests = make(map[string]*requestsOfNamespace)

	metrics, err := cli.getMetricInfos(ag)
	if err != nil {
		return
	}

	if metrics == nil {
		return
	}

	total := 0
	for _, m := range metrics {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		originalMetricName := m.MetricName
		if m.ExtraInfo != nil && m.ExtraInfo.OriginMetricName != "" {
			originalMetricName = m.ExtraInfo.OriginMetricName
		}

		if ag.checkMetricIgnore(m.Namespace, originalMetricName) {
			continue
		}

		var req *metricsRequest
		var targetNamespace *requestsOfNamespace
		ok := false
		if targetNamespace, ok = cli.requests[m.Namespace]; ok {
			req = targetNamespace.requests[m.MetricName]
		} else {
			targetNamespace = &requestsOfNamespace{
				namespace: m.Namespace,
				requests:  map[string]*metricsRequest{},
			}
			cli.requests[m.Namespace] = targetNamespace
		}

		if req == nil {
			total++
			req = newMetricReq(m.Namespace, m.MetricName, m.Unit)
			req.originalMetricName = originalMetricName
			req.interval = ag.Interval.Duration
			req.delay = ag.Delay.Duration
			req.fixDimensions = ag.applyProperty(req, cli.proj.Id)
			if m.ExtraInfo != nil && m.ExtraInfo.MetricPrefix != "" {
				req.prefix = m.ExtraInfo.MetricPrefix
				if req.tags == nil {
					req.tags = map[string]string{}
				}
				req.tags["prefix"] = m.ExtraInfo.MetricPrefix
			}
			targetNamespace.requests[m.MetricName] = req
		}

		if !req.fixDimensions {
			req.dimensoions = append(req.dimensoions, m.Dimensions...)
		}
	}

	if total > 0 {
		moduleLogger.Debugf("%s(%s): %d namespaces, %d metricnames", cli.proj.Name, cli.proj.Id, len(cli.requests), total)
	}
}

func (cli *cesCli) handleResponse(ag *agent, response []cesmodel.BatchMetricData, req *metricsRequest) {

	for _, data := range response {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		if len(data.Datapoints) == 0 {
			continue
		}

		measurement := formatMeasurement(*data.Namespace)

		tags := map[string]string{}
		tags["region"] = cli.proj.Name
		tags["project_id"] = cli.proj.Id
		extendTags(tags, req.tags, false)
		extendTags(tags, ag.Tags, false)
		if data.Dimensions != nil {
			for _, k := range *data.Dimensions {
				tags[*k.Name] = *k.Value
			}
		}
		if data.Unit != nil {
			tags["unit"] = *data.Unit
		}

		for _, dp := range data.Datapoints {

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			fields := map[string]interface{}{}

			var val float64
			find := false
			switch req.filter {
			case "max":
				if dp.Max != nil {
					val = *dp.Max
					find = true
				}
			case "min":
				if dp.Min != nil {
					val = *dp.Min
					find = true
				}
			case "sum":
				if dp.Sum != nil {
					val = *dp.Sum
					find = true
				}
			case "variance":
				if dp.Variance != nil {
					if v, err := strconv.ParseFloat(*dp.Variance, 64); err == nil {
						val = v
						find = true
					}
				}
			default:
				if dp.Average != nil {
					val = *dp.Average
					find = true
				}
			}

			if find {
				fields[fmt.Sprintf("%s_%s", req.originalMetricName, req.filter)] = val
			}

			tm := time.Unix(dp.Timestamp/1000, 0)

			if len(fields) == 0 {
				moduleLogger.Warnf("skip %s.%s datapoint for no fields, %s", req.namespace, req.metricname, dp.String())
				continue
			}

			if ag.isTestOnce() {
				// pass
			} else if ag.isDebug() {
				data, _ := io.MakeMetric(measurement, tags, fields, tm)
				fmt.Printf("%s\n", string(data))

			} else {
				io.NamedFeedEx(inputName, datakit.Metric, measurement, tags, fields, tm)
			}
		}
	}
}
