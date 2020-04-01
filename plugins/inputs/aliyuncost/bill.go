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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
)

type CostBill struct {
	interval        time.Duration
	name            string
	runningInstance *runningInstance
	logger          *models.Logger
}

func NewCostBill(cfg *CostCfg, ri *runningInstance) *CostBill {
	c := &CostBill{
		name:            "aliyun_cost_bill",
		interval:        cfg.BiilInterval.Duration,
		runningInstance: ri,
	}
	c.logger = &models.Logger{
		Name: `aliyuncost:bill`,
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

	cb.getRealtimeData(ctx)

	wg.Wait()

	cb.logger.Info("done")

	return nil
}

func (cb *CostBill) getRealtimeData(ctx context.Context) error {

	for {
		cb.runningInstance.suspendHistoryFetch()
		//以月为单位
		start := time.Now().Truncate(time.Minute)
		cycle := fmt.Sprintf("%d-%02d", start.Year(), start.Month())
		if err := cb.getBills(ctx, cycle, nil); err != nil && err != context.Canceled {
			cb.logger.Errorf("%s", err)
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
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
			return fmt.Errorf("fail to get bill of %s: %s", cycle, err)
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

func (cb *CostBill) getInstnceBills(ctx context.Context, cycle string) error {

	//var respInstill *bssopenapi.QueryInstanceBillResponse

	req := bssopenapi.CreateQueryInstanceBillRequest()
	req.BillingCycle = cycle
	req.Scheme = "https"
	req.PageSize = requests.NewInteger(300)

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
			return fmt.Errorf("fail to get instance bill of %s: %s", cycle, err)
		} else {
			cb.logger.Debugf("InstnceBills(%s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", cycle, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))

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

	return nil // cb.parseInstanceBillResponse(ctx, respInstill)
}

func (cb *CostBill) parseBillResponse(ctx context.Context, resp *bssopenapi.QueryBillResponse) error {

	for _, item := range resp.Data.Items.Item {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		tags := map[string]string{
			"AccountID":   cb.runningInstance.accountID,
			"AccountName": cb.runningInstance.accountName,
			"OwnerID":     item.OwnerID,
		}

		tags["Item"] = item.Item
		tags["ProductCode"] = item.ProductCode
		tags["ProductName"] = item.ProductName
		tags["ProductType"] = item.ProductType
		tags["SubscriptionType"] = item.SubscriptionType
		tags["Status"] = item.Status
		tags[`Currency`] = item.Currency

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
			if cb.runningInstance.agent.accumulator != nil {
				cb.runningInstance.agent.accumulator.AddFields(cb.getName(), fields, tags, t)
			}
		}
	}

	return nil
}

func (cb *CostBill) parseInstanceBillResponse(ctx context.Context, resp *bssopenapi.QueryInstanceBillResponse) error {
	for _, item := range resp.Data.Items.Item {
		tags := map[string]string{
			"AccountID":   resp.Data.AccountID,
			"AccountName": resp.Data.AccountName,
		}

		tags[`OwnerID`] = item.OwnerID
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
		tags[`IntranetIP`] = item.IntranetIP
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

		if cb.runningInstance.agent.accumulator != nil {
			cb.runningInstance.agent.accumulator.AddFields(cb.getName(), fields, tags)
		}
	}
	return nil
}

func (cb *CostBill) getInterval() time.Duration {
	return cb.interval
}

func (cb *CostBill) getName() string {
	return cb.name
}
