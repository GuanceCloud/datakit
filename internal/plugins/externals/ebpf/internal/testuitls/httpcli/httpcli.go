package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/petermattis/goid"
	"golang.org/x/sys/unix"
)

var log = logger.DefaultSLogger("httpcli")

func main() {
	fmt.Fprintf(os.Stdout, "pid %d\n", unix.Getpid())

	rawURL := "http://127.0.1.2:61095"
	_, err := url.Parse(rawURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("remote_addr: %s\n", rawURL)
	buf := make([]byte, 4096)

	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		data := buf[:n-1]
		tmpURL, _ := url.JoinPath(rawURL, string(data))
		req, err := http.NewRequest("POST", tmpURL, bytes.NewReader(data))
		if err != nil {
			log.Fatal(err)
		}

		log.Info("goid ", goid.Get())
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
		}

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		_ = resp.Body.Close()

		log.Infof("status: %s, resp: %s\n", resp.Status, data)
	}
}
