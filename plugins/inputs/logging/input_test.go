package logging

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var testcase = []struct {
	text string
}{
	{
		text: `2020-10-23 06:38:30,215 INFO /usr/local/lib/python3.6/dist-packages/werkzeug/_internal.py  * Running on http://0.0.0.0:8080/ (Press CTRL+C to quit)`,
	},
	{
		text: `2020-10-23 06:41:56,688 INFO demo.py 1.0`,
	},
	{
		text: `2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET] 
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
ZeroDivisionError: division by zero`,
	},
	{
		text: ` `,
	},
	{
		text: `2020-10-23 06:41:56,688 INFO demo.py 5.0`,
	},
	{
		text: ` `,
	},
	{
		text: `2020-10-23 06:41:56,688 INFO demo.py 6.0`,
	},
}

func TestMain(t *testing.T) {
	io.SetTest()

	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	// 最后一条message只有在新数据产生以后才会发送

	tailer := Input{
		LogFiles:       []string{file.Name()},
		FromBeginning:  true,
		Source:         "testing",
		MultilineMatch: `^\d{4}-\d{2}-\d{2}`, // Match: `^\S`
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		tailer.Run()
	}()

	for _, tc := range testcase {
		time.Sleep(time.Millisecond * 500)
		file.WriteString(tc.text)
		file.WriteString("\n")
	}

	// FIXME:
	// tailf 默认每隔 10 秒扫描一次文件路径，导致程序运行时，前 10 秒是荒废的
	time.Sleep(time.Second * 13)

	datakit.Exit.Close()

	wg.Wait()
}
