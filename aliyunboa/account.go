package aliyunboa

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/utils"
)

type BoaAccount struct {
	interval time.Duration
	name     string
	boa      *RunningBoa
}

func (b *BoaAccount) run(ctx context.Context) error {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.getLastyearData(ctx, b.boa.lmtr)
	}()

	b.getRealtimeData(ctx)

	wg.Wait()

	b.boa.logger.Info("aliyun account fetcher done")

	return nil
}

func (b *BoaAccount) getRealtimeData(ctx context.Context) error {

	b.boa.suspendLastyearFetch()
	now := time.Now().Truncate(time.Minute)
	start := now.Add(-b.boa.cfg.AccountInterval.Duration).Format(`2006-01-02T15:04:05Z`)
	end := now.Format(`2006-01-02T15:04:05Z`)
	if err := b.getTransactions(ctx, start, end, b.boa.lmtr, false); err != nil && err != context.Canceled {
		b.boa.logger.Errorf("%s", err)
	}
	b.boa.resumeLastyearFetch()

	ticker := time.NewTicker(b.boa.cfg.AccountInterval.Duration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-ticker.C:

			b.boa.suspendLastyearFetch()

			if err := b.getBalance(ctx); err != nil {
				b.boa.logger.Errorf("%s", err)
			}

			now := time.Now()
			start := now.Add(-b.boa.cfg.AccountInterval.Duration).Format(`2006-01-02T15:04:05Z`)
			end := now.Format(`2006-01-02T15:04:05Z`)
			if err := b.getTransactions(ctx, start, end, b.boa.lmtr, false); err != nil && err != context.Canceled {
				b.boa.logger.Errorf("%s", err)
			}

			b.boa.resumeLastyearFetch()
		}
	}
}

func (b *BoaAccount) getLastyearData(ctx context.Context, lmtr *utils.RateLimiter) error {

	if !b.boa.cfg.CollectHistoryData {
		return nil
	}

	b.boa.logger.Info("start get account transactions of last year")

	m := md5.New()
	m.Write([]byte(b.boa.cfg.AccessKeyID))
	m.Write([]byte(b.boa.cfg.AccessKeySecret))
	m.Write([]byte(b.boa.cfg.RegionID))
	m.Write([]byte(`account`))
	k1 := hex.EncodeToString(m.Sum(nil))
	k1 = "." + k1

	orderFlag, _ := config.GetLastyearFlag(k1)

	if orderFlag == 1 {
		return nil
	}

	now := time.Now().Truncate(time.Minute).Add(-b.boa.cfg.AccountInterval.Duration)
	start := now.Add(-time.Hour * 8760).Format(`2006-01-02T15:04:05Z`)
	end := now.Format(`2006-01-02T15:04:05Z`)
	if err := b.getTransactions(ctx, start, end, lmtr, true); err != nil && err != context.Canceled {
		b.boa.logger.Errorf("fail to get account transactions of last year(%s-%s): %s", start, end, err)
		return err
	}

	config.SetLastyearFlag(k1, 1)

	return nil
}

func (b *BoaAccount) getBalance(ctx context.Context) error {

	req := bssopenapi.CreateQueryAccountBalanceRequest()
	resp, err := b.boa.client.QueryAccountBalance(req)
	if err != nil {
		return err
	}

	return b.parseBalanceResponse(ctx, resp)
}

func (b *BoaAccount) getTransactions(ctx context.Context, start, end string, lmtr *utils.RateLimiter, fromLastyear bool) error {

	defer func() {
		recover()
	}()

	b.boa.logger.Infof("start get account transactions from %s to %s", start, end)

	var respTransactions *bssopenapi.QueryAccountTransactionsResponse

	req := bssopenapi.CreateQueryAccountTransactionsRequest()
	req.CreateTimeStart = start
	req.CreateTimeEnd = end
	req.PageSize = requests.NewInteger(300)

	for {
		if fromLastyear {
			b.boa.wait()
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		case <-lmtr.C:
		}

		resp, err := b.boa.client.QueryAccountTransactions(req)
		if err != nil {
			return fmt.Errorf("fail to get transactions from %s to %s: %s", start, end, err)
		} else {
			b.boa.logger.Debugf("[Transactions] %s - %s: TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", start, end, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))

			if respTransactions == nil {
				respTransactions = resp
			} else {
				respTransactions.Data.AccountTransactionsList.AccountTransactionsListItem = append(respTransactions.Data.AccountTransactionsList.AccountTransactionsListItem, resp.Data.AccountTransactionsList.AccountTransactionsListItem...)
			}

			if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
				req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
			} else {
				break
			}
		}
	}

	b.boa.logger.Debugf("get transactions between %s-%s finish count=%d", start, end, len(respTransactions.Data.AccountTransactionsList.AccountTransactionsListItem))

	return b.parseTransactionsResponse(ctx, respTransactions)
}

func (b *BoaAccount) parseBalanceResponse(ctx context.Context, resp *bssopenapi.QueryAccountBalanceResponse) error {

	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags[`Currency`] = resp.Data.Currency

	fields[`AvailableAmount`], _ = strconv.ParseFloat(resp.Data.AvailableAmount, 64)
	fields[`MybankCreditAmount`], _ = strconv.ParseFloat(resp.Data.MybankCreditAmount, 64)
	fields[`AvailableCashAmount`], _ = strconv.ParseFloat(resp.Data.AvailableCashAmount, 64)
	fields[`CreditAmount`], _ = strconv.ParseFloat(resp.Data.CreditAmount, 64)

	addLine(b.getName(), tags, fields, time.Now().UTC(), b.boa.uploader, b.boa.logger)

	return nil
}

func (b *BoaAccount) parseTransactionsResponse(ctx context.Context, resp *bssopenapi.QueryAccountTransactionsResponse) error {
	if resp.Data.AccountTransactionsList.AccountTransactionsListItem != nil {
		for _, item := range resp.Data.AccountTransactionsList.AccountTransactionsListItem {

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			tags := map[string]string{}
			fields := map[string]interface{}{}

			tags[`TransactionFlow`] = item.TransactionFlow
			tags[`TransactionType`] = item.TransactionType
			tags[`TransactionChannel`] = item.TransactionChannel
			tags[`FundType`] = item.FundType
			//tags[`TransactionAccount`] = item.TransactionAccount

			fields[`TransactionNumber`] = item.TransactionNumber
			fields[`TransactionChannelSN`] = item.TransactionChannelSN
			fields[`RecordID`] = item.RecordID
			fields[`Remarks`] = item.Remarks
			fields[`Amount`], _ = strconv.ParseFloat(item.Amount, 64)
			fields[`Balance`], _ = strconv.ParseFloat(item.Balance, 64)

			t, err := time.Parse(time.RFC3339, item.TransactionTime)
			if err != nil {
				b.boa.logger.Errorf("fail to parse time %s, error:%s", item.TransactionTime, err)
			} else {
				addLine(b.getName(), tags, fields, t, b.boa.uploader, b.boa.logger)
			}
		}
	}

	return nil
}

func (b *BoaAccount) getInterval() time.Duration {
	return b.interval
}

func (b *BoaAccount) getName() string {
	return b.name
}
