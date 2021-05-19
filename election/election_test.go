package election

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestElection(t *testing.T) {
	const (
		electionURL  = "http://127.0.0.1:36080/election?token=tkn_test&id=dkid_test"
		heartbeatURL = "http://127.0.0.1:36080/election/heartbeat?token=tkn_test&id=dkid_test"

		successBody = `{"content":{"status":"success"}}`
		failBody    = `{"content":{"status":"fail"}}`
	)

	go func() {
		electionCount := 0
		http.HandleFunc("/election", func(resp http.ResponseWriter, req *http.Request) {
			// 只有当调用此 API 为偶数时，才 pass
			electionCount++
			if electionCount&1 == 0 {
				resp.Write([]byte(successBody))
			} else {
				resp.Write([]byte(failBody))
			}
			// resp.WriteHeader(http.StatusOK)
		})

		heartbeatCount := 0
		http.HandleFunc("/election/heartbeat", func(resp http.ResponseWriter, req *http.Request) {
			// 5 次心跳，自动结束
			heartbeatCount++
			if heartbeatCount%5 == 0 {
				resp.Write([]byte(failBody))
			} else {
				resp.Write([]byte(successBody))
			}
			// resp.WriteHeader(http.StatusOK)
		})

		fmt.Println("start http server")
		http.ListenAndServe("0.0.0.0:36080", nil)
	}()

	time.Sleep(time.Second * 2)

	csm := NewConsensusModule(electionURL, heartbeatURL)
	fmt.Printf("start, election stauts: %s\n", csm.CurrentStats())

	go csm.StartElection()

	go func() {
		index := 0
		for {
			select {
			case <-datakit.Exit.Wait():
				return
			case <-time.After(time.Second * 1):
				index++
				fmt.Printf("index: %d, election_status: %s\n", index, csm.CurrentStats())
			}
		}
	}()

	time.Sleep(time.Second * 30)
	datakit.Exit.Close()
	time.Sleep(time.Second * 1)
}
