package aliyuncost

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type costInstanceBill struct {
	interval    time.Duration
	measurement string
	ag          *agent

	historyFlag int32
}

func newCostInstanceBill(ag *agent) *costInstanceBill {
	c := &costInstanceBill{
		ag:          ag,
		measurement: "aliyun_cost_bill",
		interval:    ag.BiilInterval.Duration,
	}
	return c
}

func (cb *costInstanceBill) run(ctx context.Context) {

	if cb.ag.CollectHistoryData && !cb.ag.isTest() {
		go func() {
			cb.getHistoryData(ctx)
		}()

	}

	cb.getData(ctx)

	moduleLogger.Info("instance bill done")
}

func (cb *costInstanceBill) getData(ctx context.Context) {

	for {

		//暂停历史数据抓取
		atomic.AddInt32(&cb.historyFlag, 1)

		select {
		case <-ctx.Done():
			return
		default:
		}

		//以月为单位
		start := time.Now().Truncate(time.Hour)
		cb.getInstnceBills(ctx, start.Year(), int(start.Month()), nil)

		atomic.AddInt32(&cb.historyFlag, -1)

		if cb.ag.isTest() {
			break
		}

		datakit.SleepContext(ctx, cb.interval)
	}

}

func (cb *costInstanceBill) getHistoryData(ctx context.Context) {

	key := "." + cb.ag.cacheFileKey(`instanceBill`)

	if !cb.ag.CollectHistoryData {
		delAliyunCostHistory(key)
		return
	}

	moduleLogger.Info("start getting history InstanceBills")

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
			setAliyunCostHistory(key, info)
			break
		}

		if err := cb.getInstnceBills(ctx, info.StartTime.Year(), int(info.StartTime.Month()), info); err != nil {
			continue
		}

		info.StartTime = info.StartTime.AddDate(0, 1, 0)
		setAliyunCostHistory(key, info)
	}
}

func (cb *costInstanceBill) getInstnceBills(ctx context.Context, year, month int, info *historyInfo) error {

	defer func() {
		if e := recover(); e != nil {
			moduleLogger.Errorf("panic: %v", e)
		}
	}()

	logPrefix := ""
	if info != nil {
		logPrefix = "(history) "
	}

	moduleLogger.Infof("%sInstanceBills of %d-%d start", logPrefix, year, month)

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

			resp, err := cb.ag.queryInstanceBill(ctx, req)
			if err != nil {
				moduleLogger.Errorf("fail to get bill of %s, error: %s", req.BillingDate, err)
				return err
			}

			select {
			case <-ctx.Done():
				return nil
			default:
			}

			moduleLogger.Debugf(" InstnceBills(%s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", req.BillingDate, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))

			cb.parseInstanceBillResponse(ctx, resp)

			if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
				req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
			} else {
				break
			}
		}
	}

	moduleLogger.Infof("%sInstanceBills of %d-%d end", logPrefix, year, month)

	return nil
}

func (cb *costInstanceBill) parseInstanceBillResponse(ctx context.Context, resp *bssopenapi.QueryInstanceBillResponse) {

	if resp == nil {
		return
	}

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
		tags[`BillingType`] = item.BillingType

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

		t, _ := time.Parse(`2006-01-02`, item.BillingDate)

		if cb.ag.isDebug() {
			data, _ := io.MakeMetric(cb.getName(), tags, fields, t)
			fmt.Printf("%s\n", string(data))
		} else {
			io.NamedFeedEx(inputName, datakit.Metric, cb.getName(), tags, fields)
		}
	}
}

func (cb *costInstanceBill) getInterval() time.Duration {
	return cb.interval
}

func (cb *costInstanceBill) getName() string {
	return cb.measurement
}
