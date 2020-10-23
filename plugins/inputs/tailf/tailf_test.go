// +build !solaris

package tailf

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestMain(t *testing.T) {
	io.TestOutput()

	w1 := writeFile1()
	time.Sleep(time.Second)

	var tailer = Tailf{
		LogFiles:      []string{w1},
		FromBeginning: true,
		Source:        "Testing",
	}

	go tailer.Run()

	time.Sleep(time.Second * 15)
	datakit.Exit.Close()
	time.Sleep(time.Second)
}

func TestFromBegin(t *testing.T) {
	io.TestOutput()

	w1 := writeFile2()
	///defer os.Remove(w1)
	var tailer = Tailf{
		LogFiles:      []string{w1},
		FromBeginning: true,
		Source:        "Testing",
	}

	go tailer.Run()
	time.Sleep(time.Second * 15)
	datakit.Exit.Close()
}

func writeFile1() (filepath string) {
	file, err := ioutil.TempFile("", "example")
	if err != nil {
		panic(err)
	}

	count := 0
	go func() {
		for {
			select {
			case <-datakit.Exit.Wait():
				os.Remove(file.Name())
				return
			default:
				str := fmt.Sprintf("%s -- %d\n%s\n%s\n%s\n",
					time.Now().Format(time.RFC3339Nano), count,
					"   SPACE", "\rRETURN", "\tTAB")
				file.WriteString(str)
				time.Sleep(200 * time.Millisecond)
				count++
			}
		}
	}()

	return file.Name()
}

func writeFile2() (filepath string) {
	content := `2020-10-23 06:38:30,215 INFO /usr/local/lib/python3.6/dist-packages/werkzeug/_internal.py  * Running on http://0.0.0.0:8080/ (Press CTRL+C to quit) 
2020-10-23 06:41:56,688 INFO demo.py 1.0 
2020-10-23 06:41:56,691 INFO /usr/local/lib/python3.6/dist-packages/werkzeug/_internal.py 192.168.56.1 - - [23/Oct/2020 06:41:56] "[37mGET /1 HTTP/1.1[0m" 200 - 
2020-10-23 06:41:57,041 INFO /usr/local/lib/python3.6/dist-packages/werkzeug/_internal.py 192.168.56.1 - - [23/Oct/2020 06:41:57] "[33mGET /favicon.ico HTTP/1.1[0m" 404 - 
2020-10-23 06:50:43,465 INFO /usr/local/lib/python3.6/dist-packages/werkzeug/_internal.py 192.168.56.1 - - [23/Oct/2020 06:50:43] "[33mGET / HTTP/1.1[0m" 404 - 
2020-10-23 06:53:16,062 INFO demo.py 0.5 
2020-10-23 06:53:16,064 INFO /usr/local/lib/python3.6/dist-packages/werkzeug/_internal.py 192.168.56.1 - - [23/Oct/2020 06:53:16] "[37mGET /2 HTTP/1.1[0m" 200 - 
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET] 
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 1952, in full_dispatch_request
    rv = self.handle_user_exception(e)
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 1821, in handle_user_exception
    reraise(exc_type, exc_value, tb)
  File "/usr/local/lib/python3.6/dist-packages/flask/_compat.py", line 39, in reraise
    raise value
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 1950, in full_dispatch_request
    rv = self.dispatch_request()
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 1936, in dispatch_request
    return self.view_functions[rule.endpoint](**req.view_args)
  File "demo.py", line 23, in index
    value=1/id
ZeroDivisionError: division by zero
2020-10-23 06:54:20,173 INFO /usr/local/lib/python3.6/dist-packages/werkzeug/_internal.py 192.168.56.1 - - [23/Oct/2020 06:54:20] "[35m[1mGET /0 HTTP/1.1[0m" 500 - 
`
	file, err := ioutil.TempFile("", "example")
	if err != nil {
		panic(err)
	}
	file.WriteString(content)
	return file.Name()
}
