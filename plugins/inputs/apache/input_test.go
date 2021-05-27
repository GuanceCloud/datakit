package apache

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"sync"
	"time"
)

var s = `127.0.0.1
ServerVersion: Apache/2.4.29 (Ubuntu)
ServerMPM: event
Server Built: 2020-08-12T21:33:25
CurrentTime: Thursday, 13-May-2021 05:40:08 UTC
RestartTime: Thursday, 13-May-2021 03:26:29 UTC
ParentServerConfigGeneration: 64
ParentServerMPMGeneration: 63
ServerUptimeSeconds: 8018
ServerUptime: 2 hours 13 minutes 38 seconds
Load1: 0.43
Load5: 0.46
Load15: 0.39
Total Accesses: 26
Total kBytes: 18
CPUUser: .02
CPUSystem: .01
CPUChildrenUser: 0
CPUChildrenSystem: 0
CPULoad: .000374158
Uptime: 8018
ReqPerSec: .0032427
BytesPerSec: 2.29883
BytesPerReq: 708.923
BusyWorkers: 1
IdleWorkers: 49
ConnsTotal: 1
ConnsAsyncWriting: 0
ConnsAsyncKeepAlive: 0
ConnsAsyncClosing: 0
Scoreboard: W_________________________________________________....................................................................................................`

func TestParse(t *testing.T) {
	body := strings.NewReader(s)
	n := Input{}
	n.parse(body)

	var m Measurement
	m.LineProto()
	m.Info()
}

func TestGetMetric(t *testing.T) {
	opt := &testutil.HTTPServerOptions{
		Bind: ":12345",
		Exit: make(chan interface{}),
		Routes: map[string]func(*gin.Context){
			"/server_status": func(c *gin.Context) {
				c.Writer.Header().Set("Content-Type", "text/plain; charset=ISO-8859-1")
				c.Writer.Write([]byte(s))
			},
		},
	}

	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()
		testutil.NewHTTPServer(t, opt)
	}()

	time.Sleep(time.Second)

	n := Input{
		Url:      "http://127.0.0.1:12345/server_status",
		Interval: datakit.Duration{Duration: time.Second * 1},
	}
	go func() {
		time.Sleep(time.Second * 2)
		datakit.Exit.Close()
		close(opt.Exit)
	}()
	n.Run()
	wg.Wait()

}
