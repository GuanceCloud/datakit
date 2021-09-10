package dialtesting

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	dt "gitlab.jiagouyun.com/cloudcare-tools/kodo/dialtesting"
)

func TestDialer(t *testing.T) {

	stopserver := make(chan interface{})
	defer close(stopserver)
	go httpServer(t, "localhost:54321", false, stopserver) // http server

	time.Sleep(time.Second)

	task := &dt.HTTPTask{
		ExternalID: cliutils.XID("dtst_"),
		Method:     "GET",
		URL:        "http://localhost:54321/_test_resp_time_less_10ms",
		Name:       "_test_resp_time_less_10ms",
		Frequency:  "1s",
		SuccessWhen: []*dt.HTTPSuccess{
			{ResponseTime: "10ms"},
		},
	}

	if err := task.Init(); err != nil {
		t.Error(err)
	}

	dialer, err := newDialer(task, nil)
	if err != nil {
		t.Error(err)
	}

	dialer.run()
	//dialer.updateTask()
}
