package docker

import (
	"fmt"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

/*
 * go test -v -c && sudo ./docker.test -test.v -test.run=TestMain
 */

func TestMain(t *testing.T) {
	var err error
	var d = newInput()

	d.client, err = d.newEnvClient()
	if err != nil {
		t.Fatal(err)
	}

	pts, err := d.gather(&gatherOption{IsObjectCategory: true})
	if err != nil {
		t.Fatal(err)
	}

	for _, pt := range pts {
		fmt.Println(pt.String())
	}
}

func TestGatherLog(t *testing.T) {
	io.SetTest()
	var err error
	var d = newInput()

	d.client, err = d.newEnvClient()
	if err != nil {
		t.Fatal(err)
	}

	if err = d.initLogOption(); err != nil {
		t.Fatal(err)
	}

	// 开始采集日志数据，非阻塞，会fork多个goroutine
	d.gatherLog()

	// 等待5秒用以采集
	time.Sleep(time.Second * 5)

	// 关闭所有采集资源
	d.cancelTails()
}
