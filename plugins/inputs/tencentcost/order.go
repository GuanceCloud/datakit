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

type order struct {
	interval    time.Duration
	rateLimiter *rate.Limiter
	agent       *TencentCost
	client      *billing.Client

	historyFlag int32
}

func newOrder(ag *TencentCost) *order {
	o := &order{
		agent:    ag,
		interval: ag.OrderInterval.Duration,
	}
	return o
}

func (o *order) getName() string {
	return "tencent_cost_order"
}

func (o *order) run(ctx context.Context) {
	limit := rate.Every(60 * time.Millisecond)
	o.rateLimiter = rate.NewLimiter(limit, 1)

	credential := common.NewCredential(
		o.agent.AccessKeyID,
		o.agent.AccessKeySecret,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "billing.tencentcloudapi.com"
	o.client, _ = billing.NewClient(credential, "", cpf)

	if o.agent.CollectHistoryData && !o.agent.isTest() {
		go func() {
			o.getHistoryData(ctx)
		}()
	}

	o.getData(ctx)
}

func (o *order) getData(ctx context.Context) {

	var lastTime time.Time
	shift := time.Minute

	for {

		if o.agent.CollectHistoryData {
			atomic.AddInt32(&o.historyFlag, 1)
		}
		start := time.Now()

		endTime := time.Now().Truncate(time.Minute)
		if lastTime.IsZero() {
			lastTime = endTime.Add(-o.interval)
		}

		err := o.describeDealsByCond(ctx, lastTime, endTime, nil)
		if err == nil {
			lastTime = endTime.Add(-shift)
		}

		if o.agent.isTest() {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		usage := time.Now().Sub(start)
		if o.interval > usage {
			if o.agent.CollectHistoryData {
				atomic.AddInt32(&o.historyFlag, -1)
			}
			datakit.SleepContext(ctx, o.interval-usage)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (o *order) getHistoryData(ctx context.Context) {

	key := "." + cacheFileKey(`orders`, o.agent.AccessKeyID, o.agent.AccessKeySecret)

	moduleLogger.Info("start getting history Orders")

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

	o.describeDealsByCond(ctx, info.StartTime, info.EndTime, info)
}

//https://cloud.tencent.com/document/api/555/19179
func (o *order) describeDealsByCond(ctx context.Context, startTime time.Time, endTime time.Time, history *historyInfo) error {

	logPrefix := ""
	if history != nil {
		logPrefix = "(history) "
	}

	var err error
	var response *billing.DescribeDealsByCondResponse
	var page int64 = 100
	var offset int64 = 0
	if history != nil {
		offset = int64(history.Offset)
	}

	st := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", startTime.Year(), startTime.Month(), startTime.Day(), startTime.Hour(), startTime.Minute(), startTime.Second())
	et := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", endTime.Year(), endTime.Month(), endTime.Day(), endTime.Hour(), endTime.Minute(), endTime.Second())

	moduleLogger.Infof("%sgetting Orders(%s - %s)", logPrefix, st, et)

	for { //分页获取

		for i := 0; i < 5; i++ {

			if history != nil {
				for atomic.LoadInt32(&o.historyFlag) == 1 {
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

			o.rateLimiter.Wait(ctx)

			request := billing.NewDescribeDealsByCondRequest()
			request.StartTime = &st
			request.EndTime = &et
			request.Offset = &offset
			request.Limit = &page

			response, err = o.client.DescribeDealsByCond(request)

			if err != nil {
				if _, ok := err.(*errors.TencentCloudSDKError); ok {
					moduleLogger.Errorf("An API error has returned: %s", err)
				} else {
					moduleLogger.Errorf("%s", err)
					break
				}
				if o.agent.isTest() {
					break
				}
			} else {
				break
			}

			datakit.SleepContext(ctx, time.Millisecond*200)
		}

		if err != nil {
			if o.agent.isTest() {
				o.agent.testError = err
			}
			break
		}

		moduleLogger.Debugf("%sOrders(%s - %s): TotalCount=%d, Offset=%d, count=%d", logPrefix, st, et, *response.Response.TotalCount, offset, len(response.Response.Deals))

		o.handleResponse(ctx, response)

		offset += int64(len(response.Response.Deals))

		if history != nil {
			history.Offset = uint64(offset)
			setAliyunCostHistory(history.key, history)
		}

		if len(response.Response.Deals) < 100 {
			break
		}
	}

	moduleLogger.Debugf("%sfinish getting Orders(%s - %s)", logPrefix, st, et)

	if history != nil {
		history.Statue = 1
		setAliyunCostHistory(history.key, history)
	}

	return err
}

func (o *order) handleResponse(ctx context.Context, response *billing.DescribeDealsByCondResponse) {
	if response == nil {
		return
	}

	for _, item := range response.Response.Deals {
		tags := map[string]string{
			"PayMode":  ensureString(item.PayMode),
			"Payer":    ensureString(item.Payer),
			"Creator":  ensureString(item.Creator),
			"Currency": ensureString(item.Currency),
			"TimeUnit": ensureString(item.TimeUnit),
		}
		if item.Status != nil {
			tags["Status"] = fmt.Sprintf("%d", *item.Status)
		}
		if item.ProjectId != nil {
			tags["ProjectId"] = fmt.Sprintf("%d", *item.ProjectId)
		}
		if item.GoodsCategoryId != nil {
			tags["GoodsCategoryId"] = fmt.Sprintf("%d", *item.GoodsCategoryId)
		}
		o.agent.appendCustomTags(tags)

		fields := map[string]interface{}{
			"OrderId":        ensureString(item.OrderId),
			"ProductCode":    ensureString(item.ProductCode),
			"SubProductCode": ensureString(item.SubProductCode),
			"BigDealId":      ensureString(item.BigDealId),
			"Formula":        ensureString(item.Formula),
			"RefReturnDeals": ensureString(item.RefReturnDeals),
		}

		if item.Price != nil {
			fields["Price"] = *item.Price
		}
		if item.Policy != nil {
			fields["Policy"] = *item.Policy
		}
		if item.TotalCost != nil {
			fields["TotalCost"] = *item.TotalCost
		}
		if item.RealTotalCost != nil {
			fields["RealTotalCost"] = *item.RealTotalCost
		}
		if item.VoucherDecline != nil {
			fields["VoucherDecline"] = *item.VoucherDecline
		}
		if item.TimeSpan != nil {
			fields["TimeSpan"] = *item.TimeSpan
		}

		for _, info := range item.ProductInfo {
			if info.Value != nil {
				fields["Product"+*info.Name] = *info.Value
			}
		}

		metrictime := time.Now().UTC()
		if item.CreateTime != nil {
			tm, e := time.ParseInLocation(`2006-01-02 15:04:05`, *item.CreateTime, time.Local)
			if e != nil {
				moduleLogger.Warnf("fail to parse time, %s", e)
			} else {
				metrictime = tm.UTC()
			}

		}

		if o.agent.isTest() {
			// pass
		} else {
			io.NamedFeedEx(inputName, datakit.Metric, o.getName(), tags, fields, metrictime)
		}
	}
}
