package jdcloudmonitor

import (
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/influxdata/toml"

	. "github.com/jdcloud-api/jdcloud-sdk-go/core"

	vmapis "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	vmclient "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"

	jcwapis "github.com/jdcloud-api/jdcloud-sdk-go/services/monitor/apis"
	jcwclient "github.com/jdcloud-api/jdcloud-sdk-go/services/monitor/client"
)

//https://docs.jdcloud.com/cn/monitoring/product-overview

func getTestClient() *jcwclient.MonitorClient {
	data, err := ioutil.ReadFile("./test.conf")
	if err != nil {
		log.Fatalf("%s", err)
	}
	var ag agent
	err = toml.Unmarshal(data, &ag)
	if err != nil {
		log.Fatalf("%s", err)
	}

	credentials := NewCredentials(ag.AccessKeyID, ag.AccessKeySecret)
	client := jcwclient.NewMonitorClient(credentials)
	return client
}

func TestMetricData(t *testing.T) {

	client := getTestClient()

	serviceCode := "vm"
	resourceid := "i-dcnxfmbf5m"

	now := time.Now().Truncate(time.Minute)
	start := now.Add(-5 * time.Minute).Format("2006-01-02T15:04:05Z0700")
	end := now.Format("2006-01-02T15:04:05Z0700")

	req := jcwapis.NewDescribeMetricDataRequest("cn-south-1", "cpu_util", resourceid)
	req.ServiceCode = &serviceCode
	//req.SetDownSampleType("max")
	req.SetStartTime(start)
	req.SetEndTime(end)
	resp, err := client.DescribeMetricData(req)

	if err != nil {
		log.Fatalf("%s", err)
		return
	}
	if resp.Error.Code != 0 {
		log.Fatalf("%v", resp.Error)
		return
	}

	for _, m := range resp.Result.MetricDatas {
		log.Printf("Aggregator=%s, Period=%s, CalculateUnit=%s, Tags=%v", m.Metric.Aggregator, m.Metric.Period, m.Metric.CalculateUnit, m.Tags)
		for _, d := range m.Data {
			tm := time.Unix(d.Timestamp/1000, 0)
			log.Printf("%v - %v", d.Value, tm)
		}
	}
}

func TestListMetrics(t *testing.T) {
	client := getTestClient()

	req := jcwapis.NewDescribeMetricsRequest("vm")
	resp, err := client.DescribeMetrics(req)
	if err != nil {
		log.Fatalf("error, %s", err)
		return
	}

	if resp.Error.Code != 0 {
		log.Fatalf("%v", resp.Error)
		return
	}

	for _, m := range resp.Result.Metrics {
		log.Printf("%s - %s: Dimension=%s, DownSample=%s, CalculateUnit=%s", m.Metric, m.MetricName, m.Dimension, m.DownSample, m.CalculateUnit)
	}
}

func TestListServices(t *testing.T) {
	client := getTestClient()

	req := jcwapis.NewDescribeServicesRequest()
	resp, err := client.DescribeServices(req)
	if err != nil {
		log.Fatalf("error, %s", err)
		return
	}
	if resp.Error.Code != 0 {
		log.Fatalf("%v", resp.Error)
		return
	}
	for _, s := range resp.Result.Services {
		log.Printf("%s (%s)", s.ServiceName, s.ServiceCode)
	}
}

func TestGetVMs(t *testing.T) {
	accessKey := ""
	secretKey := ""
	credentials := NewCredentials(accessKey, secretKey)
	client := vmclient.NewVmClient(credentials)

	req := vmapis.NewDescribeInstancesRequest("cn-south-1")
	resp, err := client.DescribeInstances(req)
	if err != nil {
		return
	}
	for _, m := range resp.Result.Instances {
		log.Printf("%s", m.InstanceId)
	}
}

func TestInput(t *testing.T) {

	data, err := ioutil.ReadFile("test.conf")
	if err != nil {
		t.Error(err)
	}
	ag := newAgent()
	ag.debugMode = true
	if err = toml.Unmarshal(data, &ag); err != nil {
		t.Error(err)
	}
	ag.Run()
}
