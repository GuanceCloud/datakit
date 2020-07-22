package aliyuncost

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type CostBill struct {
	interval        time.Duration
	name            string
	runningInstance *runningInstance
	logger          *logger.Logger
}

func NewCostBill(cfg *CostCfg, ri *runningInstance) *CostBill {
	c := &CostBill{
		name:            "aliyun_cost_bill",
		interval:        cfg.BiilInterval.Duration,
		runningInstance: ri,
		logger:          logger.SLogger(`aliyuncost:bill`),
	}
	return c
}

func (cb *CostBill) run(ctx context.Context) error {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		cb.getHistoryData(ctx)
	}()

	cb.getNormalData(ctx)

	wg.Wait()

	cb.logger.Info("done")

	return nil
}

func (cb *CostBill) getNormalData(ctx context.Context) error {

	for {
		cb.runningInstance.suspendHistoryFetch()
		//以月为单位
		//计费项 + 明细
		start := time.Now().Truncate(time.Hour)
		cycle := fmt.Sprintf("%d-%02d", start.Year(), start.Month())
		if err := cb.getBills(ctx, cycle, nil); err != nil && err != context.Canceled {
			cb.logger.Errorf("%s", err)
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		//实例账单: 实例 + 按天
		if err := cb.getInstnceBills(ctx, start.Year(), int(start.Month())); err != nil && err != context.Canceled {
			cb.logger.Errorf("%s", err)
		}

		cb.runningInstance.resumeHistoryFetch()
		internal.SleepContext(ctx, cb.interval)

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}
	}

}

func (cb *CostBill) getHistoryData(ctx context.Context) error {

	key := "." + cb.runningInstance.cacheFileKey(`bill`)

	if !cb.runningInstance.cfg.CollectHistoryData {
		DelAliyunCostHistory(key)
		return nil
	}

	cb.logger.Info("start getting history Bills")

	info, _ := GetAliyunCostHistory(key)

	if info == nil {
		info = &historyInfo{}
	} else if info.Statue == 1 {
		cb.logger.Infof("already fetched the history data")
		return nil
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
			return context.Canceled
		default:
		}

		if info.StartTime.Unix() >= info.EndTime.Unix() {
			info.Statue = 1
			SetAliyunCostHistory(key, info)
			break
		}

		cycle := fmt.Sprintf("%d-%02d", info.StartTime.Year(), info.StartTime.Month())

		if err := cb.getBills(ctx, cycle, info); err != nil {
			cb.logger.Errorf("%s", err)
		}

		info.StartTime = info.StartTime.AddDate(0, 1, 0)
		SetAliyunCostHistory(key, info)
	}

	return nil
}

func (cb *CostBill) getBills(ctx context.Context, cycle string, info *historyInfo) error {

	defer func() {
		if e := recover(); e != nil {
			cb.logger.Errorf("panic: %v", e)
		}
	}()

	if info != nil {
		cb.logger.Infof("(history)start getting Bills of %s", cycle)
	} else {
		cb.logger.Infof("start getting Bills of %s", cycle)
	}

	req := bssopenapi.CreateQueryBillRequest()
	req.BillingCycle = cycle
	req.Scheme = "https"
	req.PageSize = requests.NewInteger(300)
	req.IsHideZeroCharge = requests.NewBoolean(true) //过滤掉原价为0

	for {

		if info != nil {
			cb.runningInstance.wait()
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		resp, err := cb.runningInstance.QueryBillWrap(ctx, req)
		if err != nil {
			return fmt.Errorf("fail to get bill of %s, %s", cycle, err)
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		if info != nil {
			cb.logger.Debugf("(history)Bills(%s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", cycle, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))
		} else {
			cb.logger.Debugf("Bills(%s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", cycle, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))
		}

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

func (cb *CostBill) getInstnceBills(ctx context.Context, year, month int) error {

	start := time.Now().Truncate(time.Hour)

	for d := 1; d <= start.Day(); d++ {

		req := bssopenapi.CreateQueryInstanceBillRequest()
		req.BillingCycle = fmt.Sprintf("%d-%02d", year, month)
		req.Scheme = "https"
		req.Granularity = "DAILY"
		req.IsBillingItem = requests.NewBoolean(false) //按实例
		req.PageSize = requests.NewInteger(300)
		req.BillingDate = fmt.Sprintf("%d-%02d-%02d", year, month, d)

		for {

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			resp, err := cb.runningInstance.QueryInstanceBillWrap(ctx, req)

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			if err != nil {
				return fmt.Errorf("fail to get instance bill of %s: %s", req.BillingDate, err)
			} else {
				cb.logger.Debugf("InstnceBills(%s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", req.BillingDate, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))

				if err := cb.parseInstanceBillResponse(ctx, resp); err != nil {
					return err
				}

				if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
					req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
				} else {
					break
				}
			}
		}
	}

	return nil
}

func (cb *CostBill) parseBillResponse(ctx context.Context, resp *bssopenapi.QueryBillResponse) error {

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
		fields[`RoundDownDiscount`], _ = strconv.ParseFloat(internal.NumberFormat(item.RoundDownDiscount), 64)
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
			cb.logger.Warnf("fail to parse time:%v of product:%s, error: %s", billtime, item.ProductName, err)
		} else {
			//返回的不是utc
			t = t.Add(-8 * time.Hour)
			io.NamedFeedEx(inputName, io.Metric, cb.getName(), tags, fields, t)
		}
	}

	return nil
}

func (cb *CostBill) parseInstanceBillResponse(ctx context.Context, resp *bssopenapi.QueryInstanceBillResponse) error {
	for _, item := range resp.Data.Items.Item {
		tags := map[string]string{
			"AccountID":    resp.Data.AccountID,
			"AccountName":  resp.Data.AccountName,
			"OwnerID":      item.OwnerID,
			"BillingCycle": resp.Data.BillingCycle,
		}

		tags[`CostUnit`] = item.CostUnit
		tags[`SubscriptionType`] = item.SubscriptionType
		tags[`Item`] = item.Item
		tags[`ProductCode`] = item.ProductCode
		tags[`ProductName`] = item.ProductName
		tags[`ProductType`] = item.ProductType
		tags[`InstanceID`] = item.InstanceID
		tags[`NickName`] = item.NickName
		tags[`InstanceSpec`] = item.InstanceSpec
		tags[`InternetIP`] = item.InternetIP
		tags[`Region`] = item.Region
		tags[`Zone`] = item.Zone
		tags[`BillingItem`] = item.BillingItem
		tags[`Currency`] = item.Currency

		if item.Tag != "" {
			kvs := strings.Split(item.Tag, `;`)
			for _, kv := range kvs {
				parts := strings.Split(kv, ` `)
				if len(parts) != 2 {
					continue
				}
				k := parts[0]
				v := parts[1]
				var key, val string
				pos := strings.Index(k, `key:`)
				if pos != -1 {
					key = k[4:]
				}
				pos = strings.Index(v, `value:`)
				if pos != -1 {
					val = v[6:]
				}
				if key != "" {
					tags[key] = val
				}
			}
		}

		fields := map[string]interface{}{}

		fields[`PretaxGrossAmount`] = item.PretaxGrossAmount
		fields[`InvoiceDiscount`] = item.InvoiceDiscount
		fields[`DeductedByCoupons`] = item.DeductedByCoupons
		fields[`PretaxAmount`] = item.PretaxAmount
		fields[`DeductedByCashCoupons`] = item.DeductedByCashCoupons
		fields[`DeductedByPrepaidCard`] = item.DeductedByPrepaidCard
		fields[`PaymentAmount`] = item.PaymentAmount
		fields[`OutstandingAmount`] = item.OutstandingAmount
		fields[`BillingDate`] = item.BillingDate

		io.NamedFeedEx(inputName, io.Metric, cb.getName(), tags, fields)
	}
	return nil
}

func (cb *CostBill) getInterval() time.Duration {
	return cb.interval
}

func (cb *CostBill) getName() string {
	return cb.name
}
