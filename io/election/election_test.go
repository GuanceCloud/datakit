package election

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

func TestElection(t *testing.T) {

	ok := []byte(`{"content": {"status": "success"}}`)
	fail := []byte(`{"content": {"status": "fail"}}`)

	bind := ":12345"
	cases := []struct {
		dataways   []string
		candidates map[int]*candidate
	}{
		{
			dataways: []string{fmt.Sprintf("http://%s?token=tkn_123", bind)},
			candidates: map[int]*candidate{
				1: &candidate{status: statusFail},
				2: &candidate{status: statusFail},
				3: &candidate{status: statusFail},
				4: &candidate{status: statusFail},
				5: &candidate{status: statusFail},
				//&candidate{status: statusFail, ttl: 5 * time.Second},
			},
		},
	}

	elected := []string{}
	var heartbeat time.Time

	setElected := func(id string) bool {

		if heartbeat.UnixNano() > 0 {
			if time.Since(heartbeat) > 2*electionInterval {
				l.Debugf("%s clear elected candidates", id)
				elected = elected[:0] //clear
			}
		}

		if len(elected) == 0 {
			elected = append(elected, id)
			heartbeat = time.Now()
			return true
		}
		return false
	}

	setHeartbeat := func(id string) bool {

		if len(elected) == 1 {
			if elected[0] == id {
				return false
			} else {
				heartbeat = time.Now()
				return true
			}
		} else {
			l.Errorf("should not been here")
			return false
		}
	}

	opt := &tu.HTTPServerOptions{
		Bind: bind,
		Exit: make(chan interface{}),
		Routes: map[string]func(*gin.Context){
			datakit.Election: func(c *gin.Context) {
				tkn := c.Query("token")

				if len(tkn) == 0 {
					c.Data(403, "", []byte(`token missing`))
					return
				}

				id := c.Query("id")
				if len(id) == 0 {
					c.Data(403, "", []byte(`id missing`))
					return
				}

				//l.Debugf("%s electing...", id)

				if setElected(id) {
					l.Debugf("%s elected", id)
					c.Writer.Write(ok)
				} else {
					c.Writer.Write(fail)
				}
			},

			datakit.ElectionHeartbeat: func(c *gin.Context) {
				tkn := c.Query("token")

				if len(tkn) == 0 {
					c.Data(403, "", []byte(`token missing`))
					return
				}

				id := c.Query("id")
				if len(id) == 0 {
					c.Data(403, "", []byte(`id missing`))
					return
				}

				if setHeartbeat(id) {
					l.Debugf("candidate %s heartbeat ok", id)
					c.Writer.Write(ok)
				} else {
					l.Debugf("candidate %s heartbeat failed", id)
					c.Writer.Write(fail)
				}
			},
		},
	}

	wg := sync.WaitGroup{}
	go func() {
		defer wg.Done()
		tu.NewHTTPServer(t, opt)
	}()
	time.Sleep(time.Second)

	electionInterval = 30 * time.Millisecond // reduce election interval

	for _, tc := range cases {
		dw := &dataway.DataWayCfg{URLs: tc.dataways}
		dw.Apply()

		for n := 0; n < 10; n++ {
			l.Debug("-------------------------")
			for k, c := range tc.candidates {
				id := fmt.Sprintf("id-%d", k)
				c.id = id
				c.dw = dw
				c.runOnce()
				time.Sleep(time.Millisecond * 500)
			}
		}

		for _, c := range tc.candidates {
			t.Logf("%+#v", c)
		}
	}

	close(opt.Exit)
	wg.Wait()
}
