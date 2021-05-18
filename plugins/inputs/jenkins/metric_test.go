package jenkins

import (
	"fmt"
	"testing"
)

func TestGetMetric(t *testing.T) {
	var n = &Input{
		Url: "http://10.100.65.12:8888",
		Key: "K1uc9f4u8Mh_LzCcE9khcKTst64JpDT6GrNb4RdRLhhM72UEMqmoWQRnDEFohA__",
	}

	client, err := n.createHttpClient()
	if err != nil {
		l.Fatal(err.Error())
	}
	n.client = client
	getPluginMetric(n)
	for _, v := range n.collectCache {
		fmt.Println(v)
	}

}
