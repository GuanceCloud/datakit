package aliyuncdn

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type AliyunCDN struct {
	CDN              []*CDN `toml:"cdn"`
	runningInstances []*RunningInstance
	ctx              context.Context
	cancelFun        context.CancelFunc
	accumulator      telegraf.Accumulator
	logger           *models.Logger
}

type RunningInstance struct {
	client     *cdn.Client
	cfg        *CDN
	agent      *AliyunCDN
	logger     *models.Logger
	metricName string
	domains    []string
}

func (_ *AliyunCDN) SampleConfig() string {
	return aliyunCDNConfigSample
}

func (_ *AliyunCDN) Description() string {
	return ""
}

func (_ *AliyunCDN) Gather(telegraf.Accumulator) error {
	return nil
}

func (cdn *AliyunCDN) Start(acc telegraf.Accumulator) error {
	if len(cdn.CDN) == 0 {
		cdn.logger.Warnf("no configuration found")
		return nil
	}

	cdn.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `aliyuncdn`,
	}

	cdn.logger.Infof("aliyun cdn start...")

	cdn.accumulator = acc

	for _, instCfg := range cdn.CDN {
		r := &RunningInstance{
			cfg:    instCfg,
			agent:  cdn,
			logger: cdn.logger,
		}

		cdn.runningInstances = append(cdn.runningInstances, r)

		go r.run(cdn.ctx)
	}

	return nil
}

func (cdn *AliyunCDN) Stop() {
	cdn.cancelFun()
}

func (r *RunningInstance) run(ctx context.Context) error {
	defer func() {
		if e := recover(); e != nil {

		}
	}()

	cli, err := cdn.NewClientWithAccessKey(r.cfg.RegionID, r.cfg.AccessKeyID, r.cfg.AccessKeySecret)
	if err != nil {
		r.logger.Errorf("create client failed, %s", err)
		return err
	}
	r.client = cli

	// 域名check
	r.domains = r.cfg.DomainName
	// if len(r.domains) == 0 {
	// 	r.domains = r.getDomain()
	// }

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		go r.getMetrics()

		internal.SleepContext(ctx, 10*time.Second)
	}
}

func (r *RunningInstance) getMetrics() {
	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(10 * time.Minute)).Unix()
	var startTime = DateISO8601(st)
	var endTime = DateISO8601(et)
	var domainName = strings.Join(r.cfg.DomainName, ",")

	go r.getBpsData(domainName, startTime, endTime)
	go r.getTrafficData(domainName, startTime, endTime)
	go r.getHitRateData(domainName, startTime, endTime)
	go r.getReqHitRateData(domainName, startTime, endTime)
	go r.getDomainSrcBpsData(domainName, startTime, endTime)
	go r.getDomainSrcTrafficData(domainName, startTime, endTime)
}

// 近24小时流量带宽

func (r *RunningInstance) getDomain() []string {
	var pageNumber = 1
	var pageSize = 50
	var domains = []string{}

	for {
		request := cdn.CreateDescribeUserDomainsRequest()
		request.RegionId = r.cfg.RegionID
		request.Scheme = "https"
		request.PageSize = requests.NewInteger(50)
		request.PageNumber = requests.NewInteger(1)
		request.DomainStatus = "online"

		response, err := r.client.DescribeUserDomains(request)
		if err != nil {
			r.logger.Errorf("[DescribeUserDomainsRequest] failed, %v", err.Error())
		}
		for _, item := range response.Domains.PageData {
			domains = append(domains, item.Cname)
		}

		total := response.TotalCount
		if int64(pageNumber*pageSize) >= total {
			break
		}
	}

	return domains
}

func (r *RunningInstance) getBpsData(domain, startTime, endTime string) {
	// bps
	bpsParams := cdn.CreateDescribeDomainBpsDataRequest()
	bpsParams.Scheme = "https"
	bpsParams.DomainName = domain

	bpsParams.StartTime = startTime
	bpsParams.EndTime = endTime

	bpsRes, err := r.client.DescribeDomainBpsData(bpsParams)
	if err != nil {
		r.logger.Errorf("[cdn] action DescribeDomainBpsData failed, %s", err.Error())
	}

	// 指标收集
	for _, point := range bpsRes.BpsDataPerInterval.DataModule {
		tm := RFC3339(point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = "domain"

		fields["value"] = ConvertToNum(point.Value)                           // bps数据值，单位：bit/second。
		fields["domesticValue"] = ConvertToNum(point.DomesticValue)           // 中国内地带宽bps
		fields["tttpsDomesticValue"] = ConvertToNum(point.HttpsDomesticValue) // L1节点https中国内地带宽
		fields["httpsOverseasValue"] = ConvertToNum(point.HttpsOverseasValue) // L1节点全球（不包含中国内地）https带宽
		fields["httpsValue"] = ConvertToNum(point.HttpsValue)                 // L1节点https的带宽数据值，单位：bit/second
		fields["overseasValue"] = ConvertToNum(point.OverseasValue)           // 全球（不包含中国内地）带宽bps

		r.agent.accumulator.AddFields("domainBps", fields, tags, tm)
	}
}

func (r *RunningInstance) getTrafficData(domain, startTime, endTime string) {
	// TrafficData
	trafficParams := cdn.CreateDescribeDomainTrafficDataRequest()
	trafficParams.Scheme = "https"
	trafficParams.DomainName = domain

	trafficParams.StartTime = startTime
	trafficParams.EndTime = endTime

	trafficRes, err := r.client.DescribeDomainTrafficData(trafficParams)
	if err != nil {
		r.logger.Errorf("[cdn] action DescribeDomainTrafficData failed, %s", err.Error())
	}

	// 指标收集
	for _, point := range trafficRes.TrafficDataPerInterval.DataModule {
		tm := RFC3339(point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain

		fields["value"] = ConvertToNum(point.Value)                           // 总流量
		fields["domesticValue"] = ConvertToNum(point.DomesticValue)           // 中国内地流量
		fields["overseasValue"] = ConvertToNum(point.OverseasValue)           // 全球（不包含中国内地）流量
		fields["httpsValue"] = ConvertToNum(point.HttpsValue)                 // L1节点https总流量
		fields["httpsDomesticValue"] = ConvertToNum(point.HttpsDomesticValue) // L1节点https中国内地流量
		fields["httpsOverseasValue"] = ConvertToNum(point.HttpsOverseasValue) // L1节点https全球（不包含中国内地）流量

		r.agent.accumulator.AddFields("domainTraffic", fields, tags, tm)
	}
}

func (r *RunningInstance) getHitRateData(domain, startTime, endTime string) {
	// HitRate
	hitRateParams := cdn.CreateDescribeDomainHitRateDataRequest()
	hitRateParams.Scheme = "https"
	hitRateParams.DomainName = domain
	hitRateParams.StartTime = startTime
	hitRateParams.EndTime = endTime

	hitRateRes, err := r.client.DescribeDomainHitRateData(hitRateParams)
	if err != nil {
		r.logger.Errorf("[cdn] action DescribeDomainTrafficData failed, %s", err.Error())
	}

	// 指标收集
	for _, point := range hitRateRes.HitRateInterval.DataModule {
		tm := RFC3339(point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain

		fields["value"] = ConvertToFloat(point.Value)
		fields["httpsValue"] = ConvertToFloat(point.HttpsValue)

		r.agent.accumulator.AddFields("domainHitRate", fields, tags, tm)
	}
}

func (r *RunningInstance) getReqHitRateData(domain, startTime, endTime string) {
	// ReqHitRate
	reqHitRateParams := cdn.CreateDescribeDomainReqHitRateDataRequest()
	reqHitRateParams.Scheme = "https"
	reqHitRateParams.DomainName = domain
	reqHitRateParams.StartTime = startTime
	reqHitRateParams.EndTime = endTime

	reqHitRateRes, err := r.client.DescribeDomainReqHitRateData(reqHitRateParams)
	if err != nil {
		r.logger.Errorf("[cdn] action DescribeDomainReqHitRateData failed, %s", err.Error())
	}

	// 收集数据
	for _, point := range reqHitRateRes.ReqHitRateInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain

		fields["value"] = ConvertToFloat(point.Value)
		fields["httpsValue"] = ConvertToFloat(point.HttpsValue)

		r.agent.accumulator.AddFields("reqHitRate", fields, tags, tm)
	}
}

func (r *RunningInstance) getDomainSrcBpsData(domain, startTime, endTime string) {
	// DomainSrcBps
	domainSrcBpsParams := cdn.CreateDescribeDomainSrcBpsDataRequest()
	domainSrcBpsParams.Scheme = "https"
	domainSrcBpsParams.DomainName = domain
	domainSrcBpsParams.StartTime = startTime
	domainSrcBpsParams.EndTime = endTime

	domainSrcBpsRes, err := r.client.DescribeDomainSrcBpsData(domainSrcBpsParams)
	if err != nil {
		fmt.Print(err.Error())
		r.logger.Errorf("[cdn] action DescribeDomainSrcBpsData failed, %s", err.Error())
	}

	// 收集数据
	for _, point := range domainSrcBpsRes.SrcBpsDataPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain

		fields["Value"] = ConvertToFloat(point.Value)
		fields["HttpsValue"] = ConvertToFloat(point.HttpsValue)

		r.agent.accumulator.AddFields("domainSrcBps", fields, tags, tm)
	}
}

func (r *RunningInstance) getDomainSrcTrafficData(domain, startTime, endTime string) {
	// DomainSrcBps
	domainSrcTrafficBpsParams := cdn.CreateDescribeDomainSrcTrafficDataRequest()
	domainSrcTrafficBpsParams.Scheme = "https"
	domainSrcTrafficBpsParams.DomainName = domain
	domainSrcTrafficBpsParams.StartTime = startTime
	domainSrcTrafficBpsParams.EndTime = endTime

	domainSrcTrafficBpsRes, err := r.client.DescribeDomainSrcTrafficData(domainSrcTrafficBpsParams)
	if err != nil {
		fmt.Print(err.Error())
		r.logger.Errorf("[cdn] action DescribeDomainSrcTrafficData failed, %s", err.Error())
	}

	// 收集数据
	for _, point := range domainSrcTrafficBpsRes.SrcTrafficDataPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain

		fields["value"] = ConvertToFloat(point.Value)
		fields["httpsValue"] = ConvertToFloat(point.HttpsValue)

		r.agent.accumulator.AddFields("domainSrcTraffic", fields, tags, tm)
	}
}

func (r *RunningInstance) cmd(action string) {
	var request interface{}
	var params []reflect.Value

	switch action {
	case "DescribeDomainBpsData":
		request = cdn.CreateDescribeDomainBpsDataRequest()
		reqParams := request.(*cdn.DescribeDomainBpsDataRequest)
		reqParams.Scheme = "https"
		params = []reflect.Value{reflect.ValueOf(reqParams)}
	case "DescribeDomainTraffic":
		request = cdn.CreateDescribeDomainTrafficDataRequest()
		reqParams := request.(*cdn.DescribeDomainTrafficDataRequest)
		reqParams.Scheme = "https"
		params = []reflect.Value{reflect.ValueOf(reqParams)}
	case "DescribeDomainHitRate":
		request = cdn.CreateDescribeDomainHitRateDataRequest()
		reqParams := request.(*cdn.DescribeDomainHitRateDataRequest)
		reqParams.Scheme = "https"
		params = []reflect.Value{reflect.ValueOf(reqParams)}
	case "DescribeDomainReqHitRate":
		request = cdn.CreateDescribeDomainReqHitRateDataRequest()
		reqParams := request.(*cdn.DescribeDomainReqHitRateDataRequest)
		reqParams.Scheme = "https"
		params = []reflect.Value{reflect.ValueOf(reqParams)}
	case "DescribeDomainSrcBps":
		request = cdn.CreateDescribeDomainSrcBpsDataRequest()
		reqParams := request.(*cdn.DescribeDomainSrcBpsDataRequest)
		reqParams.Scheme = "https"
		params = []reflect.Value{reflect.ValueOf(reqParams)}
	case "DescribeDomainSrcTraffic":
		request = cdn.CreateDescribeDomainSrcTrafficDataRequest()
		reqParams := request.(*cdn.DescribeDomainSrcTrafficDataRequest)
		reqParams.Scheme = "https"
		params = []reflect.Value{reflect.ValueOf(reqParams)}
	}

	fmt.Println("======req", params)
	cli := reflect.ValueOf(r.client)
	f := cli.MethodByName("DescribeDomainBpsData")
	fmt.Printf("++++++++func = %+v\n", &f)
	res := f.Call(params)
	fmt.Println("res======>", res)
}

func init() {
	inputs.Add("aliyuncdn", func() telegraf.Input {
		ac := &AliyunCDN{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
