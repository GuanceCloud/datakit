package aliyuncost

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
)

func staticClient() *bssopenapi.Client {
	//client, err := bssopenapi.NewClientWithAccessKey(`cn-hangzhou`, `LTAIaB2ZMYy4Dej9`, `pixGuiJail10JSBZTzuaOJIw8N2pw7`)
	//client, err := bssopenapi.NewClientWithAccessKey(`cn-hangzhou`, `LTAI4G7oxhYKY5n845WuieVg`, `GhLfPOACip4hB5mDp8C0fzO4GXNvXw`)
	client, err := bssopenapi.NewClientWithAccessKey(`cn-hangzhou`, `LTAI4G4wZv87CS4EkjMEY6N8`, `qI1TO1H7wEnchMlU2aUf4JwtITMHE4`)

	if err != nil {
		log.Fatalln(err)
	}
	return client

}

//https://help.aliyun.com/document_detail/87997.html?spm=a2c4g.11186623.6.621.a5f8392dHi0imZ
func TestAccountBalance(t *testing.T) {

	cli := staticClient()

	req := bssopenapi.CreateQueryAccountBalanceRequest()

	// req := bssopenapi.CreateQueryProductListRequest()
	// req.PageNum = requests.NewInteger(0)
	// req.QueryTotalCount = requests.NewBoolean(true)

	resp, err := cli.QueryAccountBalance(req)
	if err != nil {
		log.Fatalln(err)
	}

	fields := map[string]interface{}{}
	tags := map[string]string{}

	tags[`Currency`] = resp.Data.Currency

	var fv float64
	if fv, err = strconv.ParseFloat(datakit.NumberFormat(resp.Data.AvailableAmount), 64); err == nil {
		fields[`AvailableAmount`] = fv
	}
	if fv, err = strconv.ParseFloat(datakit.NumberFormat(resp.Data.MybankCreditAmount), 64); err == nil {
		fields[`MybankCreditAmount`] = fv
	}
	if fv, err = strconv.ParseFloat(datakit.NumberFormat(resp.Data.AvailableCashAmount), 64); err == nil {
		fields[`AvailableCashAmount`] = fv
	}
	if fv, err = strconv.ParseFloat(datakit.NumberFormat(resp.Data.CreditAmount), 64); err == nil {
		fields[`CreditAmount`] = fv
	}

	log.Printf("%s", resp.String())
	log.Printf("tags: %v", tags)
	log.Printf("fields: %v", fields)
}

func TestAccountTransactions(t *testing.T) {

	cli := staticClient()

	req := bssopenapi.CreateQueryAccountTransactionsRequest()
	req.PageSize = requests.NewInteger(300)

	now := time.Now().Truncate(time.Minute)
	start := now.Add(-time.Hour * 24)
	from := unixTimeStr(start) //需要传unix时间
	end := unixTimeStr(now)

	log.Printf("from=%s, end=%s", from, end)

	//start := "2020-04-10T00:00:00Z" // now.Add(-time.Hour * 24).Format(`2006-01-02T15:04:05Z`)
	req.CreateTimeStart = from
	//req.CreateTimeStart = ""
	req.CreateTimeEnd = end // now.Format(`2006-01-02T15:04:05Z`)
	req.CreateTimeEnd = ""

	resp, err := cli.QueryAccountTransactions(req)
	if err != nil {
		log.Fatalf("err: %s", err)
	}

	log.Printf("total: %v, accountid=%s, accountname=%s", len(resp.Data.AccountTransactionsList.AccountTransactionsListItem), resp.Data.AccountID, resp.Data.AccountName)

	//og.Printf("%s", resp.String())

	//fmt.Printf("TotalCount=%d, PageSize=%d, PageNum=%d\n", resp.Data.TotalCount, resp.Data.PageSize, resp.Data.PageNum)

	for _, at := range resp.Data.AccountTransactionsList.AccountTransactionsListItem {

		//tm, _ := time.Parse("2006-01-02T15:04:05Z", at.TransactionTime)
		//tm = tm.Add(-8 * time.Hour)
		log.Printf("%s, %s", at.RecordID, at.TransactionTime)

		//log.Printf("%s - %s - %s - %s, %s", at.TransactionTime, at.RecordID, at.TransactionChannelSN, at.Amount, at.Balance)
	}
}

func TestQueryBillOverview(t *testing.T) {

	cli := staticClient()

	req := bssopenapi.CreateQueryBillOverviewRequest()
	req.BillingCycle = fmt.Sprintf("%d-%d", 2020, 2)

	resp, err := cli.QueryBillOverview(req)
	if err != nil {
		log.Fatalf("%s", err)
	}

	log.Printf("AccountID=%s, AccountName=%s", resp.Data.AccountID, resp.Data.AccountName)
}

func TestQueryBill(t *testing.T) {

	cli := staticClient()

	//计费项 + 明细
	//包含了计费项的结账时间， 每个计费项的结账时间单位可能不同
	req := bssopenapi.CreateQueryBillRequest()
	req.BillingCycle = fmt.Sprintf("%d-%d", 2020, 3)
	req.PageSize = requests.NewInteger(300)
	req.PageNum = requests.NewInteger(1)
	req.IsHideZeroCharge = requests.NewBoolean(true) //过滤掉原价为0

	for {
		resp, err := cli.QueryBill(req)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("total count: %v, accountid=%s, accountname=%s", resp.Data.TotalCount, resp.Data.AccountID, resp.Data.AccountName)

		/* unreachable code
		for _, item := range resp.Data.Items.Item {
			if item.RecordID == "2020020989169351" {

				fmt.Printf("%s; %v; %s\n", item.ProductName, item.PretaxAmount, item.UsageStartTime)

				t, _ := time.Parse(`2006-01-02 15:04:05`, item.UsageStartTime)
				t = t.Add(-8 * time.Hour)
				fmt.Printf("convert_time: %s\n", t)
			}
		}

		if resp.Data.TotalCount > 0 && resp.Data.PageNum*resp.Data.PageSize < resp.Data.TotalCount {

			req.PageNum = requests.NewInteger(resp.Data.PageNum + 1)
		} else {
			break
		} */
	}

}

func TestQueryInstBill(t *testing.T) {

	cli := staticClient()

	//计费项 + 账期
	req := bssopenapi.CreateQueryInstanceBillRequest()
	//today := time.Now()
	req.PageSize = requests.NewInteger(300)
	req.BillingCycle = "2020-04" // fmt.Sprintf("%d-%d", today.Year(), today.Month()) // `2019-10-01`
	//req.BillingDate = "2020-05-32"
	//req.Granularity = "DAILY"
	req.IsBillingItem = requests.NewBoolean(true)

	resp, err := cli.QueryInstanceBill(req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("count=%d", len(resp.Data.Items.Item))

	_ = resp
	for _, item := range resp.Data.Items.Item {
		//if item.PaymentTime != "" {
		fmt.Printf("%s - %s(%s), %v, %s\n", item.BillingDate, item.ProductName, item.InstanceID, item.PretaxAmount, item.BillingItem)
		//}
	}

}

func TestQueryOrderDetail(t *testing.T) {
	cli := staticClient()

	req := bssopenapi.CreateGetOrderDetailRequest()
	req.OrderId = "203219046480001"

	resp, err := cli.GetOrderDetail(req)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("%s", resp.Data.AccountName)
}

func TestQueryOrders(t *testing.T) {

	cli := staticClient()

	req := bssopenapi.CreateQueryOrdersRequest()
	// now := time.Now().Truncate(time.Hour)
	// start := unixTimeStr(now.Add(-time.Hour * 24 * 30))
	// log.Printf("start=%s", start)
	req.CreateTimeStart = "2020-02-18T06:26:00Z" // start
	req.CreateTimeEnd = "2020-03-18T06:26:00Z"
	req.PageNum = requests.NewInteger(1)
	req.PageSize = requests.NewInteger(300)

	resp, err := cli.QueryOrders(req)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("TotalCount=%d, PageNum=%d, PageSize=%d, count=%d\n", resp.Data.TotalCount, resp.Data.PageNum, resp.Data.PageSize, len(resp.Data.OrderList.Order))

	for _, item := range resp.Data.OrderList.Order {
		fmt.Printf("%s - %s, %v, %s\n", item.CreateTime, item.PaymentStatus, item.PretaxAmount, item.Currency)
	}

}

func TestConfig(t *testing.T) {
	// if err := Cfg.Load(`./demo.toml`); err != nil {
	// 	fmt.Printf("%s", err)
	// }

	// fmt.Printf("%#v", Cfg.Boas[0].BiilInterval)
}

func TestSvr(t *testing.T) {

	ag := newAgent()

	if data, err := ioutil.ReadFile("./test.conf"); err != nil {
		log.Fatalf("%s", err)
	} else {
		if toml.Unmarshal(data, ag); err != nil {
			log.Fatalf("%s", err)
		}
	}

	ag.debugMode = true
	ag.Run()
}
