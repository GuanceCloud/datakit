package aliyuncdn

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/telegraf"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

// project
type RunningProject struct {
	accumulator telegraf.Accumulator
	cfg         *Metric
	client      *cdn.Client
	domain      []string
	action      string
	metricName  string
	startTime   string
	endTime     string
}

// 调用DescribeDomainBpsData获取加速域名的网络带宽监控数据 (优化)
func (run *RunningProject) DescribeDomainBpsData() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainBpsData(domain)
	}
}

func (run *RunningProject) describeDomainBpsData(domain string) {
	// bps
	bpsParams := cdn.CreateDescribeDomainBpsDataRequest()
	bpsParams.Scheme = "https"
	bpsParams.DomainName = domain

	bpsParams.StartTime = run.startTime
	bpsParams.EndTime = run.endTime

	bpsParams.IspNameEn = run.cfg.IspNameEn
	bpsParams.LocationNameEn = run.cfg.LocationNameEn

	bpsRes, err := run.client.DescribeDomainBpsData(bpsParams)
	if err != nil {
		l.Errorf("[cdn] action DescribeDomainBpsData failed, %s", err.Error())
	}

	// 指标收集
	for _, point := range bpsRes.BpsDataPerInterval.DataModule {
		tm := RFC3339(point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainBpsData"
		tags["ispNameEn"] = run.cfg.IspNameEn
		tags["locationNameEn"] = run.cfg.LocationNameEn

		fields["value"] = ConvertToFloat(point.Value)                           // bps数据值，单位：bit/second。
		fields["domesticValue"] = ConvertToFloat(point.DomesticValue)           // 中国内地带宽bps
		fields["httpsDomesticValue"] = ConvertToFloat(point.HttpsDomesticValue) // L1节点https中国内地带宽
		fields["httpsOverseasValue"] = ConvertToFloat(point.HttpsOverseasValue) // L1节点全球（不包含中国内地）https带宽
		fields["httpsValue"] = ConvertToFloat(point.HttpsValue)                 // L1节点https的带宽数据值，单位：bit/second
		fields["overseasValue"] = ConvertToFloat(point.OverseasValue)           // 全球（不包含中国内地）带宽bps

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, tm)
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeDomainTrafficData获取加速域名的网络流量监控数据
func (run *RunningProject) DescribeDomainTrafficData() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainTrafficData(domain)
	}
}

func (run *RunningProject) describeDomainTrafficData(domain string) {
	// TrafficData
	trafficParams := cdn.CreateDescribeDomainTrafficDataRequest()
	trafficParams.Scheme = "https"
	trafficParams.DomainName = domain

	trafficParams.StartTime = run.startTime
	trafficParams.EndTime = run.endTime

	trafficParams.IspNameEn = run.cfg.IspNameEn
	trafficParams.LocationNameEn = run.cfg.LocationNameEn

	trafficRes, err := run.client.DescribeDomainTrafficData(trafficParams)
	if err != nil {
		l.Errorf("[cdn] action DescribeDomainTrafficData failed, %s", err.Error())
	}

	// 指标收集
	for _, point := range trafficRes.TrafficDataPerInterval.DataModule {
		tm := RFC3339(point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "DescribeDomainTrafficData"
		tags["ispNameEn"] = run.cfg.IspNameEn
		tags["locationNameEn"] = run.cfg.LocationNameEn

		fields["value"] = ConvertToFloat(point.Value)                           // 总流量
		fields["domesticValue"] = ConvertToFloat(point.DomesticValue)           // 中国内地流量
		fields["overseasValue"] = ConvertToFloat(point.OverseasValue)           // 全球（不包含中国内地）流量
		fields["httpsValue"] = ConvertToFloat(point.HttpsValue)                 // L1节点https总流量
		fields["httpsDomesticValue"] = ConvertToFloat(point.HttpsDomesticValue) // L1节点https中国内地流量
		fields["httpsOverseasValue"] = ConvertToFloat(point.HttpsOverseasValue) // L1节点https全球（不包含中国内地）流量

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, tm)
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeDomainHitRateData获取加速域名的字节命中率（命中字节百分比）
func (run *RunningProject) DescribeDomainHitRateData() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainHitRateData(domain)
	}
}

func (run *RunningProject) describeDomainHitRateData(domain string) {
	// HitRate
	hitRateParams := cdn.CreateDescribeDomainHitRateDataRequest()
	hitRateParams.Scheme = "https"
	hitRateParams.DomainName = domain
	hitRateParams.StartTime = run.startTime
	hitRateParams.EndTime = run.endTime

	hitRateRes, err := run.client.DescribeDomainHitRateData(hitRateParams)
	if err != nil {
		l.Errorf("[cdn] action DescribeDomainTrafficData failed, %s", err.Error())
	}

	// 指标收集
	for _, point := range hitRateRes.HitRateInterval.DataModule {
		tm := RFC3339(point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainHitRateData"

		fields["value"] = ConvertToFloat(point.Value)
		fields["httpsValue"] = ConvertToFloat(point.HttpsValue)

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, tm)
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeDomainReqHitRateData获取加速域名的请求命中率（命中请求百分比）
func (run *RunningProject) DescribeDomainReqHitRateData() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainReqHitRateData(domain)
	}
}

func (run *RunningProject) describeDomainReqHitRateData(domain string) {
	// ReqHitRate
	reqHitRateParams := cdn.CreateDescribeDomainReqHitRateDataRequest()
	reqHitRateParams.Scheme = "https"
	reqHitRateParams.DomainName = domain
	reqHitRateParams.StartTime = run.startTime
	reqHitRateParams.EndTime = run.endTime

	reqHitRateRes, err := run.client.DescribeDomainReqHitRateData(reqHitRateParams)
	if err != nil {
		l.Errorf("[cdn] action DescribeDomainReqHitRateData failed, %s", err.Error())
	}

	// 收集数据
	for _, point := range reqHitRateRes.ReqHitRateInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainReqHitRateData"

		fields["value"] = ConvertToFloat(point.Value)
		fields["httpsValue"] = ConvertToFloat(point.HttpsValue)

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, tm)
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeDomainSrcBpsData获取加速域名的回源带宽监控数据
func (run *RunningProject) DescribeDomainSrcBpsData() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainSrcBpsData(domain)
	}
}

func (run *RunningProject) describeDomainSrcBpsData(domain string) {
	// DomainSrcBps
	domainSrcBpsParams := cdn.CreateDescribeDomainSrcBpsDataRequest()
	domainSrcBpsParams.Scheme = "https"
	domainSrcBpsParams.DomainName = domain
	domainSrcBpsParams.StartTime = run.startTime
	domainSrcBpsParams.EndTime = run.endTime

	domainSrcBpsRes, err := run.client.DescribeDomainSrcBpsData(domainSrcBpsParams)
	if err != nil {
		l.Errorf("[cdn] action DescribeDomainSrcBpsData failed, %s", err.Error())
	}

	// 收集数据
	for _, point := range domainSrcBpsRes.SrcBpsDataPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainSrcBpsData"

		fields["Value"] = ConvertToFloat(point.Value)
		fields["HttpsValue"] = ConvertToFloat(point.HttpsValue)

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, tm)
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeDomainSrcTrafficData获取加速域名的回源流量监控数据
func (run *RunningProject) DescribeDomainSrcTrafficData() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainSrcTrafficData(domain)
	}
}

func (run *RunningProject) describeDomainSrcTrafficData(domain string) {
	// DomainSrcBps
	domainSrcTrafficBpsParams := cdn.CreateDescribeDomainSrcTrafficDataRequest()
	domainSrcTrafficBpsParams.Scheme = "https"
	domainSrcTrafficBpsParams.DomainName = domain
	domainSrcTrafficBpsParams.StartTime = run.startTime
	domainSrcTrafficBpsParams.EndTime = run.endTime

	domainSrcTrafficBpsRes, err := run.client.DescribeDomainSrcTrafficData(domainSrcTrafficBpsParams)
	if err != nil {
		l.Errorf("[cdn] action DescribeDomainSrcTrafficData failed, %s", err.Error())
	}

	// 收集数据
	for _, point := range domainSrcTrafficBpsRes.SrcTrafficDataPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainSrcTrafficData"

		fields["value"] = ConvertToFloat(point.Value)
		fields["httpsValue"] = ConvertToFloat(point.HttpsValue)

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, tm)
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeDomainRegionData获取加速域名天粒度的用户区域分布数据统计
func (run *RunningProject) DescribeDomainUvData() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainUvData(domain)
	}
}
func (run *RunningProject) describeDomainUvData(domain string) {
	uvParams := cdn.CreateDescribeDomainUvDataRequest()
	uvParams.Scheme = "https"

	uvParams.DomainName = domain
	uvParams.StartTime = run.startTime
	uvParams.EndTime = run.endTime

	uvRes, err := run.client.DescribeDomainUvData(uvParams)
	if err != nil {
		l.Warnf("action:DescribeDomainUvData, error: %s", err.Error())
	}

	for _, point := range uvRes.UvDataInterval.UsageData {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)

		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainUvData"

		fields["Value"] = point.Value

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, tm)
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeDomainPvData获取加速域名1小时粒度的PV页面访问统计
func (run *RunningProject) DescribeDomainPvData() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainPvData(domain)
	}
}

func (run *RunningProject) describeDomainPvData(domain string) {
	pvParams := cdn.CreateDescribeDomainPvDataRequest()
	pvParams.Scheme = "https"

	pvParams.DomainName = domain
	pvParams.StartTime = run.startTime
	pvParams.EndTime = run.endTime

	pvRes, err := run.client.DescribeDomainPvData(pvParams)
	if err != nil {
		l.Warnf("action:DescribeDomainPvData, error: %s", err.Error())
	}

	for _, point := range pvRes.PvDataInterval.UsageData {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)

		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainPvData"

		fields["Value"] = point.Value

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, tm)
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeDomainTopClientIpVisit获取加速域名在指定时间范围内按照访问次数或流量排序的Client IP排行
func (run *RunningProject) DescribeDomainTopClientIpVisit() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainTopClientIpVisit(domain)
	}
}
func (run *RunningProject) describeDomainTopClientIpVisit(domain string) {
	clientIpVisitParams := cdn.CreateDescribeDomainTopClientIpVisitRequest()
	clientIpVisitParams.Scheme = "https"

	clientIpVisitParams.DomainName = domain
	clientIpVisitParams.SortBy = run.cfg.SortBy
	clientIpVisitParams.StartTime = run.startTime
	clientIpVisitParams.EndTime = run.endTime

	clientIpVisitParams.LocationNameEn = run.cfg.LocationNameEn
	clientIpVisitParams.Limit = fmt.Sprintf("%d", 100)

	clientIpVisitRes, err := run.client.DescribeDomainTopClientIpVisit(clientIpVisitParams)
	if err != nil {
		l.Warnf("action:DescribeDomainTopClientIpVisit, error: %s", err.Error())
	}

	for _, point := range clientIpVisitRes.ClientIpList {
		tags := map[string]string{}
		fields := map[string]interface{}{}
		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainTopClientIpVisit"
		tags["locationNameEn"] = run.cfg.LocationNameEn

		fields["clientIp"] = point.ClientIp
		fields["rank"] = point.Rank
		fields["traffic"] = point.Traffic
		fields["acc"] = point.Acc

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, time.Now())
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeDomainBpsDataByTimeStamp获取加速域名在某个时刻的带宽数据
func (run *RunningProject) DescribeDomainISPData() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainISPData(domain)
	}
}
func (run *RunningProject) describeDomainISPData(domain string) {
	ispParams := cdn.CreateDescribeDomainISPDataRequest()
	ispParams.Scheme = "https"
	ispParams.DomainName = domain
	ispParams.StartTime = run.startTime
	ispParams.EndTime = run.endTime

	ispRes, err := run.client.DescribeDomainISPData(ispParams)
	if err != nil {
		l.Warnf("action:DescribeDomainISPData, error: %s", err.Error())
	}

	for _, point := range ispRes.Value.ISPProportionData {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainISPData"

		fields["proportion"] = ConvertToFloat(point.Proportion)
		fields["avgObjectSize"] = ConvertToFloat(point.AvgObjectSize)
		fields["avgResponseTime"] = ConvertToFloat(point.AvgResponseTime)
		fields["bps"] = ConvertToFloat(point.Bps)
		fields["qps"] = ConvertToFloat(point.Qps)
		fields["avgResponseRate"] = ConvertToFloat(point.AvgResponseRate)
		fields["totalBytes"] = ConvertToFloat(point.TotalBytes)
		fields["bytesProportion"] = ConvertToFloat(point.BytesProportion)
		fields["totalQuery"] = ConvertToFloat(point.TotalQuery)

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, time.Now())
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeDomainTopUrlVisit获取加速域名某天内的热门URL列表
func (run *RunningProject) DescribeDomainTopUrlVisit() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainTopUrlVisit(domain)
	}
}
func (run *RunningProject) describeDomainTopUrlVisit(domain string) {
	request := cdn.CreateDescribeDomainTopUrlVisitRequest()
	request.Scheme = "https"
	request.DomainName = domain
	request.SortBy = run.cfg.SortBy
	request.StartTime = run.startTime
	request.EndTime = run.endTime

	response, err := run.client.DescribeDomainTopUrlVisit(request)
	if err != nil {
		l.Warnf("action:DescribeDomainTopUrlVisit, error: %s", err.Error())
	}

	for _, point := range response.AllUrlList.UrlList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainTopUrlVisit"
		tags["code"] = "all"

		fields["flow"] = ConvertToFloat(point.Flow)
		fields["visitProportion"] = point.VisitProportion
		fields["visitData"] = ConvertToFloat(point.VisitData)
		fields["urlDetail"] = ConvertToFloat(point.UrlDetail)
		fields["flowProportion"] = point.FlowProportion

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, time.Now())
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}

	for _, point := range response.Url200List.UrlList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainSrcTopUrlVisit"
		tags["code"] = "200"

		fields["visitData"] = ConvertToFloat(point.VisitData)
		fields["urlDetail"] = point.UrlDetail
		fields["visitProportion"] = point.VisitProportion
		fields["flow"] = ConvertToFloat(point.Flow)
		fields["flowProportion"] = point.FlowProportion

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, time.Now())
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}

	for _, point := range response.Url300List.UrlList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainSrcTopUrlVisit"
		tags["code"] = "300"

		fields["visitData"] = ConvertToFloat(point.VisitData)
		fields["urlDetail"] = point.UrlDetail
		fields["visitProportion"] = point.VisitProportion
		fields["flow"] = ConvertToFloat(point.Flow)
		fields["flowProportion"] = point.FlowProportion

		run.accumulator.AddFields(run.metricName, fields, tags)
	}

	for _, point := range response.Url400List.UrlList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainSrcTopUrlVisit"
		tags["code"] = "400"

		fields["visitData"] = ConvertToFloat(point.VisitData)
		fields["urlDetail"] = point.UrlDetail
		fields["visitProportion"] = point.VisitProportion
		fields["flow"] = ConvertToFloat(point.Flow)
		fields["flowProportion"] = point.FlowProportion

		run.accumulator.AddFields(run.metricName, fields, tags)
	}

	for _, point := range response.Url500List.UrlList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainSrcTopUrlVisit"
		tags["code"] = "500"

		fields["visitData"] = ConvertToFloat(point.VisitData)
		fields["urlDetail"] = point.UrlDetail
		fields["visitProportion"] = point.VisitProportion
		fields["flow"] = ConvertToFloat(point.Flow)
		fields["flowProportion"] = point.FlowProportion

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, time.Now())
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeDomainSrcTopUrlVisit获取加速域名5分钟计算粒度的回源热门Url
func (run *RunningProject) DescribeDomainSrcTopUrlVisit() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainSrcTopUrlVisit(domain)
	}
}
func (run *RunningProject) describeDomainSrcTopUrlVisit(domain string) {
	fmt.Println("[DescribeDomainSrcTopUrlVisit] start...")
	request := cdn.CreateDescribeDomainSrcTopUrlVisitRequest()
	request.Scheme = "https"

	request.DomainName = domain
	request.SortBy = run.cfg.SortBy

	request.StartTime = run.startTime
	request.EndTime = run.endTime
	response, err := run.client.DescribeDomainSrcTopUrlVisit(request)
	if err != nil {
		l.Warnf("action:DescribeDomainSrcTopUrlVisit, error: %s", err.Error())
	}
	for _, point := range response.AllUrlList.UrlList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainSrcTopUrlVisit"
		tags["code"] = "all"

		fields["visitData"] = ConvertToFloat(point.VisitData)
		fields["urlDetail"] = point.UrlDetail
		fields["visitProportion"] = point.VisitProportion
		fields["flow"] = ConvertToFloat(point.Flow)
		fields["flowProportion"] = point.FlowProportion

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, time.Now())
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}

	for _, point := range response.Url200List.UrlList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainSrcTopUrlVisit"
		tags["code"] = "200"

		fields["visitData"] = ConvertToFloat(point.VisitData)
		fields["urlDetail"] = point.UrlDetail
		fields["visitProportion"] = point.VisitProportion
		fields["flow"] = ConvertToFloat(point.Flow)
		fields["flowProportion"] = point.FlowProportion

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, time.Now())
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}

	for _, point := range response.Url300List.UrlList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainSrcTopUrlVisit"
		tags["code"] = "300"

		fields["visitData"] = ConvertToFloat(point.VisitData)
		fields["urlDetail"] = point.UrlDetail
		fields["visitProportion"] = point.VisitProportion
		fields["flow"] = ConvertToFloat(point.Flow)
		fields["flowProportion"] = point.FlowProportion

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, time.Now())
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}

	for _, point := range response.Url400List.UrlList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainSrcTopUrlVisit"
		tags["code"] = "400"

		fields["visitData"] = ConvertToFloat(point.VisitData)
		fields["urlDetail"] = point.UrlDetail
		fields["visitProportion"] = point.VisitProportion
		fields["flow"] = ConvertToFloat(point.Flow)
		fields["flowProportion"] = point.FlowProportion

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, time.Now())
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}

	for _, point := range response.Url500List.UrlList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		tags["domain"] = domain
		tags["action"] = "describeDomainSrcTopUrlVisit"
		tags["code"] = "500"

		fields["visitData"] = ConvertToFloat(point.VisitData)
		fields["urlDetail"] = point.UrlDetail
		fields["visitProportion"] = point.VisitProportion
		fields["flow"] = ConvertToFloat(point.Flow)
		fields["flowProportion"] = point.FlowProportion

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, time.Now())
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeTopDomainsByFlow获取用户按流量排名的域名
func (run *RunningProject) DescribeTopDomainsByFlow() {
	request := cdn.CreateDescribeTopDomainsByFlowRequest()
	request.Scheme = "https"
	request.StartTime = run.startTime
	request.EndTime = run.endTime
	request.Limit = requests.NewInteger(100)

	response, err := run.client.DescribeTopDomainsByFlow(request)
	if err != nil {
		l.Warnf("action:DescribeTopDomainsByFlow, error: %s", err.Error())
	}

	for _, point := range response.TopDomains.TopDomain {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		fields["domain"] = point.DomainName
		tags["action"] = "describeDomainTopUrlVisit"

		fields["maxBps"] = point.MaxBps
		fields["rank"] = point.Rank
		fields["trafficPercent"] = ConvertToFloat(point.TrafficPercent)
		fields["totalTraffic"] = ConvertToFloat(point.TotalTraffic)
		fields["totalAccess"] = point.TotalAccess
		fields["maxBpsTime"] = RFC3339(point.MaxBpsTime)

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, time.Now())
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// 调用DescribeDomainSrcQpsData获取加速域名的回源请求QPS数据
func (run *RunningProject) DescribeDomainTopReferVisit() {
	// bps
	for _, domain := range run.domain {
		go run.describeDomainTopReferVisit(domain)
	}
}
func (run *RunningProject) describeDomainTopReferVisit(domain string) {
	fmt.Println("[DescribeDomainTopReferVisit] start...")
	request := cdn.CreateDescribeDomainTopReferVisitRequest()
	request.Scheme = "https"

	request.DomainName = domain
	request.SortBy = run.cfg.SortBy
	request.StartTime = run.startTime
	request.EndTime = run.endTime

	response, err := run.client.DescribeDomainTopReferVisit(request)
	if err != nil {
		l.Warnf("action:DescribeDomainTopReferVisit, error: %s", err.Error())
	}
	for _, point := range response.TopReferList.ReferList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "cdn"
		fields["domain"] = domain
		tags["action"] = "describeDomainTopReferVisit"

		fields["flow"] = ConvertToFloat(point.Flow)
		fields["visitProportion"] = point.VisitProportion
		fields["visitData"] = ConvertToFloat(point.VisitData)
		fields["referDetail"] = point.ReferDetail
		fields["flowProportion"] = point.FlowProportion

		pt, err := influxdb.NewPoint(run.metricName, tags, fields, time.Now())
		if err != nil {
			return
		}

		err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
	}
}

// // 调用DescribeDomainSrcHttpCodeData获取加速域名5分钟计算粒度的回源HTTP返回码的总数和占比数据
// func (run *RunningProject) DescribeDomainSrcHttpCodeData() {
// 	request := cdn.CreateDescribeDomainSrcHttpCodeDataRequest()
// 	request.Scheme = "https"
// 	request.DomainName = run.cfg.DomainName
// 	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	response, err := run.Client.DescribeDomainSrcHttpCodeData(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainSrcHttpCodeData, error: %s", err.Error())
// 		fmt.Println("error ===>", err)
// 	}

// 	for _, data := range response.HttpCodeData.UsageData {
// 		for _, point := range data.Value.CodeProportionData {
// 			tags := map[string]string{}
// 			fields := map[string]interface{}{}

// 			tags["Code"] = point.Code

// 			fields["Proportion"] = point.Proportion
// 			fields["Count"] = point.Count

// 			run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
// 		}
// 	}
// }

// // 调用DescribeDomainQpsData获取5分钟计算粒度加速域名的每秒访问次数QPS
// func (run *RunningProject) DescribeDomainQpsData() {
// 	fmt.Println("[DescribeDomainQpsData] start...")
// 	request := cdn.CreateDescribeDomainQpsDataRequest()
// 	request.Scheme = "https"
// 	request.DomainName = run.cfg.DomainName
// 	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	request.IspNameEn = run.cfg.IspNameEn
// 	request.LocationNameEn = run.cfg.LocationNameEn

// 	response, err := run.Client.DescribeDomainQpsData(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainQpsData, error: %s", err.Error())
// 		fmt.Println("error ===>", err)
// 	}

// 	for _, point := range response.QpsDataInterval.DataModule {
// 		const layout = time.RFC3339
// 		tm, _ := time.Parse(layout, point.TimeStamp)

// 		tags := map[string]string{}
// 		fields := map[string]interface{}{}

// 		tags["DomainName"] = response.DomainName
// 		fields["HttpsOverseasValue"] = point.HttpsOverseasValue
// 		fields["HttpsAccOverseasValue"] = point.HttpsAccOverseasValue
// 		fields["Value"] = point.Value
// 		fields["OverseasValue"] = point.OverseasValue
// 		fields["AccOverseasValue"] = point.AccOverseasValue
// 		fields["AccValue"] = point.AccValue
// 		fields["HttpsAccValue"] = point.HttpsAccValue
// 		fields["DomesticValue"] = point.DomesticValue
// 		fields["HttpsAccDomesticValue"] = point.HttpsAccDomesticValue
// 		fields["HttpsValue"] = point.HttpsValue

// 		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
// 	}

// }

// // 调用DescribeDomainQpsDataByLayer获取加速域名的每秒访问次数QPS
// func (run *RunningProject) DescribeDomainQpsDataByLayer() {
// 	fmt.Println("[DescribeDomainQpsData] start...")
// 	request := cdn.CreateDescribeDomainQpsDataByLayerRequest()
// 	request.Scheme = "https"
// 	request.DomainName = run.cfg.DomainName
// 	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	request.IspNameEn = run.cfg.IspNameEn
// 	request.LocationNameEn = run.cfg.LocationNameEn
// 	request.Layer = run.cfg.Layer

// 	response, err := run.Client.DescribeDomainQpsDataByLayer(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainQpsDataByLayer, error: %s", err.Error())
// 	}
// 	for _, point := range response.QpsDataInterval.DataModule {
// 		const layout = time.RFC3339
// 		tm, _ := time.Parse(layout, point.TimeStamp)

// 		tags := map[string]string{}
// 		fields := map[string]interface{}{}

// 		tags["DomainName"] = response.DomainName

// 		fields["AccDomesticValue"] = point.AccDomesticValue
// 		fields["AccOverseasValue"] = point.AccOverseasValue
// 		fields["AccValue"] = point.AccValue
// 		fields["DomesticValue"] = point.DomesticValue
// 		fields["OverseasValue"] = point.OverseasValue
// 		fields["AccValue"] = point.AccValue
// 		fields["Value"] = point.HttpsAccValue

// 		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
// 	}
// }

// // 调用DescribeDomainSrcQpsData获取加速域名的回源请求QPS数据
// func (run *RunningProject) DescribeDomainSrcQpsData() {
// 	fmt.Println("[DescribeDomainQpsData] start...")
// 	request := cdn.CreateDescribeDomainSrcQpsDataRequest()
// 	request.Scheme = "https"

// 	request.DomainName = run.cfg.DomainName
// 	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")
// 	fmt.Println("request ======>", request)
// 	response, err := run.Client.DescribeDomainSrcQpsData(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainSrcQpsData, error: %s", err.Error())
// 	}
// 	fmt.Println("response ======>", response.SrcQpsDataPerInterval.DataModule)
// 	for _, point := range response.SrcQpsDataPerInterval.DataModule {
// 		const layout = time.RFC3339
// 		tm, _ := time.Parse(layout, point.TimeStamp)

// 		tags := map[string]string{}
// 		fields := map[string]interface{}{}

// 		tags["DomainName"] = response.DomainName

// 		fields["Value"] = point.HttpsAccValue
// 		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
// 	}
// }

// // 调DescribeDomainAverageResponseTime获取加速域名的平均响应时间
// func (run *RunningProject) DescribeDomainAverageResponseTime() {
// 	request := cdn.CreateDescribeDomainAverageResponseTimeRequest()
// 	request.Scheme = "https"

// 	request.DomainName = run.cfg.DomainName
// 	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	request.IspNameEn = run.cfg.IspNameEn
// 	request.LocationNameEn = run.cfg.LocationNameEn
// 	request.DomainType = run.cfg.DomainType

// 	response, err := run.Client.DescribeDomainAverageResponseTime(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainAverageResponseTime, error: %s", err.Error())
// 	}

// 	for _, point := range response.AvgRTPerInterval.DataModule {
// 		const layout = time.RFC3339
// 		tm, _ := time.Parse(layout, point.TimeStamp)

// 		tags := map[string]string{}
// 		fields := map[string]interface{}{}

// 		tags["DomainName"] = response.DomainName

// 		fields["Value"] = point.Value

// 		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
// 	}
// }

// // 调DescribeDomainAverageResponseTime获取加速域名的平均响应时间
// func (run *RunningProject) DescribeDomainFileSizeProportionData() {
// 	request := cdn.CreateDescribeDomainFileSizeProportionDataRequest()
// 	request.Scheme = "https"

// 	request.DomainName = run.cfg.DomainName

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	response, err := run.Client.DescribeDomainFileSizeProportionData(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainFileSizeProportionData, error: %s", err.Error())
// 	}

// 	for _, data := range response.FileSizeProportionDataInterval.UsageData {
// 		for _, point := range data.Value.FileSizeProportionData {
// 			tags := map[string]string{}
// 			fields := map[string]interface{}{}

// 			tags["DomainName"] = response.DomainName

// 			fields["FileSize"] = point.FileSize
// 			fields["Proportion"] = point.Proportion

// 			run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
// 		}
// 	}
// }

// // 调用DescribeDomainBpsDataByTimeStamp获取加速域名在某个时刻的带宽数据
// func (run *RunningProject) DescribeDomainBpsDataByTimeStamp() {
// 	request := cdn.CreateDescribeDomainBpsDataByTimeStampRequest()
// 	request.Scheme = "https"

// 	request.DomainName = run.cfg.DomainName
// 	nt := time.Now()
// 	et := nt.Unix()

// 	request.TimePoint = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	request.IspNames = run.cfg.IspNames
// 	request.LocationNames = run.cfg.LocationNames

// 	response, err := run.Client.DescribeDomainBpsDataByTimeStamp(request)
// 	if err != nil {
// 		fmt.Print(err.Error())
// 	}

// 	for _, point := range response.BpsDataList.BpsDataModel {
// 		tags := map[string]string{}
// 		fields := map[string]interface{}{}

// 		tags["DomainName"] = response.DomainName
// 		tags["LocationName"] = point.LocationName
// 		tags["IspName"] = point.IspName

// 		fields["Bps"] = point.Bps

// 		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
// 	}
// }

// // 调用DescribeDomainRealTimeBpsData获取加速域名1分钟粒度带宽数据，支持查询7天内的数据
// func (run *RunningProject) DescribeDomainRealTimeBpsData() {
// 	request := cdn.CreateDescribeDomainRealTimeBpsDataRequest()
// 	request.Scheme = "https"

// 	request.DomainName = run.cfg.DomainName
// 	request.LocationNameEn = run.cfg.LocationNameEn
// 	request.IspNameEn = run.cfg.IspNameEn

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	response, err := run.Client.DescribeDomainRealTimeBpsData(request)
// 	if err != nil {
// 		fmt.Print(err.Error())
// 		l.Warnf("action:DescribeDomainRealTimeBpsData, error: %s", err.Error())
// 	}

// 	for _, point := range response.Data.BpsModel {
// 		const layout = time.RFC3339
// 		tm, _ := time.Parse(layout, point.TimeStamp)

// 		tags := map[string]string{}
// 		fields := map[string]interface{}{}

// 		fields["Bps"] = point.Bps

// 		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
// 	}
// }

// // 调用DescribeDomainRealTimeSrcBpsData获取域名1分钟粒度回源带宽数据，支持获取最近7天的数据
// func (run *RunningProject) DescribeDomainRealTimeSrcBpsData() {
// 	request := cdn.CreateDescribeDomainRealTimeSrcBpsDataRequest()
// 	request.Scheme = "https"

// 	request.DomainName = run.cfg.DomainName

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	response, err := run.Client.DescribeDomainRealTimeSrcBpsData(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainRealTimeSrcBpsData, error: %s", err.Error())
// 	}

// 	for _, point := range response.RealTimeSrcBpsDataPerInterval.DataModule {
// 		const layout = time.RFC3339
// 		tm, _ := time.Parse(layout, point.TimeStamp)
// 		tags := map[string]string{}
// 		fields := map[string]interface{}{}

// 		tags["DomainName"] = response.DomainName

// 		fields["Value"] = point.Value

// 		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
// 	}
// }

// // 调用DescribeDomainRealTimeSrcHttpCodeData获取加速域名回源1分钟粒度HTTP返回码的总数和占比数据，支持获取最近7天的数据。
// func (run *RunningProject) DescribeDomainRealTimeSrcHttpCodeData() {
// 	request := cdn.CreateDescribeDomainRealTimeSrcHttpCodeDataRequest()
// 	request.Scheme = "https"
// 	request.DomainName = run.cfg.DomainName

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	request.IspNameEn = run.cfg.IspNameEn
// 	request.LocationNameEn = run.cfg.LocationNameEn

// 	response, err := run.Client.DescribeDomainRealTimeSrcHttpCodeData(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainQpsData, error: %s", err.Error())
// 	}

// 	for _, data := range response.RealTimeSrcHttpCodeData.UsageData {
// 		for _, point := range data.Value.RealTimeSrcCodeProportionData {
// 			tags := map[string]string{}
// 			fields := map[string]interface{}{}

// 			tags["DomainName"] = data.TimeStamp
// 			fields["Proportion"] = point.Proportion
// 			fields["Code"] = point.Code

// 			run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
// 		}
// 	}
// }

// // 调用DescribeDomainRealTimeSrcTrafficData获取加速域名的1分钟回源流量监控数据，支持获取最近90天的数据，数据单位：byte。
// func (run *RunningProject) DescribeDomainRealTimeSrcTrafficData() {
// 	request := cdn.CreateDescribeDomainRealTimeSrcTrafficDataRequest()
// 	request.Scheme = "https"

// 	request.DomainName = run.cfg.DomainName

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	response, err := run.Client.DescribeDomainRealTimeSrcTrafficData(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainRealTimeSrcTrafficData, error: %s", err.Error())
// 	}

// 	for _, point := range response.RealTimeSrcTrafficDataPerInterval.DataModule {
// 		const layout = time.RFC3339
// 		tm, _ := time.Parse(layout, point.TimeStamp)
// 		tags := map[string]string{}
// 		fields := map[string]interface{}{}

// 		tags["DomainName"] = response.DomainName

// 		fields["Value"] = point.Value

// 		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
// 	}
// }

// // 调用DescribeDomainRealTimeByteHitRateData获取域名1分钟粒度字节命中率数据
// func (run *RunningProject) DescribeDomainRealTimeByteHitRateData() {
// 	request := cdn.CreateDescribeDomainRealTimeByteHitRateDataRequest()
// 	request.Scheme = "https"

// 	request.DomainName = run.cfg.DomainName

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	response, err := run.Client.DescribeDomainRealTimeByteHitRateData(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainRealTimeByteHitRateData, error: %s", err.Error())
// 	}

// 	for _, point := range response.Data.ByteHitRateDataModel {
// 		const layout = time.RFC3339
// 		tm, _ := time.Parse(layout, point.TimeStamp)
// 		tags := map[string]string{}
// 		fields := map[string]interface{}{}

// 		fields["ByteHitRate"] = point.ByteHitRate

// 		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
// 	}
// }

// // 调用DescribeDomainRealTimeQpsData获取加速域名1分钟粒度每秒访问次数数据
// func (run *RunningProject) DescribeDomainRealTimeQpsData() {
// 	request := cdn.CreateDescribeDomainRealTimeQpsDataRequest()
// 	request.Scheme = "https"

// 	request.DomainName = run.cfg.DomainName

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	request.IspNameEn = run.cfg.IspNameEn
// 	request.LocationNameEn = run.cfg.LocationNameEn

// 	response, err := run.Client.DescribeDomainRealTimeQpsData(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainRealTimeQpsData, error: %s", err.Error())
// 	}

// 	for _, point := range response.Data.QpsModel {
// 		const layout = time.RFC3339
// 		tm, _ := time.Parse(layout, point.TimeStamp)
// 		tags := map[string]string{}
// 		fields := map[string]interface{}{}

// 		fields["Qps"] = point.Qps

// 		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
// 	}
// }

// // 调用DescribeDomainRealTimeTrafficData获取加速域名的1分钟粒度流量监控数据
// func (run *RunningProject) DescribeDomainRealTimeTrafficData() {
// 	request := cdn.CreateDescribeDomainRealTimeTrafficDataRequest()
// 	request.Scheme = "https"

// 	request.DomainName = run.cfg.DomainName

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	response, err := run.Client.DescribeDomainRealTimeTrafficData(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainRealTimeTrafficData, error: %s", err.Error())
// 	}

// 	for _, point := range response.RealTimeTrafficDataPerInterval.DataModule {
// 		const layout = time.RFC3339
// 		tm, _ := time.Parse(layout, point.TimeStamp)
// 		tags := map[string]string{}
// 		fields := map[string]interface{}{}

// 		fields["Value"] = point.Value

// 		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
// 	}
// }

// // 调用DescribeDomainRealTimeReqHitRateData获取加速域名1分钟粒度请求命中率数据
// func (run *RunningProject) DescribeDomainRealTimeReqHitRateData() {
// 	request := cdn.CreateDescribeDomainRealTimeReqHitRateDataRequest()
// 	request.Scheme = "https"

// 	request.DomainName = run.cfg.DomainName

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	response, err := run.Client.DescribeDomainRealTimeReqHitRateData(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainRealTimeReqHitRateData, error: %s", err.Error())
// 	}

// 	for _, point := range response.Data.ReqHitRateDataModel {
// 		const layout = time.RFC3339
// 		tm, _ := time.Parse(layout, point.TimeStamp)
// 		tags := map[string]string{}
// 		fields := map[string]interface{}{}

// 		fields["ReqHitRate"] = point.ReqHitRate

// 		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
// 	}
// }

// // 调用DescribeDomainHttpCodeData获取加速域名HTTP返回码的总数和占比数据
// func (run *RunningProject) DescribeDomainHttpCodeData() {
// 	request := cdn.CreateDescribeDomainHttpCodeDataRequest()
// 	request.Scheme = "https"

// 	request.DomainName = run.cfg.DomainName

// 	nt := time.Now()
// 	et := nt.Unix()
// 	st := nt.Add(-(5 * time.Minute)).Unix()

// 	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
// 	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

// 	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

// 	response, err := run.Client.DescribeDomainHttpCodeData(request)
// 	if err != nil {
// 		l.Warnf("action:DescribeDomainHttpCodeData, error: %s", err.Error())
// 	}

// 	for _, data := range response.HttpCodeData.UsageData {
// 		for _, point := range data.Value.CodeProportionData {
// 			tags := map[string]string{}
// 			fields := map[string]interface{}{}

// 			tags["DomainName"] = response.DomainName

// 			tags["Code"] = point.Code

// 			fields["Proportion"] = point.Proportion

// 			run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
// 		}
// 	}
// }
