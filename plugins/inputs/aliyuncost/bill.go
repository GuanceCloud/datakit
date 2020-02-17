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
	"github.com/influxdata/telegraf/selfstat"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/limiter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
)

type CostBill struct {
	interval        time.Duration
	name            string
	runningInstance *RunningInstance
	logger          *models.Logger
}

func NewCostBill(cfg *CostCfg, ri *RunningInstance) *CostBill {
	c := &CostBill{
		name:            "aliyun_cost_bill",
		interval:        cfg.BiilInterval.Duration,
		runningInstance: ri,
	}
	c.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `aliyuncost:bill`,
	}
	return c
}

func (cb *CostBill) run(ctx context.Context) error {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		//b.getLastyearData(ctx, b.runningInstance.lmtr)
	}()

	cb.getRealtimeData(ctx)

	wg.Wait()

	cb.logger.Info("done")

	return nil
}

func (cb *CostBill) getRealtimeData(ctx context.Context) error {

	for {
		cb.runningInstance.suspendHistoryFetch()
		start := time.Now().Truncate(time.Minute)
		cycle := fmt.Sprintf("%d-%02d", start.Year(), start.Month())
		if err := cb.getBills(ctx, cycle, cb.runningInstance.lmtr, nil); err != nil && err != context.Canceled {
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

func (cb *CostBill) getLastyearData(ctx context.Context, lmtr *limiter.RateLimiter) error {

	if !cb.runningInstance.cfg.CollectHistoryData {
		return nil
	}

	// log.Printf("I! [aliyunboa:bill] start get bills of last year")

	// m := md5.New()
	// m.Write([]byte(b.runningInstance.cfg.AccessKeyID))
	// m.Write([]byte(b.runningInstance.cfg.AccessKeySecret))
	// m.Write([]byte(b.runningInstance.cfg.RegionID))
	// m.Write([]byte(`bills`))
	// k1 := hex.EncodeToString(m.Sum(nil))
	// k1 = "." + k1

	// billFlag, _ := config.GetLastyearFlag(k1)

	// m.Reset()
	// m.Write([]byte(b.runningInstance.cfg.AccessKeyID))
	// m.Write([]byte(b.runningInstance.cfg.AccessKeySecret))
	// m.Write([]byte(b.runningInstance.cfg.RegionID))
	// m.Write([]byte(`bills_instance`))
	// k2 := hex.EncodeToString(m.Sum(nil))
	// k2 = "." + k2

	// billInstanceFlag, _ := config.GetLastyearFlag(k2)

	// if billInstanceFlag == 1 && billFlag == 1 {
	// 	return nil
	// }

	// now := time.Now().Add(-24 * time.Hour * 30)
	// for index := 0; index < 11; index++ {
	// 	select {
	// 	case <-ctx.Done():
	// 		return context.Canceled
	// 	default:
	// 	}
	// 	cycle := fmt.Sprintf("%d-%02d", now.Year(), now.Month())

	// 	if billFlag == 0 {
	// 		if err := b.getBills(ctx, cycle, lmtr, true); err != nil {
	// 			log.Printf("E! [aliyunboa:bill] %s", err)
	// 		}
	// 	}

	// 	if billInstanceFlag == 0 {
	// 		if err := b.getInstnceBills(ctx, cycle, lmtr); err != nil {
	// 			log.Printf("E! [aliyunboa:bill] %s", err)
	// 		}
	// 	}

	// 	now = now.Add(-24 * time.Hour * 30)
	// }

	// config.SetLastyearFlag(k1, 1)
	// config.SetLastyearFlag(k2, 1)

	return nil
}

func (cb *CostBill) getBills(ctx context.Context, cycle string, lmtr *limiter.RateLimiter, info *historyInfo) error {

	defer func() {
		recover()
	}()

	cb.logger.Infof("start getting Bills of %s", cycle)

	var respBill *bssopenapi.QueryBillResponse

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
		case <-lmtr.C:
			//
		}

		resp, err := cb.runningInstance.client.QueryBill(req)
		if err != nil {
			return fmt.Errorf("fail to get bill of %s: %s", cycle, err)
		}

		cb.logger.Debugf("Bills(%s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", cycle, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))

		if respBill == nil {
			respBill = resp
		} else {
			respBill.Data.Items.Item = append(respBill.Data.Items.Item, resp.Data.Items.Item...)
		}

		if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
			req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
		} else {
			break
		}
	}

	cb.logger.Infof("finish getting Bill(%s), count=%d", cycle, len(respBill.Data.Items.Item))

	return cb.parseBillResponse(ctx, respBill)
}

func (cb *CostBill) getInstnceBills(ctx context.Context, cycle string, lmtr *limiter.RateLimiter) error {

	//var respInstill *bssopenapi.QueryInstanceBillResponse

	req := bssopenapi.CreateQueryInstanceBillRequest()
	req.BillingCycle = cycle
	req.Scheme = "https"
	req.PageSize = requests.NewInteger(300)

	for {

		select {
		case <-ctx.Done():
			return context.Canceled
		case <-lmtr.C:
		}

		resp, err := cb.runningInstance.client.QueryInstanceBill(req)
		if err != nil {
			return fmt.Errorf("fail to get instance bill of %s: %s", cycle, err)
		} else {
			cb.logger.Debugf("InstnceBills(%s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", cycle, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))
			// if respInstill == nil {
			// 	respInstill = resp
			// } else {
			// 	if resp.Data.TotalCount > 0 {
			// 		respInstill.Data.Items.Item = append(respInstill.Data.Items.Item, resp.Data.Items.Item...)
			// 	}
			// }

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
			"AccountID":   resp.Data.AccountID,
			"AccountName": resp.Data.AccountName,
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
		fields[`RoundDownDiscount`], _ = strconv.ParseFloat(item.RoundDownDiscount, 64)
		fields[`PretaxAmount`] = item.PretaxAmount
		fields[`DeductedByCashCoupons`] = item.DeductedByCashCoupons
		fields[`DeductedByPrepaidCard`] = item.DeductedByPrepaidCard
		fields[`PaymentAmount`] = item.PaymentAmount
		fields[`OutstandingAmount`] = item.OutstandingAmount

		billtime := item.UsageEndTime
		t, err := time.Parse(`2006-01-02 15:04:05`, billtime)
		if err != nil {
			cb.logger.Warnf("fail to parse time:%v of product:%s, error: %s", billtime, item.ProductName, err)
		} else {
			if cb.runningInstance.cost.accumulator != nil {
				cb.runningInstance.cost.accumulator.AddFields(cb.getName(), fields, tags, t)
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

		if cb.runningInstance.cost.accumulator != nil {
			cb.runningInstance.cost.accumulator.AddFields(cb.getName(), fields, tags)
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
