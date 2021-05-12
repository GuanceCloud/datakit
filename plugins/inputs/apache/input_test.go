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
ServerVersion: Apache/2.4.41 (Unix)
ServerMPM: event
Server Built: Aug 14 2019 00:35:10
CurrentTime: Friday, 16-Aug-2019 12:38:40 UTC
RestartTime: Wednesday, 14-Aug-2019 07:58:26 UTC
ParentServerConfigGeneration: 1
ParentServerMPMGeneration: 0
ServerUptimeSeconds: 189613
ServerUptime: 2 days 4 hours 40 minutes 13 seconds
Load1: 4.60
Load5: 1.20
Load15: 0.47
Total Accesses: 27860
Total kBytes: 33011
Total Duration: 54118
CPUUser: 18.02
CPUSystem: 31.76
CPUChildrenUser: 0
CPUChildrenSystem: 0
CPULoad: .0262535
Uptime: 189613
ReqPerSec: .146931
BytesPerSec: 178.275
BytesPerReq: 1213.33
DurationPerReq: 1.9425
BusyWorkers: 7
IdleWorkers: 93
Processes: 4
Stopping: 0
BusyWorkers: 7
IdleWorkers: 93
ConnsTotal: 13
ConnsAsyncWriting: 0
ConnsAsyncKeepAlive: 5
ConnsAsyncClosing: 0
Scoreboard: __________________________________________W_____________W___________________LW_____W______W_W_______.........................`

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
		Interval: datakit.Duration{Duration: time.Second * 2},
	}
	go func() {
		time.Sleep(time.Second * 5)
		datakit.Exit.Close()
	}()
	n.Run()
	close(opt.Exit)
	wg.Wait()

}
