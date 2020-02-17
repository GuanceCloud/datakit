package aliyuncost

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"github.com/influxdata/telegraf/selfstat"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/limiter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
)

type CostAccount struct {
	interval        time.Duration
	name            string
	runningInstance *RunningInstance
	logger          *models.Logger
}

func NewCostAccount(cfg *CostCfg, ri *RunningInstance) *CostAccount {
	c := &CostAccount{
		name:            "aliyun_cost_account",
		interval:        cfg.AccountInterval.Duration,
		runningInstance: ri,
	}
	c.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `aliyuncost:account`,
	}
	return c
}

func (ca *CostAccount) run(ctx context.Context) error {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		ca.getHistoryData(ctx, ca.runningInstance.lmtr)
	}()

	ca.getRealtimeData(ctx)

	wg.Wait()

	ca.logger.Info("done")

	return nil
}

func (ca *CostAccount) getRealtimeData(ctx context.Context) error {

	for {
		//暂停历史数据抓取
		ca.runningInstance.suspendHistoryFetch()
		now := time.Now().Truncate(time.Minute)
		start := now.Add(-ca.interval)

		from := unixTimeStr(start)
		end := unixTimeStr(now)
		if err := ca.getTransactions(ctx, from, end, ca.runningInstance.lmtr, nil); err != nil && err != context.Canceled {
			ca.logger.Errorf("%s", err)
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		ca.runningInstance.resumeHistoryFetch()
		internal.SleepContext(ctx, ca.interval)

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}
	}
}

func (ca *CostAccount) getHistoryData(ctx context.Context, lmtr *limiter.RateLimiter) error {

	key := "." + ca.runningInstance.cacheFileKey(`account`)

	if !ca.runningInstance.cfg.CollectHistoryData {
		DelAliyunCostHistory(key)
		return nil
	}

	ca.logger.Info("start getting history Transactions")

	info, _ := GetAliyunCostHistory(key)

	if info == nil {
		info = &historyInfo{}
	} else if info.Statue == 1 {
		return nil
	}

	if info.Start == "" {
		now := time.Now().Truncate(time.Minute)
		start := now.Add(-time.Hour * 8760)
		info.Start = unixTimeStr(start)
		info.End = unixTimeStr(now)
		info.Statue = 0
		info.PageNum = 0
	}

	info.key = key

	if err := ca.getTransactions(ctx, info.Start, info.End, lmtr, info); err != nil && err != context.Canceled {
		ca.logger.Errorf("fail to get account transactions of history(%s-%s): %s", info.Start, info.End, err)
		return err
	}

	return nil
}

//获取账户流水收支明细查询
//https://www.alibabacloud.com/help/zh/doc-detail/118472.htm?spm=a2c63.p38356.b99.35.457e524dKiDhkZ
func (ca *CostAccount) getTransactions(ctx context.Context, start, end string, lmtr *limiter.RateLimiter, info *historyInfo) error {

	defer func() {
		recover()
	}()

	//账户余额查询接口
	//https://www.alibabacloud.com/help/zh/doc-detail/87997.htm?spm=a2c63.p38356.b99.34.717125earDjnGX
	if info != nil {
		ca.logger.Info("(history) start getting Balance")
	} else {
		ca.logger.Info("start getting Balance")
	}
	balanceReq := bssopenapi.CreateQueryAccountBalanceRequest()
	//balanceReq.Scheme = "https"
	balanceResp, err := ca.runningInstance.client.QueryAccountBalance(balanceReq)
	if err != nil {
		ca.logger.Errorf("fail to get balance, %s", err)
		return err
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	if info != nil {
		ca.logger.Infof("(history)start getting Transactions(%s - %s)", start, end)
	} else {
		ca.logger.Infof("start getting Transactions(%s - %s)", start, end)
	}

	var respTransactions *bssopenapi.QueryAccountTransactionsResponse

	req := bssopenapi.CreateQueryAccountTransactionsRequest()
	req.CreateTimeStart = start
	req.CreateTimeEnd = end
	req.PageSize = requests.NewInteger(300)
	if info != nil {
		req.PageNum = requests.NewInteger(info.PageNum)
	}

	for { //分页
		if info != nil {
			ca.runningInstance.wait()
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		case <-lmtr.C:
		}

		resp, err := ca.runningInstance.client.QueryAccountTransactions(req)
		if err != nil {
			return fmt.Errorf("fail to get transactions from %s to %s: %s", start, end, err)
		}
		if info != nil {
			ca.logger.Debugf("(history)Transactions(%s - %s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", start, end, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.AccountTransactionsList.AccountTransactionsListItem))
		} else {
			ca.logger.Debugf("Transactions(%s - %s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", start, end, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.AccountTransactionsList.AccountTransactionsListItem))

		}

		if respTransactions == nil {
			respTransactions = resp
		} else {
			//累积
			respTransactions.Data.AccountTransactionsList.AccountTransactionsListItem = append(respTransactions.Data.AccountTransactionsList.AccountTransactionsListItem, resp.Data.AccountTransactionsList.AccountTransactionsListItem...)
		}

		if info != nil {
			if err := ca.parseTransactionsResponse(ctx, balanceResp, respTransactions); err != nil {
				return err
			}
			respTransactions.Data.AccountTransactionsList.AccountTransactionsListItem = respTransactions.Data.AccountTransactionsList.AccountTransactionsListItem[:0]
		}

		if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
			//有后续页
			req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
			if info != nil {
				info.PageNum = resp.Data.PageNum + 1
				SetAliyunCostHistory(info.key, info)
			}
		} else {
			if info != nil {
				info.Statue = 1
				SetAliyunCostHistory(info.key, info)
			}
			break
		}

	}

	if info != nil {
		ca.logger.Debugf("(history)finish getting Transactions(%s - %s), count=%v", start, end, len(respTransactions.Data.AccountTransactionsList.AccountTransactionsListItem))
	} else {
		ca.logger.Debugf("finish getting Transactions(%s - %s), count=%v", start, end, len(respTransactions.Data.AccountTransactionsList.AccountTransactionsListItem))
	}

	return ca.parseTransactionsResponse(ctx, balanceResp, respTransactions)
}

func (ca *CostAccount) parseTransactionsResponse(ctx context.Context, balanceResp *bssopenapi.QueryAccountBalanceResponse, resp *bssopenapi.QueryAccountTransactionsResponse) error {

	balanceTags := map[string]string{}
	balanceFields := map[string]interface{}{}

	balanceTags[`Currency`] = balanceResp.Data.Currency

	var fv float64
	var err error
	if fv, err = strconv.ParseFloat(internal.NumberFormat(balanceResp.Data.AvailableAmount), 64); err == nil {
		balanceFields[`AvailableAmount`] = fv
	}
	if fv, err = strconv.ParseFloat(internal.NumberFormat(balanceResp.Data.MybankCreditAmount), 64); err == nil {
		balanceFields[`MybankCreditAmount`] = fv
	}
	if fv, err = strconv.ParseFloat(internal.NumberFormat(balanceResp.Data.AvailableCashAmount), 64); err == nil {
		balanceFields[`AvailableCashAmount`] = fv
	}
	if fv, err = strconv.ParseFloat(internal.NumberFormat(balanceResp.Data.CreditAmount), 64); err == nil {
		balanceFields[`CreditAmount`] = fv
	}

	accname := resp.Data.AccountName
	accid := resp.Data.AccountID
	for _, item := range resp.Data.AccountTransactionsList.AccountTransactionsListItem {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		tags := map[string]string{}
		for k, v := range balanceTags {
			tags[k] = v
		}

		fields := map[string]interface{}{}
		for k, v := range balanceFields {
			fields[k] = v
		}

		tags[`TransactionFlow`] = item.TransactionFlow
		tags[`TransactionType`] = item.TransactionType
		tags[`TransactionChannel`] = item.TransactionChannel
		tags[`FundType`] = item.FundType
		tags[`TransactionAccount`] = item.TransactionAccount
		tags[`AccountName`] = accname
		tags[`AccountID`] = accid

		fields[`TransactionNumber`] = item.TransactionNumber
		fields[`TransactionChannelSN`] = item.TransactionChannelSN
		fields[`RecordID`] = item.RecordID
		fields[`Remarks`] = item.Remarks
		fields[`Amount`], _ = strconv.ParseFloat(internal.NumberFormat(item.Amount), 64)
		fields[`Balance`], _ = strconv.ParseFloat(internal.NumberFormat(item.Balance), 64)
		tm, err := time.Parse("2006-01-02T15:04:05Z", item.TransactionTime)
		if err == nil {
			fields[`TransactionTime`] = tm.Unix()
		}

		if ca.runningInstance.cost.accumulator != nil {
			ca.runningInstance.cost.accumulator.AddFields(ca.getName(), fields, tags)
		}
	}

	return nil
}

func (ca *CostAccount) getInterval() time.Duration {
	return ca.interval
}

func (ca *CostAccount) getName() string {
	return ca.name
}
