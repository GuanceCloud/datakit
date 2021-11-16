package cmds

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var workspacecli *http.Client

const (
	workspace = "/v1/workspace"
)

type DataUsage struct {
	DataMetric  int64 `json:"data_metric"`
	DataLogging int64 `json:"data_logging"`
	DataTracing int64 `json:"data_tracing"`
	DataRUM     int64 `json:"data_rum"`
	IsOverUsage bool  `json:"is_over_usage"`
}

type Workspace struct {
	Token     *Token     `json:"token"`
	DataUsage *DataUsage `json:"data_usage"`
}

type Token struct {
	WsUUID    string `json:"ws_uuid"`
	BillState string `json:"bill_state"`
	VerType   string `json:"ver_type"`
	Token     string `json:"token"`
	DBUUID    string `json:"db_uuid"`
	Status    int    `json:"status"`
	Creator   string `json:"creator"`
	ExpireAt  int64  `json:"expire_at"`

	CreateAt int64 `json:"create_at"`
	UpdateAt int64 `json:"update_at"`
	DeleteAt int64 `json:"delete_at"`
}

func doWorkspace(requrl string) ([]byte, error) {
	var body []byte
	req, err := http.NewRequest("GET", requrl, nil)
	if err != nil {
		errorf("http.NewRequest: %s\n", err.Error())
		return body, err
	}
	workspacecli = &http.Client{}
	resp, err := workspacecli.Do(req)
	if err != nil {
		errorf("httpcli.Do: %s\n", err.Error())
		return body, err
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		errorf("ioutil.ReadAll: %s\n", err.Error())
		return body, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode/100 != 2 {
		r := struct {
			Err string `json:"error_code"`
			Msg string `json:"message"`
		}{}

		if err := json.Unmarshal(body, &r); err != nil {
			errorf("json.Unmarshal: %s\n", err.Error())
			errorf("body: %s\n", string(body))
			return body, err
		}

		errorf("[%s] %s\n", r.Err, r.Msg)
		return body, err
	}
	return body, nil
}

func outputWorkspaceInfo(body []byte) {
	r := struct {
		Content []Workspace `json:"content"`
	}{}
	if err := json.Unmarshal(body, &r); err != nil {
		errorf("json.Unmarshal:%s\n", err)
	}
	for _, content := range r.Content {
		j, err := json.MarshalIndent(content, "", defaultJSONIndent)
		if err != nil {
			errorf("json.MarshalIndent %s\n", err.Error())
		}
		output("%s\n", j)
	}
}
