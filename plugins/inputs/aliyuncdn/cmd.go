package aliyuncdn

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
)

// project
type RunningProject struct {
	inst    *RunningInstance
	cfg     *Action
	Client  *cdn.Client
	logger  *models.Logger
	mainCfg *CDN
}

// 调用DescribeDomainSrcHttpCodeData获取加速域名5分钟计算粒度的回源HTTP返回码的总数和占比数据
func (run *RunningProject) DescribeDomainSrcHttpCodeData() {
	fmt.Println("[DescribeDomainSrcHttpCodeData] start...")
	request := cdn.CreateDescribeDomainSrcHttpCodeDataRequest()
	request.Scheme = "https"
	request.DomainName = run.cfg.DomainName
	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	response, err := run.Client.DescribeDomainSrcHttpCodeData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainSrcHttpCodeData, error: %s", err.Error())
		fmt.Println("error ===>", err)
	}

	for _, data := range response.HttpCodeData.UsageData {
		for _, point := range data.Value.CodeProportionData {
			tags := map[string]string{}
			fields := map[string]interface{}{}

			tags["Code"] = point.Code
			tags["TimeStamp"] = data.TimeStamp

			fields["Proportion"] = point.Proportion
			fields["Count"] = point.Count

			run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
		}
	}
}

// 调用DescribeDomainQpsData获取5分钟计算粒度加速域名的每秒访问次数QPS
func (run *RunningProject) DescribeDomainQpsData() {
	fmt.Println("[DescribeDomainQpsData] start...")
	request := cdn.CreateDescribeDomainQpsDataRequest()
	request.Scheme = "https"
	request.DomainName = run.cfg.DomainName
	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.IspNameEn = run.cfg.IspNameEn
	request.LocationNameEn = run.cfg.LocationNameEn

	response, err := run.Client.DescribeDomainQpsData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainQpsData, error: %s", err.Error())
		fmt.Println("error ===>", err)
	}

	for _, point := range response.QpsDataInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)

		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName
		fields["HttpsOverseasValue"] = point.HttpsOverseasValue
		fields["HttpsAccOverseasValue"] = point.HttpsAccOverseasValue
		fields["Value"] = point.Value
		fields["OverseasValue"] = point.OverseasValue
		fields["AccOverseasValue"] = point.AccOverseasValue
		fields["AccValue"] = point.AccValue
		fields["HttpsAccValue"] = point.HttpsAccValue
		fields["DomesticValue"] = point.DomesticValue
		fields["HttpsAccDomesticValue"] = point.HttpsAccDomesticValue
		fields["HttpsValue"] = point.HttpsValue

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}

}

// 调用DescribeDomainQpsDataByLayer获取加速域名的每秒访问次数QPS
func (run *RunningProject) DescribeDomainQpsDataByLayer() {
	fmt.Println("[DescribeDomainQpsData] start...")
	request := cdn.CreateDescribeDomainQpsDataByLayerRequest()
	request.Scheme = "https"
	request.DomainName = run.cfg.DomainName
	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.IspNameEn = run.cfg.IspNameEn
	request.LocationNameEn = run.cfg.LocationNameEn
	request.Layer = run.cfg.Layer

	response, err := run.Client.DescribeDomainQpsDataByLayer(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainQpsDataByLayer, error: %s", err.Error())
	}
	for _, point := range response.QpsDataInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)

		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["AccDomesticValue"] = point.AccDomesticValue
		fields["AccOverseasValue"] = point.AccOverseasValue
		fields["AccValue"] = point.AccValue
		fields["DomesticValue"] = point.DomesticValue
		fields["OverseasValue"] = point.OverseasValue
		fields["AccValue"] = point.AccValue
		fields["Value"] = point.HttpsAccValue

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainSrcQpsData获取加速域名的回源请求QPS数据
func (run *RunningProject) DescribeDomainSrcQpsData() {
	fmt.Println("[DescribeDomainQpsData] start...")
	request := cdn.CreateDescribeDomainSrcQpsDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName
	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")
	fmt.Println("request ======>", request)
	response, err := run.Client.DescribeDomainSrcQpsData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainSrcQpsData, error: %s", err.Error())
	}
	fmt.Println("response ======>", response.SrcQpsDataPerInterval.DataModule)
	for _, point := range response.SrcQpsDataPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)

		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["Value"] = point.HttpsAccValue
		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainSrcTopUrlVisit获取加速域名5分钟计算粒度的回源热门Url
func (run *RunningProject) DescribeDomainSrcTopUrlVisit() {
	fmt.Println("[DescribeDomainSrcTopUrlVisit] start...")
	request := cdn.CreateDescribeDomainSrcTopUrlVisitRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName
	request.SortBy = run.cfg.SortBy

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")
	fmt.Println("request ======>", request)
	response, err := run.Client.DescribeDomainSrcTopUrlVisit(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainSrcTopUrlVisit, error: %s", err.Error())
	}
	fmt.Println("response ======>", response.AllUrlList.UrlList)
	for _, point := range response.AllUrlList.UrlList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["VisitData"] = point.VisitData
		fields["UrlDetail"] = point.UrlDetail
		fields["VisitProportion"] = point.VisitProportion
		fields["Flow"] = point.Flow
		fields["FlowProportion"] = point.FlowProportion

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
	}
}

// 调用DescribeDomainTopClientIpVisit获取加速域名在指定时间范围内按照访问次数或流量排序的Client IP排行
func (run *RunningProject) DescribeDomainTopClientIpVisit() {
	fmt.Println("[DescribeDomainTopClientIpVisit] start...")
	request := cdn.CreateDescribeDomainTopClientIpVisitRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName
	request.SortBy = run.cfg.SortBy

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.LocationNameEn = run.cfg.LocationNameEn
	request.Limit = fmt.Sprintf("%d", run.cfg.Limit)
	fmt.Println("request ======>", request)

	response, err := run.Client.DescribeDomainTopClientIpVisit(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainTopClientIpVisit, error: %s", err.Error())
	}
	fmt.Println("response ======>", response.ClientIpList)

	for _, point := range response.ClientIpList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["ClientIp"] = point.ClientIp

		fields["Rank"] = point.Rank
		fields["Traffic"] = point.Traffic
		fields["Acc"] = point.Acc

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
	}
}

// 调用DescribeDomainSrcQpsData获取加速域名的回源请求QPS数据
func (run *RunningProject) DescribeDomainTopReferVisit() {
	fmt.Println("[DescribeDomainTopReferVisit] start...")
	request := cdn.CreateDescribeDomainTopReferVisitRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName
	request.SortBy = run.cfg.SortBy
	request.Percent = run.cfg.Percent

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")
	fmt.Println("request ======>", request)

	response, err := run.Client.DescribeDomainTopReferVisit(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainTopReferVisit, error: %s", err.Error())
	}
	fmt.Println("response ======>", response.TopReferList.ReferList)
	for _, point := range response.TopReferList.ReferList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["Flow"] = point.Flow
		fields["VisitProportion"] = point.VisitProportion
		fields["VisitData"] = point.VisitData
		fields["ReferDetail"] = point.ReferDetail
		fields["FlowProportion"] = point.FlowProportion

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
	}
}

// 调用DescribeDomainTopUrlVisit获取加速域名某天内的热门URL列表
func (run *RunningProject) DescribeDomainTopUrlVisit() {
	request := cdn.CreateDescribeDomainTopUrlVisitRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName
	request.SortBy = run.cfg.SortBy

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	response, err := run.Client.DescribeDomainTopUrlVisit(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainTopUrlVisit, error: %s", err.Error())
	}

	for _, point := range response.AllUrlList.UrlList {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["Flow"] = point.Flow
		fields["VisitProportion"] = point.VisitProportion
		fields["VisitData"] = point.VisitData
		fields["UrlDetail"] = point.UrlDetail
		fields["FlowProportion"] = point.FlowProportion

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
	}
}

// 调DescribeDomainAverageResponseTime获取加速域名的平均响应时间
func (run *RunningProject) DescribeDomainAverageResponseTime() {
	request := cdn.CreateDescribeDomainAverageResponseTimeRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName
	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.IspNameEn = run.cfg.IspNameEn
	request.LocationNameEn = run.cfg.LocationNameEn
	request.DomainType = run.cfg.DomainType

	response, err := run.Client.DescribeDomainAverageResponseTime(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainAverageResponseTime, error: %s", err.Error())
	}

	for _, point := range response.AvgRTPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)

		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["Value"] = point.Value

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调DescribeDomainAverageResponseTime获取加速域名的平均响应时间
func (run *RunningProject) DescribeDomainFileSizeProportionData() {
	request := cdn.CreateDescribeDomainFileSizeProportionDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	response, err := run.Client.DescribeDomainFileSizeProportionData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainFileSizeProportionData, error: %s", err.Error())
	}

	for _, data := range response.FileSizeProportionDataInterval.UsageData {
		for _, point := range data.Value.FileSizeProportionData {
			tags := map[string]string{}
			fields := map[string]interface{}{}

			tags["TimeStamp"] = data.TimeStamp
			tags["DomainName"] = response.DomainName

			fields["FileSize"] = point.FileSize
			fields["Proportion"] = point.Proportion

			run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
		}
	}
}

// 调用DescribeDomainBpsDataByTimeStamp获取加速域名在某个时刻的带宽数据
func (run *RunningProject) DescribeDomainBpsDataByTimeStamp() {
	request := cdn.CreateDescribeDomainBpsDataByTimeStampRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName
	nt := time.Now()
	et := nt.Unix()

	request.TimePoint = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.IspNames = run.cfg.IspNames
	request.LocationNames = run.cfg.LocationNames

	response, err := run.Client.DescribeDomainBpsDataByTimeStamp(request)
	if err != nil {
		fmt.Print(err.Error())
	}

	for _, point := range response.BpsDataList.BpsDataModel {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName
		tags["TimeStamp"] = response.TimeStamp
		tags["LocationName"] = point.LocationName
		tags["IspName"] = point.IspName

		fields["Bps"] = point.Bps

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
	}
}

// 调用DescribeDomainBpsDataByTimeStamp获取加速域名在某个时刻的带宽数据
func (run *RunningProject) DescribeDomainISPData() {
	request := cdn.CreateDescribeDomainISPDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	response, err := run.Client.DescribeDomainISPData(request)
	if err != nil {
		fmt.Print(err.Error())
	}

	for _, point := range response.Value.ISPProportionData {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName
		tags["ISP"] = point.ISP
		tags["IspEname"] = point.IspEname

		fields["Proportion"] = point.Proportion
		fields["AvgObjectSize"] = point.AvgObjectSize
		fields["AvgResponseTime"] = point.AvgResponseTime
		fields["Bps"] = point.Bps
		fields["Qps"] = point.Qps
		fields["AvgResponseRate"] = point.AvgResponseRate
		fields["TotalBytes"] = point.TotalBytes
		fields["BytesProportion"] = point.BytesProportion
		fields["TotalQuery"] = point.TotalQuery

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
	}
}

// 调用DescribeDomainRealTimeBpsData获取加速域名1分钟粒度带宽数据，支持查询7天内的数据
func (run *RunningProject) DescribeDomainRealTimeBpsData() {
	request := cdn.CreateDescribeDomainRealTimeBpsDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName
	request.LocationNameEn = run.cfg.LocationNameEn
	request.IspNameEn = run.cfg.IspNameEn

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	response, err := run.Client.DescribeDomainRealTimeBpsData(request)
	if err != nil {
		fmt.Print(err.Error())
		run.logger.Warnf("action:DescribeDomainRealTimeBpsData, error: %s", err.Error())
	}

	for _, point := range response.Data.BpsModel {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)

		tags := map[string]string{}
		fields := map[string]interface{}{}

		fields["Bps"] = point.Bps

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainRealTimeSrcBpsData获取域名1分钟粒度回源带宽数据，支持获取最近7天的数据
func (run *RunningProject) DescribeDomainRealTimeSrcBpsData() {
	request := cdn.CreateDescribeDomainRealTimeSrcBpsDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	response, err := run.Client.DescribeDomainRealTimeSrcBpsData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainRealTimeSrcBpsData, error: %s", err.Error())
	}

	for _, point := range response.RealTimeSrcBpsDataPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["Value"] = point.Value

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainRealTimeSrcHttpCodeData获取加速域名回源1分钟粒度HTTP返回码的总数和占比数据，支持获取最近7天的数据。
func (run *RunningProject) DescribeDomainRealTimeSrcHttpCodeData() {
	request := cdn.CreateDescribeDomainRealTimeSrcHttpCodeDataRequest()
	request.Scheme = "https"
	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.IspNameEn = run.cfg.IspNameEn
	request.LocationNameEn = run.cfg.LocationNameEn

	response, err := run.Client.DescribeDomainRealTimeSrcHttpCodeData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainQpsData, error: %s", err.Error())
	}

	for _, data := range response.RealTimeSrcHttpCodeData.UsageData {
		for _, point := range data.Value.RealTimeSrcCodeProportionData {
			tags := map[string]string{}
			fields := map[string]interface{}{}

			tags["DomainName"] = data.TimeStamp
			fields["Proportion"] = point.Proportion
			fields["Code"] = point.Code

			run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
		}
	}
}

// 调用DescribeDomainRealTimeSrcTrafficData获取加速域名的1分钟回源流量监控数据，支持获取最近90天的数据，数据单位：byte。
func (run *RunningProject) DescribeDomainRealTimeSrcTrafficData() {
	request := cdn.CreateDescribeDomainRealTimeSrcTrafficDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	response, err := run.Client.DescribeDomainRealTimeSrcTrafficData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainRealTimeSrcTrafficData, error: %s", err.Error())
	}

	for _, point := range response.RealTimeSrcTrafficDataPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["Value"] = point.Value

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainRealTimeByteHitRateData获取域名1分钟粒度字节命中率数据
func (run *RunningProject) DescribeDomainRealTimeByteHitRateData() {
	request := cdn.CreateDescribeDomainRealTimeByteHitRateDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	response, err := run.Client.DescribeDomainRealTimeByteHitRateData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainRealTimeByteHitRateData, error: %s", err.Error())
	}

	for _, point := range response.Data.ByteHitRateDataModel {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		fields["ByteHitRate"] = point.ByteHitRate

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainRealTimeQpsData获取加速域名1分钟粒度每秒访问次数数据
func (run *RunningProject) DescribeDomainRealTimeQpsData() {
	request := cdn.CreateDescribeDomainRealTimeQpsDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.IspNameEn = run.cfg.IspNameEn
	request.LocationNameEn = run.cfg.LocationNameEn

	response, err := run.Client.DescribeDomainRealTimeQpsData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainRealTimeQpsData, error: %s", err.Error())
	}

	for _, point := range response.Data.QpsModel {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		fields["Qps"] = point.Qps

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainRealTimeTrafficData获取加速域名的1分钟粒度流量监控数据
func (run *RunningProject) DescribeDomainRealTimeTrafficData() {
	request := cdn.CreateDescribeDomainRealTimeTrafficDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	response, err := run.Client.DescribeDomainRealTimeTrafficData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainRealTimeTrafficData, error: %s", err.Error())
	}

	for _, point := range response.RealTimeTrafficDataPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		fields["Value"] = point.Value

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainRealTimeReqHitRateData获取加速域名1分钟粒度请求命中率数据
func (run *RunningProject) DescribeDomainRealTimeReqHitRateData() {
	request := cdn.CreateDescribeDomainRealTimeReqHitRateDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	response, err := run.Client.DescribeDomainRealTimeReqHitRateData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainRealTimeReqHitRateData, error: %s", err.Error())
	}

	for _, point := range response.Data.ReqHitRateDataModel {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		fields["ReqHitRate"] = point.ReqHitRate

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainBpsData获取加速域名的网络带宽监控数据
func (run *RunningProject) DescribeDomainBpsData() {
	request := cdn.CreateDescribeDomainBpsDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName
	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")
	request.IspNameEn = run.cfg.IspNameEn
	request.LocationNameEn = run.cfg.LocationNameEn

	response, err := run.Client.DescribeDomainBpsData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainBpsData, error: %s", err.Error())
	}

	for _, point := range response.BpsDataPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		fields["PeakTime"] = point.PeakTime
		fields["OverseasValue"] = point.OverseasValue
		fields["SpecialValue"] = point.SpecialValue
		fields["HttpsAccOverseasValue"] = point.HttpsAccOverseasValue
		fields["HttpsOverseasValue"] = point.HttpsOverseasValue
		fields["DomesticValue"] = point.DomesticValue
		fields["AccValue"] = point.AccValue
		fields["Value"] = point.Value
		fields["AccDomesticValue"] = point.AccDomesticValue
		fields["HttpsDomesticValue"] = point.HttpsDomesticValue
		fields["HttpsValue"] = point.HttpsValue
		fields["HttpsAccValue"] = point.HttpsAccValue
		fields["AccOverseasValue"] = point.AccOverseasValue
		fields["HttpsAccDomesticValue"] = point.HttpsAccDomesticValue

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainSrcBpsData获取加速域名的回源带宽监控数据
func (run *RunningProject) DescribeDomainSrcBpsData() {
	request := cdn.CreateDescribeDomainSrcBpsDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

	response, err := run.Client.DescribeDomainSrcBpsData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainSrcBpsData, error: %s", err.Error())
	}

	for _, point := range response.SrcBpsDataPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["Value"] = point.Value
		fields["HttpsValue"] = point.HttpsValue

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainSrcTrafficData获取加速域名的回源流量监控数据
func (run *RunningProject) DescribeDomainSrcTrafficData() {
	request := cdn.CreateDescribeDomainSrcTrafficDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

	response, err := run.Client.DescribeDomainSrcTrafficData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainSrcTrafficData, error: %s", err.Error())
	}

	for _, point := range response.SrcTrafficDataPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["Value"] = point.Value
		fields["HttpsValue"] = point.HttpsValue

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainHitRateData获取加速域名的字节命中率（命中字节百分比）
func (run *RunningProject) DescribeDomainHitRateData() {
	request := cdn.CreateDescribeDomainHitRateDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

	response, err := run.Client.DescribeDomainHitRateData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainHitRateData, error: %s", err.Error())
	}

	for _, point := range response.HitRateInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["Value"] = point.Value
		fields["HttpsValue"] = point.HttpsValue

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainReqHitRateData获取加速域名的请求命中率（命中请求百分比）
func (run *RunningProject) DescribeDomainReqHitRateData() {
	request := cdn.CreateDescribeDomainReqHitRateDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

	response, err := run.Client.DescribeDomainReqHitRateData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainTrafficData, error: %s", err.Error())
	}

	for _, point := range response.ReqHitRateInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["Value"] = point.Value
		fields["HttpsValue"] = point.HttpsValue

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainHttpCodeData获取加速域名HTTP返回码的总数和占比数据
func (run *RunningProject) DescribeDomainHttpCodeData() {
	request := cdn.CreateDescribeDomainHttpCodeDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

	response, err := run.Client.DescribeDomainHttpCodeData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainHttpCodeData, error: %s", err.Error())
	}

	for _, data := range response.HttpCodeData.UsageData {
		for _, point := range data.Value.CodeProportionData {
			tags := map[string]string{}
			fields := map[string]interface{}{}

			tags["TimeStamp"] = data.TimeStamp
			tags["DomainName"] = response.DomainName

			tags["Code"] = point.Code

			fields["Proportion"] = point.Proportion

			run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags)
		}
	}
}

// 调用DescribeDomainTrafficData获取加速域名的网络流量监控数据
func (run *RunningProject) DescribeDomainTrafficData() {
	request := cdn.CreateDescribeDomainTrafficDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	request.IspNameEn = run.cfg.IspNameEn
	request.LocationNameEn = run.cfg.LocationNameEn
	request.Interval = fmt.Sprintf("%d", run.cfg.Interval)

	response, err := run.Client.DescribeDomainTrafficData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainTrafficData, error: %s", err.Error())
	}

	for _, point := range response.TrafficDataPerInterval.DataModule {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)

		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["DomainName"] = response.DomainName

		fields["Value"] = point.Value
		fields["DomesticValue"] = point.DomesticValue
		fields["OverseasValue"] = point.OverseasValue
		fields["HttpsValue"] = point.HttpsValue
		fields["HttpsDomesticValue"] = point.HttpsDomesticValue
		fields["HttpsOverseasValue"] = point.HttpsOverseasValue

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainRegionData获取加速域名天粒度的用户区域分布数据统计
func (run *RunningProject) DescribeDomainUvData() {
	request := cdn.CreateDescribeDomainUvDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	response, err := run.Client.DescribeDomainUvData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainUvData, error: %s", err.Error())
	}

	for _, point := range response.UvDataInterval.UsageData {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)

		tags := map[string]string{}
		fields := map[string]interface{}{}

		fields["Value"] = point.Value
		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}

// 调用DescribeDomainPvData获取加速域名1小时粒度的PV页面访问统计
func (run *RunningProject) DescribeDomainPvData() {
	request := cdn.CreateDescribeDomainPvDataRequest()
	request.Scheme = "https"

	request.DomainName = run.cfg.DomainName

	nt := time.Now()
	et := nt.Unix()
	st := nt.Add(-(5 * time.Minute)).Unix()

	request.StartTime = time.Unix(st, 0).Format("2006-01-02T15:04:05Z07:00")
	request.EndTime = time.Unix(et, 0).Format("2006-01-02T15:04:05Z07:00")

	response, err := run.Client.DescribeDomainPvData(request)
	if err != nil {
		run.logger.Warnf("action:DescribeDomainPvData, error: %s", err.Error())
	}

	for _, point := range response.PvDataInterval.UsageData {
		const layout = time.RFC3339
		tm, _ := time.Parse(layout, point.TimeStamp)

		tags := map[string]string{}
		fields := map[string]interface{}{}

		fields["Value"] = point.Value

		run.inst.agent.accumulator.AddFields(run.cfg.MetricName, fields, tags, tm)
	}
}
