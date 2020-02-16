package aliyuncost

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
		//ca.getLastyearData(ctx, ca.runningInstance.lmtr)
	}()

	ca.getRealtimeData(ctx)

	wg.Wait()

	ca.logger.Info("done")

	return nil
}

func (ca *CostAccount) getRealtimeData(ctx context.Context) error {

	for {
		//暂停历史数据抓取
		//ca.runningInstance.suspendLastyearFetch()
		now := time.Now().Truncate(time.Minute)
		start := now.Add(-ca.interval)

		from := strings.Replace(start.Format(time.RFC3339), "+", "Z", -1)
		end := strings.Replace(now.Format(time.RFC3339), "+", "Z", -1)
		if err := ca.getTransactions(ctx, from, end, ca.runningInstance.lmtr, false); err != nil && err != context.Canceled {
			ca.logger.Errorf("%s", err)
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		//ca.runningInstance.resumeLastyearFetch()
		internal.SleepContext(ctx, ca.interval)

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}
	}

	// ticker := time.NewTicker(ca.interval)
	// defer ticker.Stop()

	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		return context.Canceled
	// 	case <-ticker.C:

	// 		ca.runningInstance.suspendLastyearFetch()

	// 		if err := ca.getBalance(ctx); err != nil {
	// 			log.Printf("E! %s", err)
	// 		}

	// 		now := time.Now()
	// 		start := now.Add(-ca.interval).Format(`2006-01-02T15:04:05Z`)
	// 		end := now.Format(`2006-01-02T15:04:05Z`)
	// 		if err := ca.getTransactions(ctx, start, end, ca.runningInstance.lmtr, false); err != nil && err != context.Canceled {
	// 			log.Printf("E! %s", err)
	// 		}

	// 		ca.runningInstance.resumeLastyearFetch()
	// 	}
	// }
}

func (ca *CostAccount) getHistoryData(ctx context.Context, lmtr *limiter.RateLimiter) error {

	// if !ca.runningInstance.cfg.CollectHistoryData {
	// 	return nil
	// }

	// ca.logger.Info("start getting history Transactions")

	// m := md5.New()
	// m.Write([]byte(ca.runningInstance.cfg.AccessKeyID))
	// m.Write([]byte(ca.runningInstance.cfg.AccessKeySecret))
	// m.Write([]byte(ca.runningInstance.cfg.RegionID))
	// m.Write([]byte(`account`))
	// k := "." + hex.EncodeToString(m.Sum(nil))

	// orderFlag, _ := config.GetLastyearFlag(k)

	// if orderFlag == 1 {
	// 	return nil
	// }

	// now := time.Now().Truncate(time.Minute).Add(-ca.interval)
	// start := now.Add(-time.Hour * 8760).Format(`2006-01-02T15:04:05Z`)
	// end := now.Format(`2006-01-02T15:04:05Z`)
	// if err := ca.getTransactions(ctx, start, end, lmtr, true); err != nil && err != context.Canceled {
	// 	log.Printf("E! fail to get account transactions of last year(%s-%s): %s", start, end, err)
	// 	return err
	// }

	// config.SetLastyearFlag(k1, 1)

	return nil
}

//获取账户流水收支明细查询
//https://www.alibabacloud.com/help/zh/doc-detail/118472.htm?spm=a2c63.p38356.b99.35.457e524dKiDhkZ
func (ca *CostAccount) getTransactions(ctx context.Context, start, end string, lmtr *limiter.RateLimiter, fromLastyear bool) error {

	defer func() {
		recover()
	}()

	//账户余额查询接口
	//https://www.alibabacloud.com/help/zh/doc-detail/87997.htm?spm=a2c63.p38356.b99.34.717125earDjnGX
	ca.logger.Info("start getting Balance")
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

	ca.logger.Infof("start getting Transactions(%s - %s)", start, end)

	var respTransactions *bssopenapi.QueryAccountTransactionsResponse

	req := bssopenapi.CreateQueryAccountTransactionsRequest()
	req.CreateTimeStart = start
	req.CreateTimeEnd = end
	req.PageSize = requests.NewInteger(300)

	for { //分页
		if fromLastyear {
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
		ca.logger.Debugf("Transactions(%s - %s): TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", start, end, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.AccountTransactionsList.AccountTransactionsListItem))

		if respTransactions == nil {
			respTransactions = resp
		} else {
			//累积
			respTransactions.Data.AccountTransactionsList.AccountTransactionsListItem = append(respTransactions.Data.AccountTransactionsList.AccountTransactionsListItem, resp.Data.AccountTransactionsList.AccountTransactionsListItem...)
		}

		if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
			//有后续页
			req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
		} else {
			break
		}
	}

	ca.logger.Debugf("finish getting Transactions(%s - %s), count=%v", start, end, len(respTransactions.Data.AccountTransactionsList.AccountTransactionsListItem))

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
