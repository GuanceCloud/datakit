package sinklogstash

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/telkomdev/go-stash"
)

// 检查是不是开发机，如果不是开发机，则直接退出。开发机上需要定义 LOCAL_UNIT_TEST 环境变量。
func checkDevHost() bool {
	if envs := os.Getenv("LOCAL_UNIT_TEST"); envs == "" {
		return false
	}
	return true
}

//------------------------------------------------------------------------------
type Message struct {
	Data string `json:"data"`
}

type Log struct {
	Action  string    `json:"action"`
	Time    time.Time `json:"time"`
	Message Message   `json:"message"`
}

func TestAll(t *testing.T) {
	// if !checkDevHost() {
	// 	return
	// }

	// cases := []struct {
	// 	name                  string
	// 	in                    map[string]interface{}
	// 	expectLoadConfigError error
	// 	expectWriteError      error
	// }{
	// 	{
	// 		name: "required",
	// 		in: map[string]interface{}{
	// 			"host":      "10.200.7.21:8086",
	// 			"protocol":  "http",
	// 			"precision": "ns",
	// 			"database":  "db0",
	// 			"timeout":   "6s",
	// 		},
	// 	},
	// }

	// for _, tc := range cases {
	// 	t.Run(tc.name, func(t *testing.T) {
	// 		si := &SinkLogstash{}
	// 		err := si.LoadConfig(tc.in)
	// 		assert.Equal(t, tc.expectLoadConfigError, err)

	// 		// pts := getTestPoints(t, 1000, 42)
	// 		// var newPts []sinkcommon.ISinkPoint
	// 		// for _, v := range pts {
	// 		// 	newPts = append(newPts, sinkcommon.ISinkPoint(v))
	// 		// }
	// 		// err = si.Write(newPts)

	// 		assert.Equal(t, tc.expectWriteError, err)
	// 	})
	// }

	var host string = "localhost"
	var port uint64 = 5000
	s, err := stash.Connect(host, port)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		s.Close()
	}()

	logData := Log{
		Action: "get_me",
		Time:   time.Now(),
		Message: Message{
			Data: "get me for me",
		},
	}

	logDataJSON, _ := json.Marshal(logData)

	_, err = s.Write(logDataJSON)
	if err != nil {
		// fmt.Fprintf(w, err.Error())
		fmt.Println(err)
		return
	}
}

//------------------------------------------------------------------------------
