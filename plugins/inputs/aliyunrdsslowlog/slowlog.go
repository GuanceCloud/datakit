package aliyunrdsslowlog

import (
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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

var (
	l *logger.Logger
)

type AliyunRDSSlowLog struct {
	cfg    AliyunRDS
	client *rds.Client
}

type rdsInstance struct {
	id           string
	description  string
	server       string
	regionId     string
	engine       string
	instanceType string
}

func (_ *AliyunRDSSlowLog) Catalog() string {
	return "aliyunRDSSlowlog"
}

func (_ *AliyunRDSSlowLog) SampleConfig() string {
	return configSample
}

func (_ *AliyunRDSSlowLog) Description() string {
	return ""
}

func (_ *AliyunRDSSlowLog) Gather() error {
	return nil
}

func (a *AliyunRDSSlowLog) Run() {
	l = logger.SLogger("aliyunRDSSlowlog")

	l.Info("aliyunRDSSlowlog input started...")

	cli, err := rds.NewClientWithAccessKey(a.cfg.RegionID, a.cfg.AccessKeyID, a.cfg.AccessKeySecret)
	if err != nil {
		l.Errorf("create client failed, %s", err)
	}

	a.client = cli

	interval, err := time.ParseDuration(a.cfg.Interval)
	if err != nil {
		l.Error(err)
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			// 实例信息
			for _, val := range a.cfg.Product {
				a.exec(val)
			}
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (r *AliyunRDSSlowLog) exec(engine string) {
	for _, region := range regions {
		go r.getInstance(engine, region)
	}
}

func (r *AliyunRDSSlowLog) getInstance(engine string, regionID string) error {
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
			l.Error("instance failed")
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

func (r *AliyunRDSSlowLog) command(engine string, instanceObj *rdsInstance) {
	interval, err := time.ParseDuration(r.cfg.Interval)
	if err != nil {
		l.Error(err)
	}

	et := time.Now()
	st := et.Add(-interval)

	request := rds.CreateDescribeSlowLogsRequest()
	request.Scheme = "https"
	request.StartTime = unixTimeStrISO8601(st)
	request.EndTime = unixTimeStrISO8601(et)

	request.DBInstanceId = instanceObj.id
	request.SortKey = "TotalExecutionCounts"

	response, err := r.client.DescribeSlowLogs(request)
	if err != nil {
		l.Errorf("DescribeSlowLogs error %v", err)
	}

	go r.handleResponse(response, engine, instanceObj)
}

func (r *AliyunRDSSlowLog) handleResponse(response *rds.DescribeSlowLogsResponse, product string, instanceObj *rdsInstance) error {
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

		pt, err := io.MakeMetric(r.cfg.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		err = io.Feed([]byte(pt), io.Metric)
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
	inputs.Add("aliyunrdsslowLog", func() inputs.Input {
		return &AliyunRDSSlowLog{}
	})
}
