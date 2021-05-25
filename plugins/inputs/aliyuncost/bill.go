package aliyuncost

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type costBill struct {
	interval    time.Duration
	measurement string
	ag          *agent

	historyFlag int32
}

func newCostBill(ag *agent) *costBill {
	c := &costBill{
		ag:          ag,
		measurement: "aliyun_cost_bill",
		interval:    ag.BiilInterval.Duration,
	}
	return c
}

func (cb *costBill) run(ctx context.Context) {

	if cb.ag.CollectHistoryData && !cb.ag.isTest() {
		go func() {
			cb.getHistoryData(ctx)
		}()

	}

	cb.getData(ctx)

	moduleLogger.Info("bill done")
}

func (cb *costBill) getData(ctx context.Context) {

	for {
		//暂停历史数据抓取
		atomic.AddInt32(&cb.historyFlag, 1)

		//以月为单位
		//计费项 + 明细
		start := time.Now().Truncate(time.Hour)
		cycle := fmt.Sprintf("%d-%02d", start.Year(), start.Month())
		if err := cb.getBills(ctx, cycle, nil); err != nil {
			moduleLogger.Errorf("%s", err)
			if cb.ag.isTest() {
				cb.ag.testError = err
				return
			}
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		atomic.AddInt32(&cb.historyFlag, -1)

		if cb.ag.isTest() {
			break
		}

		datakit.SleepContext(ctx, cb.interval)

		select {
		case <-ctx.Done():
			return
		default:
		}
	}

}

func (cb *costBill) getHistoryData(ctx context.Context) {

	key := "." + cb.ag.cacheFileKey(`billV2`)

	if !cb.ag.CollectHistoryData {
		delAliyunCostHistory(key)
		return
	}

	moduleLogger.Info("start getting history Bills")

	var info *historyInfo

	if !cb.ag.isDebug() {
		info, _ = getAliyunCostHistory(key)
	}

	if info == nil {
		info = &historyInfo{}
	} else if info.Statue == 1 {
		moduleLogger.Infof("already fetched the history data")
		return
	}

	if info.StartTime.IsZero() {
		info.Statue = 0
		info.StartTime = time.Now().Truncate(time.Hour).AddDate(-1, 0, 0)
		info.EndTime = time.Now().Truncate(time.Hour)
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
			if !cb.ag.isDebug() {
				setAliyunCostHistory(key, info)
			}
			break
		}

		cycle := fmt.Sprintf("%d-%02d", info.StartTime.Year(), info.StartTime.Month())

		if err := cb.getBills(ctx, cycle, info); err != nil {
			moduleLogger.Errorf("%s", err)
			time.Sleep(time.Minute)
			continue
		}

		info.StartTime = info.StartTime.AddDate(0, 1, 0)
		if !cb.ag.isDebug() {
			setAliyunCostHistory(key, info)
		}
	}
}

func (cb *costBill) getBills(ctx context.Context, cycle string, info *historyInfo) error {

	defer func() {
		if e := recover(); e != nil {
			moduleLogger.Errorf("panic: %v", e)
		}
	}()

	logPrefix := ""
	if info != nil {
		logPrefix = "(history) "
	}

	moduleLogger.Infof("%sBills of %s start", logPrefix, cycle)

	req := bssopenapi.CreateQueryBillRequest()
	req.BillingCycle = cycle
	req.Scheme = "https"
	req.PageSize = requests.NewInteger(300)
	req.IsHideZeroCharge = requests.NewBoolean(true) //过滤掉原价为0

	for {

		if info != nil {
			for atomic.LoadInt32(&cb.historyFlag) > 0 {
				select {
				case <-ctx.Done():
					return nil
				default:
				}
				datakit.SleepContext(ctx, 3*time.Second)
			}
		}

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		resp, err := cb.ag.queryBill(ctx, req)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			return fmt.Errorf("fail to get bill of %s", cycle)
		}

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		moduleLogger.Debugf(" %sPage(%d): count=%d, TotalCount=%d, PageSize=%d", logPrefix, resp.Data.PageNum, len(resp.Data.Items.Item), resp.Data.TotalCount, resp.Data.PageSize)

		if err := cb.parseBillResponse(ctx, resp); err != nil {
			return err
		}

		if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
			req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
		} else {
			break
		}
	}

	moduleLogger.Infof("%sBills of %s end", logPrefix, cycle)

	return nil
}

func (cb *costBill) parseBillResponse(ctx context.Context, resp *bssopenapi.QueryBillResponse) error {

	for _, item := range resp.Data.Items.Item {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		tags := map[string]string{
			"AccountID":    resp.Data.AccountID,
			"AccountName":  resp.Data.AccountName,
			"OwnerID":      item.OwnerID,
			"BillingCycle": resp.Data.BillingCycle,
		}

		tags["Item"] = item.Item
		tags["ProductCode"] = item.ProductCode
		tags["ProductName"] = item.ProductName
		tags["ProductType"] = item.ProductType
		tags["SubscriptionType"] = item.SubscriptionType
		tags["Status"] = item.Status
		tags[`Currency`] = item.Currency
		tags[`CostUnit`] = item.CostUnit
		tags[`Region`] = item.Region

		fields := map[string]interface{}{}

		fields[`RecordID`] = item.RecordID
		fields[`PretaxGrossAmount`] = item.PretaxGrossAmount
		fields[`DeductedByCoupons`] = item.DeductedByCoupons
		fields[`InvoiceDiscount`] = item.InvoiceDiscount
		fields[`RoundDownDiscount`], _ = strconv.ParseFloat(datakit.NumberFormat(item.RoundDownDiscount), 64)
		fields[`PretaxAmount`] = item.PretaxAmount
		fields[`DeductedByCashCoupons`] = item.DeductedByCashCoupons
		fields[`DeductedByPrepaidCard`] = item.DeductedByPrepaidCard
		fields[`PaymentAmount`] = item.PaymentAmount
		fields[`OutstandingAmount`] = item.OutstandingAmount

		billtime := item.UsageStartTime
		if billtime == "" {
			continue
		}
		t, err := time.Parse(`2006-01-02 15:04:05`, billtime)
		if err != nil {
			moduleLogger.Warnf("fail to parse time:%v of product:%s, error: %s", billtime, item.ProductName, err)
		} else {
			//返回的不是utc
			t = t.Add(-8 * time.Hour)
			if cb.ag.isTest() {
				// pass
			} else if cb.ag.isDebug() {
				data, _ := io.MakeMetric(cb.getName(), tags, fields, t)
				fmt.Printf("%s\n", string(data))
			} else {
				io.NamedFeedEx(inputName, datakit.Metric, cb.getName(), tags, fields, t)
			}
		}
	}

	return nil
}

func (cb *costBill) getInterval() time.Duration {
	return cb.interval
}

func (cb *costBill) getName() string {
	return cb.measurement
}
