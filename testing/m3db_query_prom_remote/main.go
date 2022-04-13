package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
)

func makePromRemoteReadClient() remote.ReadClient {
	m3url := "http://127.0.0.1:7201/api/v1/prom/remote/read"

	urlp, err := url.Parse(m3url)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	c := &config.URL{URL: urlp}
	client, err := remote.NewReadClient("prom", &remote.ClientConfig{
		URL:              c,
		Timeout:          model.Duration(time.Second * 10),
		HTTPClientConfig: config.HTTPClientConfig{},
		SigV4Config:      nil,
		Headers:          nil,
		RetryOnRateLimit: false,
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return client
}

func main() {
	client := makePromRemoteReadClient()
	if client == nil {
		fmt.Println("client is nil")
		return
	}
	timenow := time.Now().UnixNano() / 1e6
	q := &prompb.Query{
		StartTimestampMs: timenow - (1000 * 60 * 30), // 1h
		EndTimestampMs:   timenow,
		Matchers:         nil,
	}
	resource, err := client.Read(context.Background(), q)
	if err != nil {
		fmt.Println("read err")
		fmt.Println(err)
		return
	}
	for _, ts := range resource.Timeseries {
		fmt.Printf("labels = %+v \n", ts.Labels)
		fmt.Printf("len(ts.Samples)=%d \n", len(ts.Samples))
		fmt.Println("----------------------------------------------------------------")
	}
}
