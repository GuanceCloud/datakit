package aliyunboa

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/log"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/uploader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/utils"
)

func init() {
	config.AddConfig("aliyunboa", &Cfg)
	service.Add("aliyunboa", func(logger log.Logger) service.Service {
		if len(Cfg.Boas) == 0 {
			return nil
		}

		return &AliyunBoaSvr{
			logger: logger,
		}
	})
}

var (
	batchInterval = time.Duration(5) * time.Minute
	metricPeriod  = time.Duration(5 * time.Minute)
	rateLimit     = 20
)

type (
	RunningBoa struct {
		cfg *Boa

		wg sync.WaitGroup

		uploader uploader.IUploader
		logger   log.Logger

		client *bssopenapi.Client

		modules []BoaModule
	}

	AliyunBoaSvr struct {
		boas   []*RunningBoa
		logger log.Logger
	}

	BoaModule interface {
		fetch(context.Context, *bssopenapi.Client, log.Logger, uploader.IUploader) error
		getInterval() time.Duration
		getName() string
	}

	BoaAccount struct {
		interval time.Duration
		name     string
	}

	BoaBill struct {
		interval time.Duration
		name     string
	}

	BoaOrder struct {
		interval time.Duration
		name     string
	}
)

func (s *AliyunBoaSvr) Start(ctx context.Context, up uploader.IUploader) error {

	if len(Cfg.Boas) == 0 {
		return nil
	}

	s.boas = []*RunningBoa{}

	for _, c := range Cfg.Boas {
		a := &RunningBoa{
			cfg:      c,
			uploader: up,
			logger:   s.logger,
			modules: []BoaModule{
				&BoaAccount{
					name:     "aliyun_cost_account",
					interval: c.AccountInterval.Duration,
				},
				&BoaBill{
					name:     "aliyun_cost_bill",
					interval: c.AccountInterval.Duration,
				},
				&BoaOrder{
					name:     "aliyun_cost_order",
					interval: c.AccountInterval.Duration,
				},
			},
		}

		s.boas = append(s.boas, a)
	}

	var wg sync.WaitGroup

	s.logger.Info("Starting AliyunBoaSvr...")

	for _, c := range s.boas {
		wg.Add(1)
		go func(ac *RunningBoa) {
			defer wg.Done()

			if err := ac.Run(ctx); err != nil && err != context.Canceled {
				s.logger.Errorf("%s", err)
			}
		}(c)
	}

	wg.Wait()

	s.logger.Info("AliyunBoaSvr done")
	return nil
}

func (b *BoaOrder) getInterval() time.Duration {
	return b.interval
}

func (b *BoaOrder) getName() string {
	return b.name
}

func (b *BoaOrder) fetch(ctx context.Context, client *bssopenapi.Client, l log.Logger, up uploader.IUploader) error {

	reqOrder := bssopenapi.CreateQueryOrdersRequest()
	respOrder, err := client.QueryOrders(reqOrder)
	if err != nil {
		return err
	}

	for _, item := range respOrder.Data.OrderList.Order {

		tags := map[string]string{
			"ProductCode":      item.ProductCode,
			"ProductType":      item.ProductType,
			"SubscriptionType": item.SubscriptionType,
			"OrderType":        item.OrderType,
			"Currency":         item.Currency,
		}

		fields := map[string]interface{}{
			"OrderID":           item.OrderId,
			"RelatedOrderId":    item.RelatedOrderId,
			"PretaxGrossAmount": item.PretaxGrossAmount,
			"PretaxAmount":      item.PretaxAmount,
		}

		t, err := time.Parse(time.RFC3339, item.PaymentTime)
		if err != nil {
			l.Errorf("fail to parse time %s, error:%s", item.PaymentTime, err)
		} else {
			addLine(b.getName(), tags, fields, t, up, l)
		}

	}

	return nil
}

func (b *BoaBill) fetch(ctx context.Context, client *bssopenapi.Client, l log.Logger, up uploader.IUploader) error {

	reqBill := bssopenapi.CreateQueryBillRequest()
	today := time.Now()
	reqBill.BillingCycle = fmt.Sprintf("%d-%d", today.Year(), today.Month())
	respBill, err := client.QueryBill(reqBill)
	if err != nil {
		return err
	}

	for _, item := range respBill.Data.Items.Item {
		tags := map[string]string{
			"AccountID":   respBill.Data.AccountID,
			"AccountName": respBill.Data.AccountName,
			"OwnerID":     fmt.Sprintf("%v", respBill.Data.OwnerUid),
		}

		tags["OwnerID"] = item.OwnerID
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
		fields[`RoundDownDiscount`] = item.RoundDownDiscount
		fields[`PretaxAmount`] = item.PretaxAmount
		fields[`DeductedByCashCoupons`] = item.DeductedByCashCoupons
		fields[`DeductedByPrepaidCard`] = item.DeductedByPrepaidCard
		fields[`PaymentAmount`] = item.PaymentAmount
		fields[`OutstandingAmount`] = item.OutstandingAmount

		t, err := time.Parse(`2006-01-02 15:04:05`, item.UsageEndTime)
		if err != nil {
			l.Errorf("fail to parse time %s, error:%s", item.UsageEndTime, err)
		} else {
			addLine(b.getName(), tags, fields, t, up, l)
		}
	}

	reqInstBill := bssopenapi.CreateQueryInstanceBillRequest()
	reqInstBill.BillingCycle = fmt.Sprintf("%d-%d", today.Year(), today.Month())
	respInstill, err := client.QueryInstanceBill(reqInstBill)
	if err != nil {
		return err
	}

	//QueryInstanceBill
	for _, item := range respInstill.Data.Items.Item {
		tags := map[string]string{
			"AccountID":   respInstill.Data.AccountID,
			"AccountName": respInstill.Data.AccountName,
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

		addLine(b.getName(), tags, fields, time.Now().UTC(), up, l)
	}

	return nil
}

func (b *BoaBill) getInterval() time.Duration {
	return b.interval
}

func (b *BoaBill) getName() string {
	return b.name
}

func (b *BoaAccount) fetch(ctx context.Context, client *bssopenapi.Client, l log.Logger, up uploader.IUploader) error {

	reqAmount := bssopenapi.CreateQueryAccountBalanceRequest()
	respAmount, err := client.QueryAccountBalance(reqAmount)
	if err != nil {
		return err
	}
	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags[`Currency`] = respAmount.Data.Currency
	fields[`AvailableAmount`] = respAmount.Data.AvailableAmount
	fields[`MybankCreditAmount`] = respAmount.Data.MybankCreditAmount
	fields[`AvailableCashAmount`] = respAmount.Data.AvailableCashAmount
	fields[`CreditAmount`] = respAmount.Data.CreditAmount

	addLine(b.getName(), tags, fields, time.Now().UTC(), up, l)

	reqTransaction := bssopenapi.CreateQueryAccountTransactionsRequest()
	respTransaction, err := client.QueryAccountTransactions(reqTransaction)
	if err != nil {
		return err
	}

	if respTransaction.Data.AccountTransactionsList.AccountTransactionsListItem != nil {
		for _, item := range respTransaction.Data.AccountTransactionsList.AccountTransactionsListItem {
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
			fields[`Amount`] = item.Amount
			fields[`Balance`] = item.Balance

			t, err := time.Parse(time.RFC3339, item.TransactionTime)
			if err != nil {
				l.Errorf("fail to parse time %s, error:%s", item.TransactionTime, err)
			} else {
				addLine(b.getName(), tags, fields, t, up, l)
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

func (s *RunningBoa) Run(ctx context.Context) error {

	var err error
	s.client, err = bssopenapi.NewClientWithAccessKey(s.cfg.RegionID, s.cfg.AccessKeyID, s.cfg.AccessKeySecret)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	lmtr := utils.NewRateLimiter(rateLimit, time.Second)
	defer lmtr.Stop()

	for _, boaModule := range s.modules {
		s.wg.Add(1)

		go func(m BoaModule, ctx context.Context) {
			defer s.wg.Done()

			ticker := time.NewTicker(m.getInterval())

			for {

				s.logger.Debugf("start fetch")

				err = m.fetch(ctx, s.client, s.logger, s.uploader)
				if err != nil {
					s.logger.Errorf(`fail to get metric %s: %s`, m.getName(), err)
				}

				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
			}

		}(boaModule, ctx)

	}

	s.wg.Wait()

	return nil
}

func addLine(metricName string, tags map[string]string, fields map[string]interface{}, tm time.Time, up uploader.IUploader, l log.Logger) error {

	serializer := influx.NewSerializer()

	m, _ := metric.New(metricName, tags, fields, tm)

	output, err := serializer.Serialize(m)
	l.Debug(string(output))
	if err == nil {
		if up != nil {
			up.AddLog(&uploader.LogItem{
				Log: string(output),
			})
		}
	} else {
		l.Warnf("[warn] Serialize to influx protocol line fail(%s): %s; tags:%#v, fields:%#v", metricName, err, tags, fields)
	}

	return err
}
