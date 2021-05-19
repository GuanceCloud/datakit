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

type transaction struct {
	interval    time.Duration
	rateLimiter *rate.Limiter
	agent       *TencentCost
	client      *billing.Client

	historyFlag int32
}

func newTransaction(ag *TencentCost) *transaction {
	b := &transaction{
		agent:    ag,
		interval: ag.TransactionInterval.Duration,
	}
	return b
}

func (b *transaction) getName() string {
	return "tencent_cost_transaction"
}

func (b *transaction) run(ctx context.Context) {
	limit := rate.Every(60 * time.Millisecond)
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

func (b *transaction) getData(ctx context.Context) {

	var lastTime time.Time
	shift := time.Minute

	for {

		atomic.AddInt32(&b.historyFlag, 1)
		start := time.Now()

		endTime := time.Now().Truncate(time.Minute)
		if lastTime.IsZero() {
			lastTime = endTime.Add(-b.interval)
		}

		err := b.describeBillList(ctx, lastTime, endTime, nil)
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
			atomic.AddInt32(&b.historyFlag, -1)
			datakit.SleepContext(ctx, b.interval-usage)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (b *transaction) getHistoryData(ctx context.Context) {

	key := "." + cacheFileKey(`transactions`, b.agent.AccessKeyID, b.agent.AccessKeySecret)

	moduleLogger.Info("start getting history Transactions")

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

	b.describeBillList(ctx, info.StartTime, info.EndTime, info)
}

//https://cloud.tencent.com/document/api/555/41874
func (b *transaction) describeBillList(ctx context.Context, startTime time.Time, endTime time.Time, history *historyInfo) error {

	logPrefix := ""
	if history != nil {
		logPrefix = "(history) "
	}

	var err error
	var response *billing.DescribeBillListResponse
	var page uint64 = 100
	var offset uint64 = 0
	if history != nil {
		offset = history.Offset
	}

	st := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", startTime.Year(), startTime.Month(), startTime.Day(), startTime.Hour(), startTime.Minute(), startTime.Second())
	et := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", endTime.Year(), endTime.Month(), endTime.Day(), endTime.Hour(), endTime.Minute(), endTime.Second())

	moduleLogger.Infof("%sgetting Transactions(%s - %s)", logPrefix, st, et)

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

			request := billing.NewDescribeBillListRequest()
			request.StartTime = &st
			request.EndTime = &et
			request.Offset = &offset
			request.Limit = &page

			response, err = b.client.DescribeBillList(request)

			if err != nil {
				if _, ok := err.(*errors.TencentCloudSDKError); ok {
					moduleLogger.Errorf("An API error has returned: %s", err)
				} else {
					moduleLogger.Errorf("%s", err)
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

		moduleLogger.Debugf("%sTransactions(%s - %s): TotalCount=%d, Offset=%d, count=%d", logPrefix, st, et, *response.Response.Total, offset, len(response.Response.TransactionList))

		b.handleAccountBalance(ctx, response)

		offset += uint64(len(response.Response.TransactionList))

		if history != nil {
			history.Offset = offset
			setAliyunCostHistory(history.key, history)
		}

		if len(response.Response.TransactionList) < 100 {
			break
		}
	}

	moduleLogger.Debugf("%sfinish getting Transactions(%s - %s)", logPrefix, st, et)

	if history != nil {
		history.Statue = 1
		setAliyunCostHistory(history.key, history)
	}

	return err
}

func (b *transaction) handleAccountBalance(ctx context.Context, response *billing.DescribeBillListResponse) {
	if response == nil {
		return
	}

	for _, item := range response.Response.TransactionList {
		tags := map[string]string{
			"ActionType": ensureString(item.ActionType),
			"PayChannel": ensureString(item.PayChannel),
			"DeductMode": ensureString(item.DeductMode),
		}
		b.agent.appendCustomTags(tags)

		fields := map[string]interface{}{
			"BillId":        ensureString(item.BillId),
			"OperationInfo": ensureString(item.OperationInfo),
		}

		if item.Balance != nil {
			fields["Balance"] = *item.Balance
		}
		if item.Amount != nil {
			fields["Amount"] = *item.Amount
		}
		if item.Cash != nil {
			fields["Cash"] = *item.Cash
		}
		if item.Incentive != nil {
			fields["Incentive"] = *item.Incentive
		}
		if item.Freezing != nil {
			fields["Freezing"] = *item.Freezing
		}

		metrictime := time.Now().UTC()
		if item.OperationTime != nil {
			tm, e := time.ParseInLocation(`2006-01-02 15:04:05`, *item.OperationTime, time.Local)
			if e != nil {
				moduleLogger.Warnf("fail to parse time, %s", e)
			} else {
				metrictime = tm.UTC()
			}

		}

		io.NamedFeedEx(inputName, datakit.Metric, b.getName(), tags, fields, metrictime)
	}
}
