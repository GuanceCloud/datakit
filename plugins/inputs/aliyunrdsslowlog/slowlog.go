package aliyunrdsslowlog

import (
	"context"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

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
			r.cfg.Interval.Duration = time.Minute * 10
		}

		a.runningInstances = append(a.runningInstances, r)

		go r.run(a.ctx)
	}
	return nil
}

func (a *AliyunRDSSlowLog) Stop() {
	a.cancelFun()
}

// func (r *runningInstance) getHistory(ctx context.Context) error {
// 	if r.cfg.From == "" {
// 		return nil
// 	}

// 	endTm := time.Now().Truncate(time.Minute).Add(-r.cfg.Interval.Duration)
// 	request := actiontrail.CreateLookupEventsRequest()
// 	request.Scheme = "https"
// 	request.StartTime = r.cfg.From
// 	request.EndTime = unixTimeStrISO8601(endTm)

// 	reqid, response, err := r.lookupEvents(ctx, request, r.client.LookupEvents)
// 	if err != nil {
// 		r.logger.Errorf("(history)LookupEvents(%s) between %s - %s failed", reqid, request.StartTime, request.EndTime)
// 		return err
// 	}

// 	r.handleResponse(ctx, response)

// 	return nil
// }

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

	//

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		// 实例信息
		// r.getInstance(ctx)

		response, err := r.command()
		if err != nil {

		}

		fmt.Println("======resp", response)

		internal.SleepContext(ctx, r.cfg.Interval.Duration)
	}
}

func (r *runningInstance) command() (*rds.DescribeSlowLogsResponse, error) {
	request := rds.CreateDescribeSlowLogsRequest()
	request.Scheme = "https"
	// request.StartTime = unixTimeStrISO8601(startTm)
	// request.EndTime = unixTimeStrISO8601(startTm.Add(r.cfg.Interval.Duration))

	request.DBInstanceId = "xxxxx"
	request.SortKey = "TotalExecutionCounts"

	response, err := r.client.DescribeSlowLogs(request)
	if err != nil {
		// r.logger.Errorf("LookupEvents(%s) between %s - %s failed", reqid, request.StartTime, request.EndTime)
	}

	return response, err
}

func (r *runningInstance) handleMysqlResponse(response *rds.DescribeSlowLogsResponse) error {
	if response == nil {
		return nil
	}

	for _, point := range response.Items.SQLSlowLog {
		tags := map[string]string{}
		fields := map[string]interface{}{}
		tags["db_server"] = "server"                          // todo
		tags["instance_id"] = "instance_id"                   //todo
		tags["instance_description"] = "instance_description" //todo
		tags["region_id"] = "region_id"                       //todo
		tags["product"] = "rds"                               //todo
		tags["engine"] = "engine"                             //todo
		tags["sql_text"] = "sql_text"                         //todo

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

		r.agent.accumulator.AddFields(r.metricName, fields, tags)
	}

	return nil
}

func (r *runningInstance) handleSqlServerResponse(response *rds.DescribeSlowLogsResponse) error {
	if response == nil {
		return nil
	}

	for _, point := range response.Items.SQLSlowLog {
		tags := map[string]string{}
		fields := map[string]interface{}{}
		tags["db_server"] = "server"                          // todo
		tags["instance_id"] = "instance_id"                   //todo
		tags["instance_description"] = "instance_description" //todo
		tags["region_id"] = "region_id"                       //todo
		tags["product"] = "rds"                               //todo
		tags["engine"] = "engine"                             //todo
		tags["sql_text"] = "sql_text"                         //todo

		fields["create_time"] = point.CreateTime
		fields["db_name"] = point.DBName
		fields["max_execution_time"] = point.MaxExecutionTime
		fields["max_lock_time"] = point.MaxLockTime
		fields["sqlserver_total_execution_counts"] = point.MySQLTotalExecutionCounts
		fields["sqlserver_total_execution_times"] = point.MySQLTotalExecutionTimes
		fields["parse_max_row_count"] = point.ParseMaxRowCount
		fields["parse_total_row_counts"] = point.ParseTotalRowCounts
		fields["return_max_row_count"] = point.ReturnMaxRowCount
		fields["return_total_row_counts"] = point.ReturnTotalRowCounts
		fields["sql_text"] = point.SQLText
		fields["total_lock_times"] = point.TotalLockTimes

		r.agent.accumulator.AddFields(r.metricName, fields, tags)
	}

	return nil
}

func unixTimeStrISO8601(t time.Time) string {
	_, zoff := t.Zone()
	nt := t.Add(-(time.Duration(zoff) * time.Second))
	s := nt.Format(`2006-01-02T15:04:05Z`)
	return s
}

func init() {
	inputs.Add("aliyunrdsslowLog", func() telegraf.Input {
		ac := &AliyunRDSSlowLog{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
