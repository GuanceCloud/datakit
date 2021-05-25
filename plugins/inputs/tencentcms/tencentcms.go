package tencentcms

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"

	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	es "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/es/v20180416"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	postgres "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/postgres/v20170312"
	redis "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
	sqlserver "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sqlserver/v20180328"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `tencentcms`
	moduleLogger *logger.Logger

	batchInterval = time.Duration(5) * time.Minute
	metricPeriod  = time.Duration(5 * time.Minute)
	rateLimit     = 20

	ErrDimensionsRequired = fmt.Errorf(`dimension for this metric is required`)
	ErrUnsupportAllInst   = fmt.Errorf(`unsupport for fetch all`)
)

type (
	MetricsPeriodInfo map[string][]*monitor.PeriodsSt
)

func (_ *CMS) Catalog() string {
	return `tencentcloud`
}

func (_ *CMS) SampleConfig() string {
	return cmsConfigSample
}

func (c *CMS) initialize() error {

	var err error
	for _, p := range c.Namespace {
		for _, m := range p.Metrics.MetricNames {
			req := &MetricsRequest{q: monitor.NewGetMonitorDataRequest()}
			req.q.Period = common.Uint64Ptr(60)
			req.q.MetricName = common.StringPtr(m)
			req.q.Namespace = common.StringPtr(p.Name)
			if req.q.Instances, err = p.MakeDimension(m); err != nil {
				return err
			}
			MetricsRequests = append(MetricsRequests, req)
		}
	}
	return nil
}

func (tc *CMS) Run() {

	moduleLogger = logger.SLogger(inputName)

	for {
		select {
		case <-datakit.Exit.Wait():
			return
		default:
		}

		if err := tc.initialize(); err != nil {
			moduleLogger.Errorf("%s", err)
			if tc.isTest() {
				tc.testError = err
				return
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	go func() {
		<-datakit.Exit.Wait()
		tc.cancelFun()
	}()

	tc.periodsInfos = map[string]MetricsPeriodInfo{}
	tc.run()
}

func (s *CMS) run() error {

	s.credential = common.NewCredential(s.AccessKeyID, s.AccessKeySecret)
	s.cpf = profile.NewClientProfile()
	s.cpf.HttpProfile.Endpoint = "monitor.tencentcloudapi.com"
	s.client, _ = monitor.NewClient(s.credential, s.RegionID, s.cpf)

	for _, ns := range s.Namespace {
		if err := s.fetchAvatiableMetrics(ns); err != nil {
			moduleLogger.Errorf("fail to get base metrics of namespace \"%s\": %s", ns.Name, err)
			if s.isTest() {
				s.testError = err
			}
			return err
		}
	}

	select {
	case <-s.ctx.Done():
		return context.Canceled
	default:
	}

	limit := rate.Every(50 * time.Millisecond)
	rateLimiter := rate.NewLimiter(limit, 1)

	var err error

	for {

		select {
		case <-s.ctx.Done():
			return nil
		default:
		}

		for _, req := range MetricsRequests {

			select {
			case <-s.ctx.Done():
				return nil
			default:
			}

			rateLimiter.Wait(s.ctx)
			if err = s.fetchMetrics(req); err != nil {
				moduleLogger.Errorf(`get tencent metric "%s.%s" failed: %s`, *req.q.Namespace, *req.q.MetricName, err)
			}
		}

		if s.isTest() {
			return nil
		}

		datakit.SleepContext(s.ctx, batchInterval)
	}
}

func (c *CMS) fetchMetrics(req *MetricsRequest) error {

	if len(req.q.Instances) == 0 {
		ids, err := c.fetchAllInstanceIds(strings.ToUpper(*req.q.Namespace))
		if err == nil {
			for _, id := range ids {
				inst := &monitor.Instance{
					Dimensions: []*monitor.Dimension{
						&monitor.Dimension{
							Name:  common.StringPtr(getInstanceKeyName(strings.ToUpper(*req.q.Namespace))),
							Value: common.StringPtr(id),
						},
					},
				}
				req.q.Instances = append(req.q.Instances, inst)
			}
		}
	}

	if !req.checkPeriod {
		validPeriod := false
		var minpt int64
		if pinfo, ok := c.periodsInfos[*req.q.Namespace]; ok {
			if pr, ok := pinfo[*req.q.MetricName]; ok {
				for _, pst := range pr {
					ptime, err := strconv.ParseInt(*pst.Period, 10, 64)
					if err == nil {
						if *req.q.Period == uint64(ptime) {
							validPeriod = true
							break
						}
						if ptime < minpt || minpt == 0 {
							minpt = ptime
						}
					}
				}
			}
		}

		if !validPeriod {
			moduleLogger.Warnf("period of %v for %s not support, change to %v", *req.q.Period, *req.q.MetricName, minpt)
			req.q.Period = common.Uint64Ptr(uint64(minpt))
		}

		req.checkPeriod = true
	}

	nt := time.Now().Truncate(time.Minute)
	delta, _ := time.ParseDuration(`-1m`)
	nt = nt.Add(delta)

	et := nt.Format(time.RFC3339)
	delta, _ = time.ParseDuration(`-4m`)
	st := nt.Add(delta).Format(time.RFC3339)

	req.q.EndTime = common.StringPtr(et)
	req.q.StartTime = common.StringPtr(st)

	moduleLogger.Debugf("request: %s", req.q.ToJsonString())

	resp, err := c.client.GetMonitorData(req.q)
	if err != nil {

		nq := monitor.NewGetMonitorDataRequest()
		nq.Period = common.Uint64Ptr(*req.q.Period)
		nq.MetricName = common.StringPtr(*req.q.MetricName)
		nq.Namespace = common.StringPtr(*req.q.Namespace)
		nq.Instances = req.q.Instances
		nq.EndTime = common.StringPtr(*req.q.StartTime)
		nq.StartTime = common.StringPtr(*req.q.EndTime)

		req.q = nq

		if resp, err = c.client.GetMonitorData(req.q); err != nil {
			if c.isTest() {
				c.testError = err
			}
			return err
		}

	}

	moduleLogger.Debugf("D! [tencentcms] Response: Period=%v, StartTime=%s, EndTime=%s", *resp.Response.Period, *resp.Response.StartTime, *resp.Response.EndTime)

	for _, dp := range resp.Response.DataPoints {

		tags := map[string]string{}
		if c.tags != nil {
			for k, v := range c.tags {
				tags[k] = v
			}
		}

		for _, dm := range dp.Dimensions {
			tags[*dm.Name] = *dm.Value
		}

		for _, val := range dp.Values {

			fields := map[string]interface{}{}
			fields[*req.q.MetricName] = *val

			if c.isTest() {
				// pass
			} else {
				io.NamedFeedEx(inputName, datakit.Metric, foramtNamespaceName(*req.q.Namespace), tags, fields)
			}

		}
	}

	return nil
}

func (c *CMS) fetchAvatiableMetrics(namespace *Namespace) error {

	request := monitor.NewDescribeBaseMetricsRequest()
	request.Namespace = common.StringPtr(namespace.Name)
	response, err := c.client.DescribeBaseMetrics(request)

	metricsNames := ";" + strings.Join(namespace.Metrics.MetricNames, ";") + ";"

	if response != nil && response.Response != nil && response.Response.MetricSet != nil {

		pinfo := make(MetricsPeriodInfo)

		for _, m := range response.Response.MetricSet {
			if strings.Contains(metricsNames, ";"+*m.MetricName+";") {
				pinfo[*m.MetricName] = m.Periods
			}
		}

		c.periodsInfos[namespace.Name] = pinfo

		moduleLogger.Debugf("get base metrics of %s ok: %s", namespace.Name, pinfo.String())

		return nil

	}

	return err
}

func (c *CMS) fetchAllInstanceIds(namespace string) ([]string, error) {

	cpf := profile.NewClientProfile()

	var err error

	instanceIds := []string{}

	switch namespace {
	case "QCE/CVM":
		client, _ := cvm.NewClient(c.credential, c.RegionID, cpf)
		request := cvm.NewDescribeInstancesRequest()
		var response *cvm.DescribeInstancesResponse
		response, err = client.DescribeInstances(request)
		if err == nil {
			for _, inst := range response.Response.InstanceSet {
				instanceIds = append(instanceIds, *inst.InstanceId)
			}
		}
	case "QCE/CDB":
		client, _ := cdb.NewClient(c.credential, c.RegionID, cpf)
		request := cdb.NewDescribeDBInstancesRequest()
		var response *cdb.DescribeDBInstancesResponse
		response, err = client.DescribeDBInstances(request)
		if err == nil {
			for _, inst := range response.Response.Items {
				instanceIds = append(instanceIds, *inst.InstanceId)
			}
		}

	case `QCE/POSTGRES`:
		client, _ := postgres.NewClient(c.credential, c.RegionID, cpf)
		request := postgres.NewDescribeDBInstancesRequest()
		var response *postgres.DescribeDBInstancesResponse
		response, err = client.DescribeDBInstances(request)
		if err == nil {
			for _, inst := range response.Response.DBInstanceSet {
				instanceIds = append(instanceIds, *inst.DBInstanceId)
			}
		}

	case `QCE/REDIS`:
		client, _ := redis.NewClient(c.credential, c.RegionID, cpf)
		request := redis.NewDescribeInstancesRequest()
		var response *redis.DescribeInstancesResponse
		response, err = client.DescribeInstances(request)
		if err == nil {
			for _, inst := range response.Response.InstanceSet {
				instanceIds = append(instanceIds, *inst.InstanceId)
			}
		}

	case `QCE/CES`:
		client, _ := es.NewClient(c.credential, c.RegionID, cpf)
		request := es.NewDescribeInstancesRequest()
		var response *es.DescribeInstancesResponse
		response, err = client.DescribeInstances(request)
		if err == nil {
			for _, inst := range response.Response.InstanceList {
				instanceIds = append(instanceIds, *inst.InstanceId)
			}
		}

	case `QCE/SQLSERVER`:
		client, _ := sqlserver.NewClient(c.credential, c.RegionID, cpf)
		request := sqlserver.NewDescribeDBInstancesRequest()
		var response *sqlserver.DescribeDBInstancesResponse
		response, err = client.DescribeDBInstances(request)
		if err == nil {
			for _, inst := range response.Response.DBInstances {
				instanceIds = append(instanceIds, *inst.InstanceId)
			}
		}

	default:
		err = ErrUnsupportAllInst
	}

	if err != nil {
		return nil, err
	}

	moduleLogger.Debugf("all instanceids: %#v", instanceIds)

	return instanceIds, err
}

func getInstanceKeyName(namespace string) string {
	switch namespace {
	case `QCE/CDB`, `QCE/CVM`:
		return `InstanceId`
	case `QCE/POSTGRES`, `QCE/SQLSERVER`:
		return `resourceId`
	case `QCE/REDIS`:
		return `redis_uuid`
	case `QCE/CES`:
		return `uInstanceId`
	}
	return ""
}

func foramtNamespaceName(namespace string) string {
	return strings.Replace(strings.ToLower(namespace), `/`, `_`, -1)
}

func (m MetricsPeriodInfo) String() string {
	res := ""
	for k, v := range m {
		period := ""
		for _, pr := range v {
			statTyps := []string{}
			for _, st := range pr.StatType {
				statTyps = append(statTyps, *st)
			}
			period += fmt.Sprintf("%s(%s),", *pr.Period, strings.Join(statTyps, ","))
		}
		res += fmt.Sprintf("MetricName=%s, Period=%s\n", k, period)
	}

	return res
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		ac := &CMS{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
