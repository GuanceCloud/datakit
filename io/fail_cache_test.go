package io

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	pb "google.golang.org/protobuf/proto"
)

// 检查是不是开发机，如果不是开发机，则直接退出。开发机上需要定义 LOCAL_UNIT_TEST 环境变量。
func checkDevHost() bool {
	if envs := os.Getenv("LOCAL_UNIT_TEST"); envs == "" {
		return false
	}
	return true
}

func send(category string, pts []*point.Point) ([]*point.Point, error) {
	return nil, nil
}

var loop int

func call(data []byte, fn funcSend) error {
	if string(data) != dataTest[loop] {
		panic(fmt.Sprintf("not equal, left = %s, right = %s", string(data), dataTest[loop]))
	}
	loop++
	return nil
}

var dataTest = []string{
	"123",    // 3
	"1234",   // 4
	"12345",  // 5
	"123456", // 6
}

// go test -v -timeout 30s -count=1 -run ^TestPut$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io
func TestPut(t *testing.T) {
	if !checkDevHost() {
		return
	}

	cases := []struct {
		name       string
		inFilePath string
		inData     []string
		inCap      int64
		outSize    uint64
	}{
		{
			name:       "normal",
			inFilePath: "/Users/mac/Desktop/cache/file",
			inData:     dataTest,
			inCap:      1024,
			outSize:    18,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// clear target files
			err := os.RemoveAll(tc.inFilePath)
			assert.NoError(t, err)

			// write target
			fc, err := initFailCache(tc.inFilePath, tc.inCap)
			assert.NoError(t, err)
			for _, v := range tc.inData {
				err = fc.put([]byte(v))
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.outSize, fc.bytesPut)
		})
	}
}

// go test -v -timeout 30s -count=1 -run ^TestGet$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io
func TestGet(t *testing.T) {
	if !checkDevHost() {
		return
	}

	cases := []struct {
		name       string
		inFilePath string
		inCap      int64
		inLoop     int
		outSize    uint64
	}{
		{
			name:       "normal",
			inFilePath: "/Users/mac/Desktop/cache/file",
			inCap:      1024,
			inLoop:     4,
			outSize:    18,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			loop = 0
			// read target
			fc, err := initFailCache(tc.inFilePath, tc.inCap)
			assert.NoError(t, err)
			for i := 0; i < tc.inLoop; i++ {
				err = fc.get(call, send)
				if err != nil && err.Error() == "not found" {
					panic(err)
				}
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.outSize, fc.bytesTruncated)
			loop = 0
		})
	}
}

//------------------------------------------------------------------------------

var dataCategory = map[string]string{
	"metric":  "1234",   // 4
	"network": "123456", // 6
}

func pbCall(data []byte, fn funcSend) error {
	pd := &PBData{}
	if err := pb.Unmarshal(data, pd); err != nil {
		return err
	}
	gotCategory := pd.GetCategory()
	if v, ok := dataCategory[gotCategory]; ok {
		if string(pd.Lines) != v {
			panic(fmt.Sprintf("not equal, left = %s, right = %s", string(pd.Lines), v))
		}
	} else {
		panic(fmt.Sprintf("not found category: %s", gotCategory))
	}
	loop++
	return nil
}

// go test -v -timeout 30s -count=1 -run ^TestPbPut$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io
func TestPbPut(t *testing.T) {
	if !checkDevHost() {
		return
	}

	cases := []struct {
		name       string
		inFilePath string
		inData     map[string]string
		inCap      int64
		outSize    uint64
	}{
		{
			name:       "normal",
			inFilePath: "/Users/mac/Desktop/cache/pbfile",
			inData:     dataCategory,
			inCap:      1024,
			outSize:    31,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// clear target files
			err := os.RemoveAll(tc.inFilePath)
			assert.NoError(t, err)

			// write target
			fc, err := initFailCache(tc.inFilePath, tc.inCap)
			assert.NoError(t, err)
			for k, v := range tc.inData {
				buf, err := pb.Marshal(&PBData{
					Category: k,
					Lines:    []byte(v),
				})
				assert.NoError(t, err)
				err = fc.put(buf)
				assert.NoError(t, err)
				fmt.Printf("written: category = %s, lines = %s\n", k, v)
			}
			assert.Equal(t, tc.outSize, fc.bytesPut)
		})
	}
}

// go test -v -timeout 30s -count=1 -run ^TestPbGet$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io
func TestPbGet(t *testing.T) {
	if !checkDevHost() {
		return
	}

	cases := []struct {
		name       string
		inFilePath string
		inData     map[string]string
		inCap      int64
		inLoop     int
		outSize    uint64
	}{
		{
			name:       "normal",
			inFilePath: "/Users/mac/Desktop/cache/pbfile",
			inData:     dataCategory,
			inCap:      1024,
			inLoop:     2,
			outSize:    31,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			loop = 0
			// read target
			fc, err := initFailCache(tc.inFilePath, tc.inCap)
			assert.NoError(t, err)
			for i := 0; i < tc.inLoop; i++ {
				err = fc.get(pbCall, send)
				if err != nil && err.Error() == "not found" {
					panic(err)
				}
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.outSize, fc.bytesTruncated)
			loop = 0
		})
	}
}

//------------------------------------------------------------------------------

func printCall(data []byte, fn funcSend) error {
	pd := &PBData{}
	if err := pb.Unmarshal(data, pd); err != nil {
		return err
	}
	fmt.Printf("got: category = %s, lines = %s\n", pd.GetCategory(), string(pd.GetLines()))
	return nil
}

// go test -v -timeout 30s -count=1 -run ^TestPbPrintAll$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io
func TestPbPrintAll(t *testing.T) {
	if !checkDevHost() {
		return
	}

	cases := []struct {
		name       string
		inFilePath string
		inCap      int64
	}{
		{
			name: "normal",
			// inFilePath: "/Users/mac/Desktop/cache/pbfile",
			inFilePath: "/usr/local/datakit/cache/network",
			inCap:      1024,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// read target
			fc, err := initFailCache(tc.inFilePath, tc.inCap)
			assert.NoError(t, err)
			first, err := fc.l.FirstIndex()
			assert.NoError(t, err)
			last, err := fc.l.LastIndex()
			assert.NoError(t, err)
			if first == 0 && last == 0 {
				panic("empty")
			}
			length := last - first + 1
			fmt.Printf("length = %d\n", length)
			for i := uint64(0); i < length; i++ {
				err = fc.get(printCall, send)
				assert.NoError(t, err)
			}
		})
	}
}
