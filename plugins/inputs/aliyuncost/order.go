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

type costOrder struct {
	interval    time.Duration
	measurement string
	ag          *agent

	historyFlag int32
}

func newCostOrder(ag *agent) *costOrder {
	c := &costOrder{
		ag:          ag,
		measurement: "aliyun_cost_order",
		interval:    ag.OrdertInterval.Duration,
	}
	return c
}

func (co *costOrder) run(ctx context.Context) {

	if co.ag.CollectHistoryData && !co.ag.isTest() {
		go func() {
			time.Sleep(time.Millisecond * 10)
			co.getHistoryData(ctx)
		}()
	}

	co.getData(ctx)

	moduleLogger.Info("order done")
}

func (co *costOrder) getData(ctx context.Context) {

	for {
		//暂停历史数据抓取
		atomic.AddInt32(&co.historyFlag, 1)

		start := time.Now().UTC()

		endTime := start.Truncate(time.Minute)

		shift := time.Hour
		if co.interval > shift {
			shift = co.interval
		}
		shift += 5 * time.Minute

		from := unixTimeStr(endTime.Add(-shift))
		end := unixTimeStr(endTime)

		if err := co.getOrders(ctx, from, end, nil); err != nil {
			moduleLogger.Errorf("%s", err)
			if co.ag.isTest() {
				return
			}
		} else {
			//lastTime = endTime.Add(-shift)
		}

		if co.ag.isTest() {
			break
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		usage := time.Now().UTC().Sub(start)
		if co.interval > usage {
			atomic.AddInt32(&co.historyFlag, -1)
			datakit.SleepContext(ctx, co.interval-usage)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (co *costOrder) getHistoryData(ctx context.Context) {

	key := "." + co.ag.cacheFileKey(`ordersV2`)

	if !co.ag.CollectHistoryData {
		return
	}

	moduleLogger.Info("start getting history Orders")

	var info *historyInfo
	if !co.ag.isDebug() {
		info, _ = getAliyunCostHistory(key)
	}

	if info == nil {
		info = &historyInfo{}
	} else if info.Statue == 1 {
		moduleLogger.Infof("already fetched the history data")
		return
	}

	if info.Start == "" {
		now := time.Now().UTC().Truncate(time.Minute)
		start := now.Add(-time.Hour * 8760)
		info.Start = unixTimeStr(start)
		info.End = unixTimeStr(now)
		info.Statue = 0
		info.PageNum = 1
	}

	info.key = key

	if err := co.getOrders(ctx, info.Start, info.End, info); err != nil && err != context.Canceled {
		moduleLogger.Errorf("fail to get orders of history(%s-%s): %s", info.Start, info.End, err)
	}
}

func (co *costOrder) getOrders(ctx context.Context, start, end string, info *historyInfo) error {

	defer func() {
		if e := recover(); e != nil {
			moduleLogger.Errorf("panic, %v", e)
		}
	}()

	logPrefix := ""
	if info != nil {
		logPrefix = "(history) "
	}

	moduleLogger.Infof("%sOrders(%s - %s) start", logPrefix, start, end)

	req := bssopenapi.CreateQueryOrdersRequest()
	req.Scheme = "https"
	if start != "" {
		req.CreateTimeStart = start
	}
	if end != "" {
		req.CreateTimeEnd = end
	}
	req.PageSize = requests.NewInteger(100)
	if info != nil {
		req.PageNum = requests.NewInteger(info.PageNum)
	} else {
		req.PageNum = requests.NewInteger(1)
	}

	for {
		if info != nil {
			for atomic.LoadInt32(&co.historyFlag) > 0 {
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

		resp, err := co.ag.queryOrders(ctx, req)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			return fmt.Errorf("%sfail to get Orders(%s - %s)", logPrefix, start, end)
		}

		moduleLogger.Debugf(" %sPage%d: count=%d, TotalCount=%d, PageSize=%d", logPrefix, resp.Data.PageNum, len(resp.Data.OrderList.Order), resp.Data.TotalCount, resp.Data.PageSize)

		if err = co.parseOrderResponse(ctx, resp); err != nil {
			return err
		}

		if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
			req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
			if info != nil {
				info.PageNum = resp.Data.PageNum + 1
				if !co.ag.isDebug() {
					setAliyunCostHistory(info.key, info)
				}
			}
		} else {
			break
		}
	}

	moduleLogger.Debugf("%sOrders(%s - %s) end", logPrefix, start, end)

	if info != nil {
		info.Statue = 1
		if !co.ag.isDebug() {
			setAliyunCostHistory(info.key, info)
		}
	}

	return nil
}

func (co *costOrder) parseOrderResponse(ctx context.Context, resp *bssopenapi.QueryOrdersResponse) error {

	for _, item := range resp.Data.OrderList.Order {

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		tags := map[string]string{
			"ProductCode":      item.ProductCode,
			"ProductType":      item.ProductType,
			"SubscriptionType": item.SubscriptionType,
			"OrderType":        item.OrderType,
			"Currency":         item.Currency,
			"AccountName":      co.ag.accountName,
			"AccountID":        co.ag.accountID,
		}

		fields := map[string]interface{}{
			"OrderID":        item.OrderId,
			"RelatedOrderId": item.RelatedOrderId,
		}

		fields["PretaxGrossAmount"], _ = strconv.ParseFloat(datakit.NumberFormat(item.PretaxGrossAmount), 64)
		fields["PretaxAmount"], _ = strconv.ParseFloat(datakit.NumberFormat(item.PretaxAmount), 64)

		// reqDetail := bssopenapi.CreateGetOrderDetailRequest()
		// reqDetail.OrderId = item.OrderId

		// respDetail, err := co.runningInstance.client.GetOrderDetail(reqDetail)
		// if err != nil {
		// 	co.logger.Warnf("fail to get order detail of %s, %s", item.OrderId, err)
		// } else {
		// 	fields["OutstandingAmount"] = respDetail.Data.OutstandingAmount

		// }

		t, err := time.Parse(time.RFC3339, item.CreateTime)
		if err != nil {
			moduleLogger.Warnf("fail to parse time:%v, error:%s", item.CreateTime, err)
			continue
		}
		if co.ag.isTest() {
			// pass
		} else if co.ag.isDebug() {
			//data, _ := io.MakeMetric(co.getName(), tags, fields, t)
			//fmt.Printf("-----%s\n", string(data))
		} else {
			io.NamedFeedEx(inputName, datakit.Metric, co.getName(), tags, fields, t)
		}
	}

	return nil
}

func (co *costOrder) getInterval() time.Duration {
	return co.interval
}

func (co *costOrder) getName() string {
	return co.measurement
}
