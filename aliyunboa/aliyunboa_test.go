package aliyunboa

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/siddontang/go-log/log"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
)

func staticClient() *bssopenapi.Client {
	client, err := bssopenapi.NewClientWithAccessKey(`cn-hangzhou`, `LTAI4Fc72xGdZKKr6cTBV72S`, `QXZ4FFCq3yhN5TCGC9rj1kBNZNJksc`)
	if err != nil {
		log.Fatalln(err)
	}
	return client
}

//https://help.aliyun.com/document_detail/87997.html?spm=a2c4g.11186623.6.621.a5f8392dHi0imZ
func TestAccountBalance(t *testing.T) {

	cli := staticClient()

	bacReq := bssopenapi.CreateQueryAccountBalanceRequest()

	// req := bssopenapi.CreateQueryProductListRequest()
	// req.PageNum = requests.NewInteger(0)
	// req.QueryTotalCount = requests.NewBoolean(true)

	resp, err := cli.QueryAccountBalance(bacReq)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println(resp.String())
}

func TestAccountTransactions(t *testing.T) {

	cli := staticClient()

	req := bssopenapi.CreateQueryAccountTransactionsRequest()

	resp, err := cli.QueryAccountTransactions(req)
	if err != nil {
		log.Fatalln(err)
	}

	for _, at := range resp.Data.AccountTransactionsList.AccountTransactionsListItem {
		log.Printf("%s - %s - %s, %s", at.TransactionTime, at.TransactionAccount, at.Amount, at.Balance)
	}
}

func TestQueryBill(t *testing.T) {

	cli := staticClient()

	req := bssopenapi.CreateQueryBillRequest()
	today := time.Now()
	req.BillingCycle = fmt.Sprintf("%d-%d", today.Year(), today.Month()) // `2019-10-01`

	resp, err := cli.QueryBill(req)
	if err != nil {
		log.Fatalln(err)
	}

	for _, item := range resp.Data.Items.Item {
		fmt.Printf("%s - %s, %v\n", item.UsageStartTime, item.ProductName, item.PretaxAmount)
	}

}

func TestQueryInstBill(t *testing.T) {

	cli := staticClient()

	req := bssopenapi.CreateQueryInstanceBillRequest()
	//today := time.Now()
	req.BillingCycle = "2019-10" // fmt.Sprintf("%d-%d", today.Year(), today.Month()) // `2019-10-01`

	resp, err := cli.QueryInstanceBill(req)
	if err != nil {
		log.Fatalln(err)
	}

	for _, item := range resp.Data.Items.Item {
		fmt.Printf("%s - %s, %v, %s\n", item.UsageStartTime, item.ProductName, item.PretaxAmount, item.Tag)
	}

}

func TestQueryOrder(t *testing.T) {

	cli := staticClient()

	req := bssopenapi.CreateQueryOrdersRequest()

	resp, err := cli.QueryOrders(req)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(resp.String())

	for _, item := range resp.Data.OrderList.Order {
		fmt.Printf("%s - %s, %v, %s\n", item.PaymentTime, item.PaymentStatus, item.PretaxAmount, item.Currency)
	}

}

func TestSvr(t *testing.T) {

	if err := Cfg.Load(`./demo.toml`); err != nil {
		log.Fatalln(err)
	}

	logHandler, _ := log.NewStreamHandler(os.Stdout)

	ll := log.NewDefault(logHandler)
	ll.SetLevel(log.LevelDebug)

	ll.Debugf("acckey: %s, accountInterval: %v", Cfg.Boas[0].AccessKeySecret, Cfg.Boas[0].AccountInterval)

	svr := &AliyunBoaSvr{
		logger: ll,
	}

	ctx, _ := context.WithCancel(context.Background())

	svr.Start(ctx, nil)

}
