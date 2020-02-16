package tencentcms

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/limiter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	es "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/es/v20180416"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	postgres "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/postgres/v20170312"
	redis "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
	sqlserver "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sqlserver/v20180328"
)

var (
	batchInterval = time.Duration(5) * time.Minute
	metricPeriod  = time.Duration(5 * time.Minute)
	rateLimit     = 20

	ErrDimensionsRequired = fmt.Errorf(`dimension for this metric is required`)
	ErrUnsupportAllInst   = fmt.Errorf(`unsupport for fetch all`)
)

type (
	MetricsPeriodInfo map[string][]*monitor.PeriodsSt

	TencentCms struct {
		CMSs []*CMS `toml:"cms"`

		runningCms []*RunningCMS

		tags map[string]string

		ctx       context.Context
		cancelFun context.CancelFunc

		accumulator telegraf.Accumulator

		logger *models.Logger
	}

	RunningCMS struct {
		cfg *CMS
		tc  *TencentCms

		timer *time.Timer

		wg sync.WaitGroup

		credential *common.Credential
		cpf        *profile.ClientProfile
		client     *monitor.Client

		periodsInfos map[string]MetricsPeriodInfo

		ctx context.Context

		logger *models.Logger
	}
)

func (_ *TencentCms) SampleConfig() string {
	return cmsConfigSample
}

func (_ *TencentCms) Description() string {
	return ""
}

func (tc *TencentCms) Gather(telegraf.Accumulator) error {
	return nil
}

func (tc *TencentCms) Stop() {
	tc.cancelFun()
}

func (c *TencentCms) Init() error {

	c.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `tencentcms`,
	}

	var err error
	for _, item := range c.CMSs {
		if item.AccessKeySecret == "" || item.AccessKeyID == "" {
			return fmt.Errorf("access_key_id or access_key_secret must not be empty")
		}
		for _, p := range item.Namespace {
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
	}

	return nil
}

func (tc *TencentCms) Start(acc telegraf.Accumulator) error {

	if len(tc.CMSs) == 0 {
		tc.logger.Warnf("no configuration found")
		return nil
	}

	tc.logger.Infof("starting...")
	tc.accumulator = acc

	for _, c := range tc.CMSs {
		a := &RunningCMS{
			cfg:          c,
			ctx:          tc.ctx,
			tc:           tc,
			periodsInfos: map[string]MetricsPeriodInfo{},
			logger:       tc.logger,
		}
		tc.runningCms = append(tc.runningCms, a)
	}

	for _, c := range tc.runningCms {
		go func(ac *RunningCMS) {

			if err := ac.run(); err != nil && err != context.Canceled {
				ac.logger.Errorf("%s", err)
			}
		}(c)

		c.logger.Infof("%s done", c.cfg.AccessKeyID)
	}

	return nil
}

func (s *RunningCMS) run() error {

	s.credential = common.NewCredential(s.cfg.AccessKeyID, s.cfg.AccessKeySecret)
	s.cpf = profile.NewClientProfile()
	s.cpf.HttpProfile.Endpoint = "monitor.tencentcloudapi.com"
	s.client, _ = monitor.NewClient(s.credential, s.cfg.RegionID, s.cpf)

	for _, ns := range s.cfg.Namespace {
		if err := s.fetchAvatiableMetrics(ns); err != nil {
			s.logger.Errorf("fail to get base metrics of namespace \"%s\": %s", ns.Name, err)
			return err
		}
	}

	select {
	case <-s.ctx.Done():
		return context.Canceled
	default:
	}

	lmtr := limiter.NewRateLimiter(rateLimit, time.Second)
	defer lmtr.Stop()

	s.wg.Add(1)
	defer s.wg.Done()

	var err error

	for {

		select {
		case <-s.ctx.Done():
			return context.Canceled
		default:
		}

		t := time.Now()
		for _, req := range MetricsRequests {

			select {
			case <-s.ctx.Done():
				return context.Canceled
			default:
			}

			<-lmtr.C
			_ = req
			if err = s.fetchMetrics(req); err != nil {
				s.logger.Errorf(`get tencent metric "%s.%s" failed: %s`, *req.q.Namespace, *req.q.MetricName, err)
			}
		}

		useage := time.Now().Sub(t)
		if useage < batchInterval {
			remain := batchInterval - useage

			if s.timer == nil {
				s.timer = time.NewTimer(remain)
			} else {
				s.timer.Reset(remain)
			}
			select {
			case <-s.ctx.Done():
				if s.timer != nil {
					s.timer.Stop()
					s.timer = nil
				}
				return context.Canceled
			case <-s.timer.C:
			}
		}
	}
}

func (c *RunningCMS) fetchMetrics(req *MetricsRequest) error {

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
			c.logger.Warnf("period of %v for %s not support, change to %v", *req.q.Period, *req.q.MetricName, minpt)
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

	c.logger.Debugf("request: %s", req.q.ToJsonString())

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
			return err
		}

	}

	c.logger.Debugf("D! [tencentcms] Response: Period=%v, StartTime=%s, EndTime=%s", *resp.Response.Period, *resp.Response.StartTime, *resp.Response.EndTime)

	for _, dp := range resp.Response.DataPoints {

		tags := map[string]string{}
		if c.tc.tags != nil {
			for k, v := range c.tc.tags {
				tags[k] = v
			}
		}

		for _, dm := range dp.Dimensions {
			tags[*dm.Name] = *dm.Value
		}

		for _, val := range dp.Values {

			fields := map[string]interface{}{}
			fields[*req.q.MetricName] = *val

			if c.tc.accumulator != nil {
				c.tc.accumulator.AddFields(foramtNamespaceName(*req.q.Namespace), fields, tags)
			}
		}
	}

	return nil
}

func (c *RunningCMS) fetchAvatiableMetrics(namespace *Namespace) error {

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

		c.logger.Debugf("get base metrics of %s ok: %s", namespace.Name, pinfo.String())

		return nil

	}

	return err
}

func (c *RunningCMS) fetchAllInstanceIds(namespace string) ([]string, error) {

	cpf := profile.NewClientProfile()

	var err error

	instanceIds := []string{}

	switch namespace {
	case "QCE/CVM":
		client, _ := cvm.NewClient(c.credential, c.cfg.RegionID, cpf)
		request := cvm.NewDescribeInstancesRequest()
		var response *cvm.DescribeInstancesResponse
		response, err = client.DescribeInstances(request)
		if err == nil {
			for _, inst := range response.Response.InstanceSet {
				instanceIds = append(instanceIds, *inst.InstanceId)
			}
		}
	case "QCE/CDB":
		client, _ := cdb.NewClient(c.credential, c.cfg.RegionID, cpf)
		request := cdb.NewDescribeDBInstancesRequest()
		var response *cdb.DescribeDBInstancesResponse
		response, err = client.DescribeDBInstances(request)
		if err == nil {
			for _, inst := range response.Response.Items {
				instanceIds = append(instanceIds, *inst.InstanceId)
			}
		}

	case `QCE/POSTGRES`:
		client, _ := postgres.NewClient(c.credential, c.cfg.RegionID, cpf)
		request := postgres.NewDescribeDBInstancesRequest()
		var response *postgres.DescribeDBInstancesResponse
		response, err = client.DescribeDBInstances(request)
		if err == nil {
			for _, inst := range response.Response.DBInstanceSet {
				instanceIds = append(instanceIds, *inst.DBInstanceId)
			}
		}

	case `QCE/REDIS`:
		client, _ := redis.NewClient(c.credential, c.cfg.RegionID, cpf)
		request := redis.NewDescribeInstancesRequest()
		var response *redis.DescribeInstancesResponse
		response, err = client.DescribeInstances(request)
		if err == nil {
			for _, inst := range response.Response.InstanceSet {
				instanceIds = append(instanceIds, *inst.InstanceId)
			}
		}

	case `QCE/CES`:
		client, _ := es.NewClient(c.credential, c.cfg.RegionID, cpf)
		request := es.NewDescribeInstancesRequest()
		var response *es.DescribeInstancesResponse
		response, err = client.DescribeInstances(request)
		if err == nil {
			for _, inst := range response.Response.InstanceList {
				instanceIds = append(instanceIds, *inst.InstanceId)
			}
		}

	case `QCE/SQLSERVER`:
		client, _ := sqlserver.NewClient(c.credential, c.cfg.RegionID, cpf)
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

	c.logger.Debugf("all instanceids: %#v", instanceIds)

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
	inputs.Add("tencentcms", func() telegraf.Input {
		ac := &TencentCms{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
