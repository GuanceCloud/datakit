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

type costAccount struct {
	interval    time.Duration
	measurement string
	ag          *agent

	historyFlag int32
}

func newCostAccount(ag *agent) *costAccount {
	c := &costAccount{
		ag:          ag,
		measurement: "aliyun_cost_account",
		interval:    ag.AccountInterval.Duration,
	}
	return c
}

func (ca *costAccount) run(ctx context.Context) {

	if ca.ag.CollectHistoryData {
		go func() {
			ca.getHistoryData(ctx)
		}()
	}

	ca.getData(ctx)

	moduleLogger.Info("transactions done")
}

func (ca *costAccount) getData(ctx context.Context) {

	for {
		//暂停历史数据抓取
		atomic.AddInt32(&ca.historyFlag, 1)

		start := time.Now().UTC()

		endTime := start.Truncate(time.Minute)

		shift := time.Hour
		if ca.interval > shift {
			shift = ca.interval
		}
		shift += 5 * time.Minute

		from := unixTimeStr(endTime.Add(-shift)) //需要传unix时间
		end := unixTimeStr(endTime)

		if err := ca.getTransactions(ctx, from, end, nil); err != nil {
			moduleLogger.Errorf("%s", err)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		usage := time.Now().UTC().Sub(start)
		if ca.interval > usage {
			atomic.AddInt32(&ca.historyFlag, -1)
			datakit.SleepContext(ctx, ca.interval-usage)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (ca *costAccount) getHistoryData(ctx context.Context) error {

	key := "." + ca.ag.cacheFileKey(`account`)

	if !ca.ag.CollectHistoryData {
		if !ca.ag.debugMode {
			delAliyunCostHistory(key)
		}
		return nil
	}

	moduleLogger.Info("start getting history Transactions")

	var info *historyInfo

	if !ca.ag.debugMode {
		info, _ = getAliyunCostHistory(key)
	}

	if info == nil {
		info = &historyInfo{}
	} else if info.Statue == 1 {
		moduleLogger.Infof("already fetched the history data")
		return nil
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

	if err := ca.getTransactions(ctx, info.Start, info.End, info); err != nil && err != context.Canceled {
		moduleLogger.Errorf("fail to get account transactions of history(%s-%s): %s", info.Start, info.End, err)
		return err
	}

	return nil
}

//获取账户流水收支明细查询
//https://www.alibabacloud.com/help/zh/doc-detail/118472.htm?spm=a2c63.p38356.b99.35.457e524dKiDhkZ
func (ca *costAccount) getTransactions(ctx context.Context, start, end string, info *historyInfo) error {

	defer func() {
		if e := recover(); e != nil {
			moduleLogger.Errorf("panic, %v", e)
		}
	}()

	logPrefix := ""
	if info != nil {
		logPrefix = "(history) "
	}

	//账户余额查询接口
	//https://www.alibabacloud.com/help/zh/doc-detail/87997.htm?spm=a2c63.p38356.b99.34.717125earDjnGX
	moduleLogger.Infof("%sgetting Balance", logPrefix)

	balanceReq := bssopenapi.CreateQueryAccountBalanceRequest()
	balanceReq.Scheme = "https"
	balanceResp, err := ca.ag.queryAccountBalanceWrap(ctx, balanceReq)
	if err != nil {
		moduleLogger.Errorf("%sfail to get balance, %s", logPrefix, err)
		return err
	}

	select {
	case <-ctx.Done():
		return nil
	default:
	}

	moduleLogger.Infof("%sgetting Transactions(%s - %s)", logPrefix, start, end)

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
			for atomic.LoadInt32(&ca.historyFlag) == 1 {
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

		resp, err := ca.ag.queryAccountTransactionsWrap(ctx, req)
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if err != nil {
			moduleLogger.Errorf("%s", err)
			return fmt.Errorf("%sfail to get transactions from %s to %s", logPrefix, start, end)
		}

		moduleLogger.Debugf(" %sPage%d: count=%d, TotalCount=%d, PageSize=%d", logPrefix, resp.Data.PageNum, len(resp.Data.AccountTransactionsList.AccountTransactionsListItem), resp.Data.TotalCount, resp.Data.PageSize)

		if err := ca.parseTransactionsResponse(ctx, balanceResp, resp); err != nil {
			return err
		}

		if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
			//有后续页
			req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
			if info != nil {
				info.PageNum = resp.Data.PageNum + 1
				if !ca.ag.debugMode {
					setAliyunCostHistory(info.key, info)
				}
			}
		} else {

			break
		}

	}

	moduleLogger.Debugf("%sfinish Transactions(%s - %s)", logPrefix, start, end)

	if info != nil {
		info.Statue = 1
		if !ca.ag.debugMode {
			setAliyunCostHistory(info.key, info)
		}
	}

	return nil
}

func (ca *costAccount) parseTransactionsResponse(ctx context.Context, balanceResp *bssopenapi.QueryAccountBalanceResponse, resp *bssopenapi.QueryAccountTransactionsResponse) error {

	balanceTags := map[string]string{}
	balanceFields := map[string]interface{}{}

	balanceTags[`Currency`] = balanceResp.Data.Currency

	var fv float64
	var err error
	if fv, err = strconv.ParseFloat(datakit.NumberFormat(balanceResp.Data.AvailableAmount), 64); err == nil {
		balanceFields[`AvailableAmount`] = fv
	}
	if fv, err = strconv.ParseFloat(datakit.NumberFormat(balanceResp.Data.MybankCreditAmount), 64); err == nil {
		balanceFields[`MybankCreditAmount`] = fv
	}
	if fv, err = strconv.ParseFloat(datakit.NumberFormat(balanceResp.Data.AvailableCashAmount), 64); err == nil {
		balanceFields[`AvailableCashAmount`] = fv
	}
	if fv, err = strconv.ParseFloat(datakit.NumberFormat(balanceResp.Data.CreditAmount), 64); err == nil {
		balanceFields[`CreditAmount`] = fv
	}

	accname := resp.Data.AccountName
	accid := resp.Data.AccountID
	if accid == "" {
		accid = ca.ag.accountID
	}
	if accname == "" {
		accname = ca.ag.accountName
	}

	for _, item := range resp.Data.AccountTransactionsList.AccountTransactionsListItem {

		select {
		case <-ctx.Done():
			return nil
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
		fields[`RecordID`] = item.RecordID                                                //订单号/账单号
		fields[`Remarks`] = item.Remarks                                                  //交易备注
		fields[`Amount`], _ = strconv.ParseFloat(datakit.NumberFormat(item.Amount), 64)   //交易金额
		fields[`Balance`], _ = strconv.ParseFloat(datakit.NumberFormat(item.Balance), 64) //交易发生前当前余额

		tm, err := time.Parse("2006-01-02T15:04:05Z", item.TransactionTime)
		if err != nil {
			moduleLogger.Warnf("fail to parse time:%v %s, error: %s", item.TransactionTime, item.RecordID, err)
		} else {
			tm = tm.Add(-8 * time.Hour) //返回的不是unix时间字符串

			if ca.ag.debugMode {
				//data, _ := io.MakeMetric(ca.getName(), tags, fields, tm)
				//fmt.Printf("-----%s\n", string(data))
			} else {
				io.NamedFeedEx(inputName, io.Metric, ca.getName(), tags, fields, tm)
			}
		}
	}

	return nil
}

func (ca *costAccount) getInterval() time.Duration {
	return ca.interval
}

func (ca *costAccount) getName() string {
	return ca.measurement
}
