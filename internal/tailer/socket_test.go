package tailer

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func Test_socketLogger_doSocket(t *testing.T) {
	sleep := time.Second * 10
	// 打开tcpserver
	lis, err := net.Listen("tcp", ":4712")
	if err != nil {
		fmt.Println(err)
		return
	}
	sl := &socketLogger{
		pending:         make([]worker.TaskData, 0),
		socketBufferLen: 200,
		opt:             &Option{log: logger.DefaultSLogger("testcast")},
		msg:             make(chan *sample, 10),
	}
	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				continue
			}
			go sl.doSocket(conn)
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			select {
			case d := <-sl.msg:
				fmt.Println("case date now")
				for i, line := range d.data {
					fmt.Printf("i=%d  line=%s \n", i, line)
				}
			case <-time.After(time.Second * 10):
				sleep = 0
				return
			}
		}
	}()

	for {
		time.Sleep(sleep)
		if sleep == 0 {
			_ = lis.Close()
			break
		}
	}
}

func Test_spiltBuffer(t *testing.T) {
	type args struct {
		fromCache string
		date      string
		full      bool
	}
	tests := []struct {
		name          string
		args          args
		wantPipdata   []string
		wantCacheDate string
	}{
		// 0055-rc.local-exist update to monitor\n0055-rc.local-exist update to monitor
		{
			name: "case01", args: args{
				fromCache: "",
				date: `0055-rc.local-exist update to monitor
0055-rc.local-exist update to`, full: true,
			},
			wantCacheDate: "0055-rc.local-exist update to",
			wantPipdata:   []string{"0055-rc.local-exist update to monitor"},
		},

		{
			name: "case02", args: args{
				fromCache: "0055-rc",
				date: `.local-exist update to monitor
0055-rc.local-exist update to`, full: true,
			},
			wantCacheDate: "0055-rc.local-exist update to",
			wantPipdata:   []string{"0055-rc.local-exist update to monitor"},
		},

		{
			name: "case03", args: args{
				fromCache: "",
				date: `2021-12-22T14:12:42 ERROR internal.lua luafuncs/monitor.go:297  0055update to mon
0055-rc.local-exist update to
`, full: false,
			},
			wantCacheDate: "",
			wantPipdata:   []string{"2021-12-22T14:12:42 ERROR internal.lua luafuncs/monitor.go:297  0055update to mon", "0055-rc.local-exist update to"},
		},
	}
	sl := &socketLogger{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPipdata, gotCacheDate := sl.spiltBuffer(tt.args.fromCache, tt.args.date, tt.args.full)
			if !reflect.DeepEqual(gotPipdata, tt.wantPipdata) {
				t.Errorf("gotPipdata len=%d want len=%d", len(gotPipdata), len(tt.wantCacheDate))
				t.Errorf("spiltBuffer() gotPipdata = %v, want %v", gotPipdata, tt.wantPipdata)
			}
			if gotCacheDate != tt.wantCacheDate {
				t.Errorf("spiltBuffer() gotCacheDate = %v, want %v", gotCacheDate, tt.wantCacheDate)
			}
		})
	}
}

func Test_socketLogger_startReceive(t *testing.T) {
	type fields struct {
		count           int
		socketBufferLen int
		ignorePatterns  []string
		opt             *Option
		stop            chan struct{}
		pendingLock     *sync.Mutex
		pending         []worker.TaskData
		msg             chan *sample
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "test receive", fields: fields{
			pending: make([]worker.TaskData, 0), ignorePatterns: make([]string, 0),
			opt: &Option{Source: "scheck"}, stop: make(chan struct{}, 1),
			pendingLock: new(sync.Mutex), msg: make(chan *sample, 1), count: 20,
		}},

		{name: "test receive count 50", fields: fields{
			pending: make([]worker.TaskData, 0), ignorePatterns: make([]string, 0),
			opt: &Option{Source: "scheck"}, stop: make(chan struct{}, 1),
			pendingLock: new(sync.Mutex), msg: make(chan *sample, 1), count: 50,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := &socketLogger{
				socketBufferLen: tt.fields.socketBufferLen,
				opt:             tt.fields.opt,
				stop:            tt.fields.stop,
				pending:         tt.fields.pending,
				msg:             tt.fields.msg,
			}
			t.Logf("start func case %s..", tt.name)
			stopChan := make(chan int, 1)
			go func() {
				sl.startReceive()
				// wait startReceive() stop.send stop channel
				stopChan <- 0
			}()

			go func() {
				for i := 0; i < tt.fields.count; i++ {
					sl.msg <- &sample{data: []string{"123"}}
				}
			}()

			go func() {
				time.Sleep(time.Second * 5)
				t.Log("set stop channel")
				sl.stop <- struct{}{}
			}()

			<-stopChan
			t.Logf("func %s is stop ,return", tt.name)
		})
	}
}
