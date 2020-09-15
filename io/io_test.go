package io

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestRCPServer(t *testing.T) {
	uds := "/tmp/dk.sock"
	GRPCServer(uds)
}

func TestRPC(t *testing.T) {
	wg := sync.WaitGroup{}

	uds := "/tmp/test.sock"

	wg.Add(1)
	go func() {
		defer wg.Done()
		GRPCServer(uds)
	}()

	time.Sleep(time.Second)

	conn, err := grpc.Dial("unix://"+uds, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		panic(err)
	}

	defer conn.Close()
	c := NewDataKitClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Send(ctx, &Request{
		Lines: []byte(strings.Join([]string{
			`test_a,tag1=val1,tag2=val2 f1=1i,f2=3,f3="abc",f4=T ` + fmt.Sprintf("%d", time.Now().UnixNano()),
			`test_b,tag1=val1,tag2=val2 f1=1i,f2=3,f3="abc",f4=T ` + fmt.Sprintf("%d", time.Now().UnixNano()),
			`test_c,tag1=val1,tag2=val2 f1=1i,f2=3,f3="abc",f4=T ` + fmt.Sprintf("%d", time.Now().UnixNano()),
		}, "\n"))})

	if err != nil {
		t.Fatal(err)
	}

	log.Printf("[C] sending %d points ok, err: %s", r.GetPoints(), r.GetErr())

	r, err = c.Send(ctx, &Request{
		Lines: []byte(strings.Join([]string{ // bad body
			`test_a tag1=val1,tag2=val2 f1=1i,f2=3,f3="abc",f4=T ` + fmt.Sprintf("%d", time.Now().UnixNano()),
			`test_b tag1=val1,tag2=val2 f1=1i,f2=3,f3="abc",f4=T ` + fmt.Sprintf("%d", time.Now().UnixNano()),
			`test_c tag1=val1,tag2=val2 f1=1i,f2=3,f3="abc",f4=T ` + fmt.Sprintf("%d", time.Now().UnixNano()),
		}, "\n"))})

	if err != nil {
		t.Fatal("should not been here")
	}

	log.Printf("[C] sending points: %d, err: %s", r.GetPoints(), r.GetErr())

	log.Printf("stopping server...")

	wg.Wait()
}

func TestMeasurement(t *testing.T) {
	tmpFields := map[string]interface{}{"year": 2020}

	data, err := MakeMetric("/dazzling_jackson", nil, tmpFields)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", data)
}

func TestMakeMetric(t *testing.T) {
	l, err := MakeMetric("abc", map[string]string{
		"t1": `c:\\\\\\\\\\\\\`,
		"t2": `\dddddd`,
		"t3": "def",
	},
		map[string]interface{}{
			"f1": uint64(time.Now().UnixNano()),
			"f2": false,
			"f3": 1.234,
			"f5": "haha",
		},
		time.Now())

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", string(l))
}
