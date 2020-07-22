package aliyuncost

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type CostAccount struct {
	interval        time.Duration
	name            string
	runningInstance *runningInstance
	logger          *logger.Logger
}

func NewCostAccount(cfg *CostCfg, ri *runningInstance) *CostAccount {
	c := &CostAccount{
		name:            "aliyun_cost_account",
		interval:        cfg.AccountInterval.Duration,
		runningInstance: ri,
	}
	c.logger = logger.SLogger(`aliyuncost:account`)
	return c
}

func (ca *CostAccount) run(ctx context.Context) error {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond * 10)
		ca.getHistoryData(ctx)
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
		start := now.Add(-time.Hour * 24)
		from := unixTimeStr(start) //需要传unix时间

		end := unixTimeStr(now)
		if err := ca.getTransactions(ctx, from, end, nil); err != nil && err != context.Canceled {
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

func (ca *CostAccount) getHistoryData(ctx context.Context) error {

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
		ca.logger.Infof("already fetched the history data")
		return nil
	}

	if info.Start == "" {
		now := time.Now().Truncate(time.Minute)
		start := now.Add(-time.Hour * 8760)
		info.Start = unixTimeStr(start)
		info.End = unixTimeStr(now)
		info.Statue = 0
		info.PageNum = 1
	}

	info.key = key

	if err := ca.getTransactions(ctx, info.Start, info.End, info); err != nil && err != context.Canceled {
		ca.logger.Errorf("fail to get account transactions of history(%s-%s): %s", info.Start, info.End, err)
		return err
	}

	return nil
}

//获取账户流水收支明细查询
//https://www.alibabacloud.com/help/zh/doc-detail/118472.htm?spm=a2c63.p38356.b99.35.457e524dKiDhkZ
func (ca *CostAccount) getTransactions(ctx context.Context, start, end string, info *historyInfo) error {

	defer func() {
		if e := recover(); e != nil {
			ca.logger.Errorf("panic, %v", e)
		}
	}()

	logPrefix := ""
	if info != nil {
		logPrefix = "(history) "
	}

	//账户余额查询接口
	//https://www.alibabacloud.com/help/zh/doc-detail/87997.htm?spm=a2c63.p38356.b99.34.717125earDjnGX
	ca.logger.Infof("%sstart getting Balance", logPrefix)

	balanceReq := bssopenapi.CreateQueryAccountBalanceRequest()
	balanceReq.Scheme = "https"
	balanceResp, err := ca.runningInstance.QueryAccountBalanceWrap(ctx, balanceReq)
	if err != nil {
		ca.logger.Errorf("%sfail to get balance, %s", logPrefix, err)
		return err
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	ca.logger.Infof("%sstart getting Transactions(%s - %s)", logPrefix, start, end)

	req := bssopenapi.CreateQueryAccountTransactionsRequest()
	req.CreateTimeStart = start
	req.CreateTimeEnd = end
	req.PageSize = requests.NewInteger(300)
	if info != nil {
		req.PageNum = requests.NewInteger(info.PageNum)
	} else {
		req.PageNum = requests.NewInteger(1)
	}

	for { //分页
		if info != nil {
			ca.runningInstance.wait()
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		resp, err := ca.runningInstance.QueryAccountTransactionsWrap(ctx, req)
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}
		if err != nil {
			return fmt.Errorf("%sfail to get transactions from %s to %s: %s", logPrefix, start, end, err)
		}

		ca.logger.Debugf("%sTransactions(%s - %s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", logPrefix, start, end, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.AccountTransactionsList.AccountTransactionsListItem))

		if err := ca.parseTransactionsResponse(ctx, balanceResp, resp); err != nil {
			return err
		}

		if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
			//有后续页
			req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
			if info != nil {
				info.PageNum = resp.Data.PageNum + 1
				SetAliyunCostHistory(info.key, info)
			}
		} else {

			break
		}

	}

	ca.logger.Debugf("%sfinish getting Transactions(%s - %s)", logPrefix, start, end)

	if info != nil {
		info.Statue = 1
		SetAliyunCostHistory(info.key, info)
	}

	return nil
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
	if accid == "" {
		accid = ca.runningInstance.accountID
	}
	if accname == "" {
		accname = ca.runningInstance.accountName
	}

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

		tags[`TransactionFlow`] = item.TransactionFlow //收支类型
		tags[`BillingCycle`] = item.BillingCycle
		tags[`TransactionType`] = item.TransactionType       //交易类型
		tags[`TransactionChannel`] = item.TransactionChannel //交易渠道
		tags[`FundType`] = item.FundType                     //资金形式
		tags[`TransactionAccount`] = item.TransactionAccount
		tags[`AccountName`] = accname
		tags[`AccountID`] = accid

		fields[`TransactionNumber`] = item.TransactionNumber //交易编号
		fields[`TransactionChannelSN`] = item.TransactionChannelSN
		fields[`RecordID`] = item.RecordID                                                 //订单号/账单号
		fields[`Remarks`] = item.Remarks                                                   //交易备注
		fields[`Amount`], _ = strconv.ParseFloat(internal.NumberFormat(item.Amount), 64)   //交易金额
		fields[`Balance`], _ = strconv.ParseFloat(internal.NumberFormat(item.Balance), 64) //交易发生前当前余额

		tm, err := time.Parse("2006-01-02T15:04:05Z", item.TransactionTime)
		if err != nil {
			ca.logger.Warnf("fail to parse time:%v %s, error: %s", item.TransactionTime, item.RecordID, err)
		} else {
			tm = tm.Add(-8 * time.Hour) //返回的不是unix时间字符串
			io.NamedFeedEx(inputName, io.Metric, ca.getName(), tags, fields, tm)
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
