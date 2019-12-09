package aliyunboa

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/utils"
)

type BoaBill struct {
	interval time.Duration
	name     string
	boa      *RunningBoa
}

func (b *BoaBill) run(ctx context.Context) error {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.getLastyearData(ctx, b.boa.lmtr)
	}()

	b.getRealtimeData(ctx)

	wg.Wait()

	b.boa.logger.Info("aliyun bill fetcher done")

	return nil
}

func (b *BoaBill) getRealtimeData(ctx context.Context) error {

	ticker := time.NewTicker(b.boa.cfg.BiilInterval.Duration)
	defer ticker.Stop()

	b.boa.suspendLastyearFetch()
	now := time.Now()
	cycle := fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	if err := b.getBills(ctx, cycle, b.boa.lmtr, false); err != nil && err != context.Canceled {
		b.boa.logger.Errorf("%s", err)
	}
	b.boa.resumeLastyearFetch()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-ticker.C:

			b.boa.suspendLastyearFetch()
			now := time.Now()
			cycle := fmt.Sprintf("%d-%02d", now.Year(), now.Month())
			if err := b.getBills(ctx, cycle, b.boa.lmtr, false); err != nil && err != context.Canceled {
				b.boa.logger.Errorf("%s", err)
			}
			if err := b.getInstnceBills(ctx, cycle, b.boa.lmtr); err != nil && err != context.Canceled {
				b.boa.logger.Errorf("%s", err)
			}
			b.boa.resumeLastyearFetch()
		}
	}
}

func (b *BoaBill) getLastyearData(ctx context.Context, lmtr *utils.RateLimiter) error {

	if !b.boa.cfg.CollectHistoryData {
		return nil
	}

	b.boa.logger.Info("start get bills of last year")

	m := md5.New()
	m.Write([]byte(b.boa.cfg.AccessKeyID))
	m.Write([]byte(b.boa.cfg.AccessKeySecret))
	m.Write([]byte(b.boa.cfg.RegionID))
	m.Write([]byte(`bills`))
	k1 := hex.EncodeToString(m.Sum(nil))
	k1 = "." + k1

	billFlag, _ := config.GetLastyearFlag(k1)

	m.Reset()
	m.Write([]byte(b.boa.cfg.AccessKeyID))
	m.Write([]byte(b.boa.cfg.AccessKeySecret))
	m.Write([]byte(b.boa.cfg.RegionID))
	m.Write([]byte(`bills_instance`))
	k2 := hex.EncodeToString(m.Sum(nil))
	k2 = "." + k2

	billInstanceFlag, _ := config.GetLastyearFlag(k2)

	if billInstanceFlag == 1 && billFlag == 1 {
		return nil
	}

	now := time.Now().Add(-24 * time.Hour * 30)
	for index := 0; index < 11; index++ {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}
		cycle := fmt.Sprintf("%d-%02d", now.Year(), now.Month())

		if billFlag == 0 {
			if err := b.getBills(ctx, cycle, lmtr, true); err != nil {
				b.boa.logger.Errorf("%s", err)
			}
		}

		if billInstanceFlag == 0 {
			if err := b.getInstnceBills(ctx, cycle, lmtr); err != nil {
				b.boa.logger.Errorf("%s", err)
			}
		}

		now = now.Add(-24 * time.Hour * 30)
	}

	config.SetLastyearFlag(k1, 1)
	config.SetLastyearFlag(k2, 1)

	return nil
}

func (b *BoaBill) getBills(ctx context.Context, cycle string, lmtr *utils.RateLimiter, fromLastyear bool) error {

	defer func() {
		recover()
	}()

	b.boa.logger.Infof("start get bills of %s", cycle)

	var respBill *bssopenapi.QueryBillResponse

	req := bssopenapi.CreateQueryBillRequest()
	req.BillingCycle = cycle
	req.PageSize = requests.NewInteger(300)

	for {

		if fromLastyear {
			b.boa.wait()
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		case <-lmtr.C:
			//
		}

		resp, err := b.boa.client.QueryBill(req)
		if err != nil {
			return fmt.Errorf("fail to get bill of %s: %s", cycle, err)
		}

		b.boa.logger.Debugf("[Bill] %s: TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", cycle, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))

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

	b.boa.logger.Debugf("finish getting Bill of %s finish count=%d", cycle, len(respBill.Data.Items.Item))

	return b.parseBillResponse(ctx, respBill)
}

func (b *BoaBill) getInstnceBills(ctx context.Context, cycle string, lmtr *utils.RateLimiter) error {

	var respInstill *bssopenapi.QueryInstanceBillResponse

	req := bssopenapi.CreateQueryInstanceBillRequest()
	req.BillingCycle = cycle
	req.PageSize = requests.NewInteger(300)

	for {

		select {
		case <-ctx.Done():
			return context.Canceled
		case <-lmtr.C:
		}

		resp, err := b.boa.client.QueryInstanceBill(req)
		if err != nil {
			return fmt.Errorf("fail to get instance bill of %s: %s", cycle, err)
		} else {
			b.boa.logger.Debugf("[InstanceBill] %s: TotalCount=%d, PageNum=%d, PageSize=%d, count=%d", cycle, resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.Items.Item))
			if respInstill == nil {
				respInstill = resp
			} else {
				if resp.Data.TotalCount > 0 {
					respInstill.Data.Items.Item = append(respInstill.Data.Items.Item, resp.Data.Items.Item...)
				}
			}

			if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {
				req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
			} else {
				break
			}
		}
	}

	return b.parseInstanceBillResponse(ctx, respInstill)
}

func (b *BoaBill) parseBillResponse(ctx context.Context, resp *bssopenapi.QueryBillResponse) error {

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
		if billtime == "" {
			billtime = item.PaymentTime
		}
		t, err := time.Parse(`2006-01-02 15:04:05`, billtime)
		if err != nil {
			b.boa.logger.Errorf("[Bill] fail to parse time [%s] of product [%s], error: %s", billtime, item.ProductName, err)
		} else {
			addLine(b.getName(), tags, fields, t, b.boa.uploader, b.boa.logger)
		}
	}

	return nil
}

func (b *BoaBill) parseInstanceBillResponse(ctx context.Context, resp *bssopenapi.QueryInstanceBillResponse) error {
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

		addLine(b.getName(), tags, fields, time.Now().UTC(), b.boa.uploader, b.boa.logger)
	}
	return nil
}

func (b *BoaBill) getInterval() time.Duration {
	return b.interval
}

func (b *BoaBill) getName() string {
	return b.name
}
