// Package utils wraps eBPF-network utils.
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

var DataKitAPIServer = "0.0.0.0:9529"

var dkLastErr = "/v1/lasterror"

type ExternalLastErr struct {
	Input      string `json:"input"`
	ErrContent string `json:"err_content"`
}

func FeedLastError(extnlErr ExternalLastErr) error {
	lastErrPostURL := fmt.Sprintf("http://%s%s", DataKitAPIServer, dkLastErr)
	client := http.Client{}
	data, err := json.Marshal(extnlErr)
	if err != nil {
		return err
	}
	rq, err := http.NewRequest("POST", lastErrPostURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	rsp, err := client.Do(rq)
	if err != nil {
		return err
	}

	defer rsp.Body.Close() //nolint:errcheck

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("lastErrPostURL, http status code: %d", rsp.StatusCode)
	}

	return fmt.Errorf("%s %d", lastErrPostURL, rsp.StatusCode)
}
