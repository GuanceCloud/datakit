package aliyunrdsslowlog

import (
	"context"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var regions = []string{
	"cn-hangzhou",
	"cn-shanghai",
	"cn-qingdao",
	"cn-beijing",
	"cn-zhangjiakou",
	"cn-huhehaote",
	"cn-shenzhen",
	"cn-heyuan",
	"cn-chengdu",
	"cn-hongkong",
	"ap-southeast-1",
	"ap-southeast-2",
	"ap-southeast-3",
	"ap-southeast-5",
	"ap-northeast-1",
	"ap-south-1",
	"eu-central-1",
	"eu-west-1",
	"us-east-1",
}

type AliyunRDSSlowLog struct {
	RDSslowlog  []*RDSslowlog
	ctx         context.Context
	cancelFun   context.CancelFunc
	accumulator telegraf.Accumulator
	logger      *models.Logger

	runningInstances []*runningInstance
}

type runningInstance struct {
	cfg        *RDSslowlog
	agent      *AliyunRDSSlowLog
	logger     *models.Logger
	client     *rds.Client
	metricName string
}

type rdsInstance struct {
	id           string
	description  string
	server       string
	regionId     string
	engine       string
	instanceType string
}

func (_ *AliyunRDSSlowLog) SampleConfig() string {
	return configSample
}

func (_ *AliyunRDSSlowLog) Description() string {
	return ""
}

func (_ *AliyunRDSSlowLog) Gather(telegraf.Accumulator) error {
	return nil
}

func (_ *AliyunRDSSlowLog) Init() error {
	return nil
}

func (a *AliyunRDSSlowLog) Start(acc telegraf.Accumulator) error {
	a.logger = &models.Logger{
		Name: `aliyunrdsslowlog`,
	}

	if len(a.RDSslowlog) == 0 {
		a.logger.Warnf("no configuration found")
		return nil
	}

	a.logger.Infof("starting...")

	a.accumulator = acc

	for _, instCfg := range a.RDSslowlog {
		r := &runningInstance{
			cfg:    instCfg,
			agent:  a,
			logger: a.logger,
		}
		r.metricName = instCfg.MetricName
		if r.metricName == "" {
			r.metricName = "aliyun_rds_slow_log"
		}

		if r.cfg.Interval.Duration == 0 {
			r.cfg.Interval.Duration = time.Hour * 24
		}

		cli, err := rds.NewClientWithAccessKey(instCfg.RegionID, instCfg.AccessKeyID, instCfg.AccessKeySecret)
		if err != nil {
			r.logger.Errorf("create client failed, %s", err)
			return err
		}

		r.client = cli

		a.runningInstances = append(a.runningInstances, r)

		go r.run(a.ctx)
	}
	return nil
}

func (a *AliyunRDSSlowLog) Stop() {
	a.cancelFun()
}

func (r *runningInstance) run(ctx context.Context) error {
	defer func() {
		if e := recover(); e != nil {

		}
	}()

	cli, err := rds.NewClientWithAccessKey(r.cfg.RegionID, r.cfg.AccessKeyID, r.cfg.AccessKeySecret)
	if err != nil {
		r.logger.Errorf("create client failed, %s", err)
		return err
	}
	r.client = cli

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		// 实例信息
		for _, val := range r.cfg.Product {
			r.exec(val)
		}

		internal.SleepContext(ctx, r.cfg.Interval.Duration)
	}
}

func (r *runningInstance) exec(engine string) {
	for _, region := range regions {
		go r.getInstance(engine, region)
	}
}

func (r *runningInstance) getInstance(engine string, regionID string) error {
	var pageNumber = 1
	var pageSize = 50

	for {
		request := rds.CreateDescribeDBInstancesRequest()
		request.Scheme = "https"
		request.RegionId = regionID
		request.Engine = engine
		request.PageSize = requests.NewInteger(pageSize)
		request.PageNumber = requests.NewInteger(pageNumber)

		response, err := r.client.DescribeDBInstances(request)
		if err != nil {
			r.logger.Error("instance failed")
			return err
		}

		for _, item := range response.Items.DBInstance {
			instanceObj := &rdsInstance{
				id:           item.DBInstanceId,
				description:  item.DBInstanceDescription,
				server:       item.ConnectionMode,
				regionId:     item.RegionId,
				engine:       item.Engine,
				instanceType: item.DBInstanceType,
			}

			go r.command(engine, instanceObj)
		}

		total := response.TotalRecordCount
		if pageNumber*pageSize >= total {
			break
		}

		pageNumber = pageNumber + 1
	}
	return nil
}

func (r *runningInstance) command(engine string, instanceObj *rdsInstance) {
	et := time.Now()
	st := et.Add(-time.Hour * 24)

	request := rds.CreateDescribeSlowLogsRequest()
	request.Scheme = "https"
	request.StartTime = unixTimeStrISO8601(st)
	request.EndTime = unixTimeStrISO8601(et)

	request.DBInstanceId = instanceObj.id
	request.SortKey = "TotalExecutionCounts"

	response, err := r.client.DescribeSlowLogs(request)
	if err != nil {
		r.logger.Errorf("DescribeSlowLogs error %v", err)
	}

	go r.handleResponse(response, engine, instanceObj)
}

func (r *runningInstance) handleResponse(response *rds.DescribeSlowLogsResponse, product string, instanceObj *rdsInstance) error {
	if response == nil {
		return nil
	}

	for _, point := range response.Items.SQLSlowLog {
		tags := map[string]string{}
		fields := map[string]interface{}{}
		tags["instance_id"] = instanceObj.id
		tags["instance_description"] = instanceObj.description
		tags["region_id"] = instanceObj.regionId
		tags["product"] = "rds"
		tags["engine"] = product

		fields["create_time"] = point.CreateTime
		fields["db_name"] = point.DBName
		fields["max_execution_time"] = point.MaxExecutionTime
		fields["max_lock_time"] = point.MaxLockTime
		fields["mysql_total_execution_counts"] = point.MySQLTotalExecutionCounts
		fields["mysql_total_execution_times"] = point.MySQLTotalExecutionTimes
		fields["parse_max_row_count"] = point.ParseMaxRowCount
		fields["parse_total_row_counts"] = point.ParseTotalRowCounts
		fields["return_max_row_count"] = point.ReturnMaxRowCount
		fields["return_total_row_counts"] = point.ReturnTotalRowCounts
		fields["sql_text"] = point.SQLText
		fields["total_lock_times"] = point.TotalLockTimes

		fields["sqlserver_total_execution_counts"] = point.SQLServerTotalExecutionCounts
		fields["sqlserver_total_execution_times"] = point.SQLServerTotalExecutionTimes
		fields["return_max_row_count"] = point.ReturnMaxRowCount

		r.agent.accumulator.AddFields(r.metricName, fields, tags)
	}

	return nil
}

func unixTimeStrISO8601(t time.Time) string {
	_, zoff := t.Zone()
	nt := t.Add(-(time.Duration(zoff) * time.Second))
	s := nt.Format(`2006-01-02Z`)
	return s
}

func init() {
	inputs.Add("aliyunrdsslowLog", func() telegraf.Input {
		ac := &AliyunRDSSlowLog{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
