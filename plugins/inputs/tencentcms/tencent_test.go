package tencentcms

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

func TestCMSServe(t *testing.T) {

	// if err := Cfg.Load(`./demo.toml`); err != nil {
	// 	log.Fatalln(err)
	// }

	// logHandler, _ := log.NewStreamHandler(os.Stdout)

	// ll := log.NewDefault(logHandler)
	// ll.SetLevel(log.LevelDebug)

	// svr := &TencentCMSSvr{
	// 	logger: ll,
	// }

	// ctx, _ := context.WithCancel(context.Background())

	// svr.Start(ctx, nil)
}

func TestGetBaseMetrics(t *testing.T) {
	credential := common.NewCredential(
		`AKIDXVFk7EfwUmzvc9mZR3TnqlTwZuyoJTxK`,
		`HH5f4PUJsBLFBrQjlY5TcXEuKTP7i5HG`,
	)
	cpf := profile.NewClientProfile()
	//cpf.HttpProfile.Endpoint = "monitor.tencentcloudapi.com"

	regionID := regions.Shanghai
	namespace := `qce/cvm`

	client, _ := monitor.NewClient(credential, regionID, cpf)
	request := monitor.NewDescribeBaseMetricsRequest()
	request.Namespace = common.StringPtr(namespace)
	response, err := client.DescribeBaseMetrics(request)

	if response != nil {
		for _, m := range response.Response.MetricSet {
			period := ""
			for _, pr := range m.Periods {
				statTyps := []string{}
				for _, st := range pr.StatType {
					statTyps = append(statTyps, *st)
				}
				period += fmt.Sprintf("%s(%s),", *pr.Period, strings.Join(statTyps, ","))
			}
			fmt.Printf("MetricName=%s, Period=%s\n", *m.MetricName, period)
		}
	} else {
		fmt.Printf("An API error has returned: %s", err)
		return
	}
}

func TestGetCVMInstanceIds(t *testing.T) {

	credential := common.NewCredential(
		`AKIDXVFk7EfwUmzvc9mZR3TnqlTwZuyoJTxK`,
		`HH5f4PUJsBLFBrQjlY5TcXEuKTP7i5HG`,
	)

	regionID := regions.Shanghai

	cpf := profile.NewClientProfile()
	client, _ := cvm.NewClient(credential, regionID, cpf)

	request := cvm.NewDescribeInstancesRequest()
	response, err := client.DescribeInstances(request)

	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Printf("An API error has returned: %s", err)
		return
	}

	for _, v := range response.Response.InstanceSet {
		log.Printf("*** InstanceName=%s, InstanceId=%s, InstanceType=%s, InstanceState=%s", *v.InstanceName, *v.InstanceId, *v.InstanceType, *v.InstanceState)
	}
}

func TestGetCVMMetrics(t *testing.T) {

	credential := common.NewCredential(
		`AKIDXVFk7EfwUmzvc9mZR3TnqlTwZuyoJTxK`,
		`HH5f4PUJsBLFBrQjlY5TcXEuKTP7i5HG`,
	)

	regionID := regions.Shanghai

	cpf := profile.NewClientProfile()
	client, _ := monitor.NewClient(credential, regionID, cpf)

	request := monitor.NewGetMonitorDataRequest()
	request.Namespace = common.StringPtr(`QCE/CVM`)
	request.MetricName = common.StringPtr("acc_outtraffic")
	request.Instances = []*monitor.Instance{
		&monitor.Instance{
			Dimensions: []*monitor.Dimension{
				&monitor.Dimension{
					Name:  common.StringPtr(`InstanceId`),
					Value: common.StringPtr(`ins-9bpjauir`),
				},
			},
		},
	}
	request.Period = common.Uint64Ptr(60)

	nt := time.Now().Add(-time.Hour)
	et := nt.Format(time.RFC3339)
	delta, _ := time.ParseDuration("-15m")
	st := nt.Add(delta).Format(time.RFC3339)

	log.Printf("StartTime=%s, EndTime=%s", st, et)

	request.StartTime = common.StringPtr(st)
	request.EndTime = common.StringPtr(et)

	response, err := client.GetMonitorData(request)

	if err != nil {
		log.Printf("An API error has returned: %s", err)
		return
	}

	for _, dp := range response.Response.DataPoints {
		log.Printf("--- DataPoint ---")
		for i, val := range dp.Values {
			log.Printf("*** %v - %v", *val, *dp.Timestamps[i])
		}
	}
}
