package rabbitmq

import (
	"flag"
	"fmt"
	"testing"
)

var (
	flagUser     = flag.String("user", "", ``)
	flagPassword = flag.String("password", "", ``)
)

func TestGetOverview(t *testing.T) {
	n := initInput()
	getOverview(n)
	for _, v := range collectCache {
		fmt.Println(v.LineProto())

	}
}

func TestGetNode(t *testing.T) {
	n := initInput()
	getNode(n)
	for _, v := range collectCache {
		fmt.Println(v.LineProto())

	}
}

func TestGetExchange(t *testing.T) {
	n := initInput()
	getExchange(n)
	for _, v := range collectCache {
		fmt.Println(v.LineProto())

	}
}

func TestGetQueue(t *testing.T) {
	n := initInput()
	getQueues(n)
	for _, v := range collectCache {
		fmt.Println(v.LineProto())

	}
}

func initInput() *Input {
	flag.Parse()
	n := &Input{
		Url:      "http://10.100.65.53:15672",
		Username: *flagUser,
		Password: *flagPassword,
	}
	cli, err := n.createHttpClient()
	if err != nil {
		l.Fatal(err)
	}
	n.client = cli
	return n
}
