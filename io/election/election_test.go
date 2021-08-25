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

type mockInput struct {
	candidatesID string
	inputName    string
	t            *testing.T
}

func (i *mockInput) Pause() error {
	i.t.Logf("%s/%s pause ok", i.candidatesID, i.inputName)
	return nil
}

func (i *mockInput) Resume() error {
	i.t.Logf("%s/%s resume ok", i.candidatesID, i.inputName)
	return nil
}

func TestElection(t *testing.T) {

	respOk := []byte(`{"content": {"status": "success"}}`)
	respFail := []byte(`{"content": {"status": "fail"}}`)

	bind := ":12345"
	cases := []struct {
		dataways   []string
		candidates []*candidate
	}{
		{
			dataways: []string{fmt.Sprintf("http://%s?token=tkn_123", bind)},
			candidates: []*candidate{
				&candidate{status: statusFail},
				&candidate{status: statusFail},
				&candidate{status: statusFail},
				&candidate{status: statusFail},
				&candidate{status: statusFail},

				&candidate{status: statusFail},
				&candidate{status: statusFail},
				&candidate{status: statusFail},
				&candidate{status: statusFail},
				&candidate{status: statusFail},

				&candidate{status: statusFail},
				&candidate{status: statusFail},
				&candidate{status: statusFail},
				&candidate{status: statusFail},
				&candidate{status: statusFail},
			},
		},
	}

	type tryElect struct {
		id string    // candidate id
		ok chan bool // elect ok or fail
	}

	type tryHeartbeat struct {
		id string
		ok chan bool
	}

	wg := sync.WaitGroup{}
	chElection := make(chan *tryElect)
	chHeartbeat := make(chan *tryHeartbeat)
	chExit := make(chan interface{})

	wg.Add(1)
	go func() { // mocked election server
		defer wg.Done()
		var elected string
		var lastHeartbeat time.Time

		for {
			select {
			case e := <-chElection:
				if elected == "" || time.Since(lastHeartbeat) > 300*time.Millisecond {
					elected = e.id
					lastHeartbeat = time.Now()

					t.Logf("%s elected", e.id)
					e.ok <- true
				} else {
					e.ok <- false
				}
			case hb := <-chHeartbeat:
				if hb.id == elected {
					if time.Since(lastHeartbeat) > 300*time.Millisecond {
						t.Logf("%s dopped on too late heartbeat", hb.id)
						elected = "" // reset candidator
						hb.ok <- false
					} else {
						lastHeartbeat = time.Now()
						hb.ok <- true
					}
				} else {
					hb.ok <- false
				}
			case <-chExit:
				break
			}
		}
	}()

	opt := &tu.HTTPServerOptions{ // mocked election API
		Bind: bind,
		Exit: chExit,
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

				ok := make(chan bool)
				defer close(ok)
				chElection <- &tryElect{id: id, ok: ok}
				select {
				case elected := <-ok:
					if elected {
						c.Writer.Write(respOk)
					} else {
						c.Writer.Write(respFail)
					}
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

				ok := make(chan bool)
				defer close(ok)
				chHeartbeat <- &tryHeartbeat{id: id, ok: ok}
				select {
				case hbOK := <-ok:
					if hbOK {
						c.Writer.Write(respOk)
					} else {
						c.Writer.Write(respFail)
					}
				}
			},
		},
	}

	go func() {
		defer wg.Done()
		tu.NewHTTPServer(t, opt)
		t.Logf("election server exited")
	}()
	time.Sleep(time.Second) // wait HTTP server

	wg2 := sync.WaitGroup{}

	maxElected := 3

	for _, tc := range cases {
		dw := &dataway.DataWayCfg{URLs: tc.dataways}
		dw.Apply()

		for k, c := range tc.candidates {
			id := fmt.Sprintf("id-%d", k)
			c.id = id
			c.dw = dw

			mi := &mockInput{candidatesID: id, inputName: fmt.Sprintf("%d-input", k), t: t}
			c.plugins = append(c.plugins, mi)

			wg2.Add(1)
			go func(x *candidate) {
				defer wg2.Done()
				for {

					x.runOnce()

					time.Sleep(time.Duration(x.nHeartbeat+1) * 100 * time.Millisecond) // heart beat timeout
					if x.nElected >= maxElected {
						t.Logf("%s(%d/%d/%d) exited", x.id, x.nElected, x.nHeartbeat, x.nOffline)
						return
					}
				}
			}(c)
		}
	}

	t.Logf("wait candidates done")
	wg2.Wait()

	close(chExit)

	t.Logf("wait done")
	wg.Wait()
}
