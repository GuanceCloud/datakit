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

	if cb.ag.CollectHistoryData {
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
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		/*
			//实例账单: 实例 + 按天
			if err := cb.getInstnceBills(ctx, start.Year(), int(start.Month())); err != nil && err != context.Canceled {
				moduleLogger.Errorf("%s", err)
			}
		*/

		atomic.AddInt32(&cb.historyFlag, -1)
		datakit.SleepContext(ctx, cb.interval)

		select {
		case <-ctx.Done():
			return
		default:
		}
	}

}

func (cb *costBill) getHistoryData(ctx context.Context) {

	key := "." + cb.ag.cacheFileKey(`bill`)

	if !cb.ag.CollectHistoryData {
		delAliyunCostHistory(key)
		return
	}

	moduleLogger.Info("start getting history Bills")

	var info *historyInfo

	if !cb.ag.debugMode {
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
			if !cb.ag.debugMode {
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
		if !cb.ag.debugMode {
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

	moduleLogger.Infof("%sgetting Bills of %s", logPrefix, cycle)

	req := bssopenapi.CreateQueryBillRequest()
	req.BillingCycle = cycle
	req.Scheme = "https"
	req.PageSize = requests.NewInteger(300)
	req.IsHideZeroCharge = requests.NewBoolean(true) //过滤掉原价为0

	for {

		if info != nil {
			for atomic.LoadInt32(&cb.historyFlag) == 1 {
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

		resp, err := cb.ag.queryBillWrap(ctx, req)
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

	return nil
}

// func (cb *costBill) getInstnceBills(ctx context.Context, year, month int) error {

// 	start := time.Now().Truncate(time.Hour)

// 	for d := 1; d <= start.Day(); d++ {

// 		req := bssopenapi.CreateQueryInstanceBillRequest()
// 		req.BillingCycle = fmt.Sprintf("%d-%02d", year, month)
// 		req.Scheme = "https"
// 		req.Granularity = "DAILY"
// 		req.IsBillingItem = requests.NewBoolean(false) //按实例
// 		req.PageSize = requests.NewInteger(300)
// 		req.BillingDate = fmt.Sprintf("%d-%02d-%02d", year, month, d)

// 		for {

// 			select {
// 			case <-ctx.Done():
// 				return context.Canceled
// 			default:
// 			}

// 			resp, err := cb.runningInstance.QueryInstanceBillWrap(ctx, req)

// 			select {
// 			case <-ctx.Done():
// 				return context.Canceled
// 			default:
// 			}

// 			if err != nil {
// 				return fmt.Errorf("fail to get instance bill of %s: %s", req.BillingDate, err)
// 			} else {
// 				cb.logger.Debugf("InstnceBills(%s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", req.BillingDate, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))

// 				if err := cb.parseInstanceBillResponse(ctx, resp); err != nil {
// 					return err
// 				}

// 				if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
// 					req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
// 				} else {
// 					break
// 				}
// 			}
// 		}
// 	}

// 	return nil
// }

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
			if cb.ag.debugMode {
				//data, _ := io.MakeMetric(cb.getName(), tags, fields, t)
				//fmt.Printf("-----%s\n", string(data))
			} else {
				io.NamedFeedEx(inputName, io.Metric, cb.getName(), tags, fields, t)
			}
		}
	}

	return nil
}

// func (cb *costBill) parseInstanceBillResponse(ctx context.Context, resp *bssopenapi.QueryInstanceBillResponse) error {
// 	for _, item := range resp.Data.Items.Item {
// 		tags := map[string]string{
// 			"AccountID":    resp.Data.AccountID,
// 			"AccountName":  resp.Data.AccountName,
// 			"OwnerID":      item.OwnerID,
// 			"BillingCycle": resp.Data.BillingCycle,
// 		}

// 		tags[`CostUnit`] = item.CostUnit
// 		tags[`SubscriptionType`] = item.SubscriptionType
// 		tags[`Item`] = item.Item
// 		tags[`ProductCode`] = item.ProductCode
// 		tags[`ProductName`] = item.ProductName
// 		tags[`ProductType`] = item.ProductType
// 		tags[`InstanceID`] = item.InstanceID
// 		tags[`NickName`] = item.NickName
// 		tags[`InstanceSpec`] = item.InstanceSpec
// 		tags[`InternetIP`] = item.InternetIP
// 		tags[`Region`] = item.Region
// 		tags[`Zone`] = item.Zone
// 		tags[`BillingItem`] = item.BillingItem
// 		tags[`Currency`] = item.Currency

// 		if item.Tag != "" {
// 			kvs := strings.Split(item.Tag, `;`)
// 			for _, kv := range kvs {
// 				parts := strings.Split(kv, ` `)
// 				if len(parts) != 2 {
// 					continue
// 				}
// 				k := parts[0]
// 				v := parts[1]
// 				var key, val string
// 				pos := strings.Index(k, `key:`)
// 				if pos != -1 {
// 					key = k[4:]
// 				}
// 				pos = strings.Index(v, `value:`)
// 				if pos != -1 {
// 					val = v[6:]
// 				}
// 				if key != "" {
// 					tags[key] = val
// 				}
// 			}
// 		}

// 		fields := map[string]interface{}{}

// 		fields[`PretaxGrossAmount`] = item.PretaxGrossAmount
// 		fields[`InvoiceDiscount`] = item.InvoiceDiscount
// 		fields[`DeductedByCoupons`] = item.DeductedByCoupons
// 		fields[`PretaxAmount`] = item.PretaxAmount
// 		fields[`DeductedByCashCoupons`] = item.DeductedByCashCoupons
// 		fields[`DeductedByPrepaidCard`] = item.DeductedByPrepaidCard
// 		fields[`PaymentAmount`] = item.PaymentAmount
// 		fields[`OutstandingAmount`] = item.OutstandingAmount
// 		fields[`BillingDate`] = item.BillingDate

// 		io.NamedFeedEx(inputName, io.Metric, cb.getName(), tags, fields)
// 	}
// 	return nil
// }

func (cb *costBill) getInterval() time.Duration {
	return cb.interval
}

func (cb *costBill) getName() string {
	return cb.measurement
}
