package tencentcost

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"

	billing "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/billing/v20180709"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type bill struct {
	interval    time.Duration
	rateLimiter *rate.Limiter
	agent       *TencentCost
	client      *billing.Client

	historyFlag int32
}

func newBill(ag *TencentCost) *bill {
	b := &bill{
		agent:    ag,
		interval: ag.BillInterval.Duration,
	}
	return b
}

func (b *bill) getName() string {
	return "tencent_cost_bill"
}

func (b *bill) run(ctx context.Context) {
	limit := rate.Every(300 * time.Millisecond)
	b.rateLimiter = rate.NewLimiter(limit, 1)

	credential := common.NewCredential(
		b.agent.AccessKeyID,
		b.agent.AccessKeySecret,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "billing.tencentcloudapi.com"
	b.client, _ = billing.NewClient(credential, "", cpf)

	if b.agent.CollectHistoryData {
		go func() {
			b.getHistoryData(ctx)
		}()
	}

	b.getData(ctx)
}

func (b *bill) getData(ctx context.Context) {
	var lastTime time.Time
	shift := time.Minute

	for {

		if b.agent.CollectHistoryData {
			atomic.AddInt32(&b.historyFlag, 1)
		}
		start := time.Now()

		endTime := time.Now().Truncate(time.Minute)
		if lastTime.IsZero() {
			lastTime = endTime.Add(-b.interval)
		}

		err := b.describeDealsByCond(ctx, lastTime, endTime, "", nil)
		if err == nil {
			lastTime = endTime.Add(-shift)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		usage := time.Now().Sub(start)
		if b.interval > usage {
			if b.agent.CollectHistoryData {
				atomic.AddInt32(&b.historyFlag, -1)
			}
			datakit.SleepContext(ctx, b.interval-usage)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (b *bill) getHistoryData(ctx context.Context) {
	key := "." + cacheFileKey(`bills`, b.agent.AccessKeyID, b.agent.AccessKeySecret)

	moduleLogger.Info("start getting history Bills")

	info, _ := getAliyunCostHistory(key)

	if info == nil {
		info = &historyInfo{}
	} else if info.Statue == 1 {
		moduleLogger.Infof("already fetched the history data")
		return
	}

	if info.StartTime.IsZero() {
		now := time.Now().Truncate(time.Minute)
		start := now.Add(-time.Hour * 8760)
		info.StartTime = start
		info.EndTime = now
		info.Statue = 0
		info.Offset = 0
	}

	info.key = key

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if info.StartTime.Unix() >= info.EndTime.Unix() {
			info.Statue = 1
			setAliyunCostHistory(key, info)
			break
		}

		month := fmt.Sprintf("%d-%02d", info.StartTime.Year(), info.StartTime.Month())

		if err := b.describeDealsByCond(ctx, info.StartTime, info.EndTime, month, info); err != nil {
			moduleLogger.Errorf("%s", err)
		}

		info.StartTime = info.StartTime.AddDate(0, 1, 0)
		setAliyunCostHistory(key, info)
	}

}

//https://cloud.tencent.com/document/api/555/19182
func (b *bill) describeDealsByCond(ctx context.Context, startTime time.Time, endTime time.Time, month string, history *historyInfo) error {

	logPrefix := ""
	if history != nil {
		logPrefix = "(history) "
	}

	var err error
	var response *billing.DescribeBillDetailResponse
	var periodType = "byPayTime"
	var page uint64 = 100
	var offset uint64 = 0
	if history != nil {
		offset = history.Offset
	}

	st := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", startTime.Year(), startTime.Month(), startTime.Day(), startTime.Hour(), startTime.Minute(), startTime.Second())
	et := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", endTime.Year(), endTime.Month(), endTime.Day(), endTime.Hour(), endTime.Minute(), endTime.Second())

	if month == "" {
		moduleLogger.Infof("%sgetting Bills(%s - %s)", logPrefix, st, et)
	} else {
		moduleLogger.Infof("%sgetting Bills(%s)", logPrefix, month)
	}

	for { //分页获取

		for i := 0; i < 5; i++ {

			if history != nil {
				for atomic.LoadInt32(&b.historyFlag) == 1 {
					select {
					case <-ctx.Done():
						return nil
					default:
					}
					time.Sleep(time.Second)
				}
			}

			select {
			case <-ctx.Done():
				return nil
			default:
			}

			b.rateLimiter.Wait(ctx)

			request := billing.NewDescribeBillDetailRequest()
			if month != "" {
				request.Month = &month
			} else {
				request.BeginTime = &st
				request.EndTime = &et
			}
			request.Offset = &offset
			request.Limit = &page
			request.PeriodType = &periodType

			response, err = b.client.DescribeBillDetail(request)

			if err != nil {
				if _, ok := err.(*errors.TencentCloudSDKError); ok {
					moduleLogger.Errorf("%sAn API error has returned: %s", logPrefix, err)
				} else {
					moduleLogger.Errorf("%s%s", logPrefix, err)
					break
				}
			} else {
				break
			}

			datakit.SleepContext(ctx, time.Millisecond*200)
		}

		if err != nil {
			break
		}

		var totalCount uint64 = 0
		if response.Response != nil && response.Response.Total != nil {
			totalCount = *response.Response.Total
		}

		if month != "" {
			moduleLogger.Debugf("%sBills(%s): TotalCount=%d, Offset=%d, count=%d", logPrefix, month, totalCount, offset, len(response.Response.DetailSet))
		} else {
			moduleLogger.Debugf("%sBills(%s - %s): TotalCount=%d, Offset=%d, count=%d", logPrefix, st, et, totalCount, offset, len(response.Response.DetailSet))
		}

		b.handleResponse(ctx, response)

		offset += uint64(len(response.Response.DetailSet))

		if history != nil {
			history.Offset = uint64(offset)
			setAliyunCostHistory(history.key, history)
		}

		if len(response.Response.DetailSet) < 100 {
			break
		}
	}

	if month == "" {
		moduleLogger.Debugf("%sfinish getting Bills(%s - %s)", logPrefix, st, et)
	} else {
		moduleLogger.Debugf("%sfinish getting Bills(%s)", logPrefix, month)
	}

	if history != nil {
		history.Statue = 1
		setAliyunCostHistory(history.key, history)
	}

	return err
}

func (b *bill) handleResponse(ctx context.Context, response *billing.DescribeBillDetailResponse) {
	if response == nil {
		return
	}

	for _, item := range response.Response.DetailSet {

		for _, comp := range item.ComponentSet {

			tags := map[string]string{
				"BusinessCodeName": ensureString(item.BusinessCodeName),
				"ProductCodeName":  ensureString(item.ProductCodeName),
				"PayModeName":      ensureString(item.PayModeName),
				"ProjectName":      ensureString(item.ProjectName),
				"RegionName":       ensureString(item.RegionName),
				"ZoneName":         ensureString(item.ZoneName),
				"ActionTypeName":   ensureString(item.ActionTypeName),
				"PayerUin":         ensureString(item.PayerUin),
				"OwnerUin":         ensureString(item.OwnerUin),
				"OperateUin":       ensureString(item.OperateUin),
			}
			for _, t := range item.Tags {
				if t.TagValue != nil {
					tags[*t.TagKey] = *t.TagValue
				}
			}
			b.agent.appendCustomTags(tags)

			fields := map[string]interface{}{
				"OrderId":      ensureString(item.OrderId),
				"BillId":       ensureString(item.BillId),
				"ResourceId":   ensureString(item.ResourceId),
				"FeeBeginTime": ensureString(item.FeeBeginTime),
				"FeeEndTime":   ensureString(item.FeeEndTime),
				"ResourceName": ensureString(item.ResourceName),

				"ComponentCodeName":  ensureString(comp.ComponentCodeName),
				"ItemCodeName":       ensureString(comp.ItemCodeName),
				"SinglePrice":        ensureString(comp.SinglePrice),
				"SpecifiedPrice":     ensureString(comp.SpecifiedPrice),
				"PriceUnit":          ensureString(comp.PriceUnit),
				"UsedAmount":         ensureString(comp.UsedAmount),
				"UsedAmountUnit":     ensureString(comp.UsedAmountUnit),
				"TimeSpan":           ensureString(comp.TimeSpan),
				"TimeUnitName":       ensureString(comp.TimeUnitName),
				"Cost":               ensureString(comp.Cost),
				"Discount":           ensureString(comp.Discount),
				"ReduceType":         ensureString(comp.ReduceType),
				"RealCost":           ensureString(comp.RealCost),
				"VoucherPayAmount":   ensureString(comp.VoucherPayAmount),
				"CashPayAmount":      ensureString(comp.CashPayAmount),
				"IncentivePayAmount": ensureString(comp.IncentivePayAmount),
				"ContractPrice":      ensureString(comp.ContractPrice),
			}

			metrictime := time.Now().UTC()
			if item.PayTime != nil {
				tm, e := time.ParseInLocation(`2006-01-02 15:04:05`, *item.PayTime, time.Local)
				if e != nil {
					moduleLogger.Warnf("fail to parse time, %s", e)
				} else {
					metrictime = tm.UTC()
				}

			}

			io.NamedFeedEx(inputName, io.Metric, b.getName(), tags, fields, metrictime)
		}
	}
}
