package dataway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func (dw *DataWayCfg) GetLogFilter() ([]byte, error) {
	if dw.httpCli != nil {
		defer dw.httpCli.CloseIdleConnections()
	}

	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("[error] dataway url empty")
	}

	return dw.endPoints[0].getLogFilter()
}

func (dc *endPoint) getLogFilter() ([]byte, error) {
	url, ok := dc.categoryURL[datakit.LogFilter]
	if !ok {
		return nil, fmt.Errorf("LogFilter API missing, should not been here")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := dc.dw.sendReq(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getLogFilter failed with status code %d", resp.StatusCode)
	}
	defer resp.Body.Close() //nolint:errcheck

	return ioutil.ReadAll(resp.Body)
}

func (dw *DataWayCfg) DQLQuery(body []byte) (*http.Response, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0]
	requrl, ok := dc.categoryURL[datakit.QueryRaw]
	if !ok {
		return nil, fmt.Errorf("no DQL query URL available")
	}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return dw.sendReq(req)
}

func (dw *DataWayCfg) Election(namespace, id string) ([]byte, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0] // 选举相关接口只只发送给第一个 dataway

	requrl, ok := dc.categoryURL[datakit.Election]
	if !ok {
		return nil, fmt.Errorf("no election URL available")
	}

	if strings.Contains(requrl, "?token") {
		requrl += fmt.Sprintf("&namespace=%s&id=%s", namespace, id)
	} else {
		return nil, fmt.Errorf("token missing")
	}

	defer dw.httpCli.CloseIdleConnections()

	l.Debugf("election sending %s", requrl)
	resp, err := dw.httpCli.Post(requrl, "", nil)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck
	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("election %s ok", requrl)
		return body, nil
	default:
		l.Debugf("election failed: %d", resp.StatusCode)
		return nil, fmt.Errorf("election failed: %s", string(body))
	}
}

func (dw *DataWayCfg) ElectionHeartbeat(namespace, id string) ([]byte, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0] // 选举相关接口只只发送给第一个 dataway

	requrl, ok := dc.categoryURL[datakit.ElectionHeartbeat]
	if !ok {
		return nil, fmt.Errorf("no election URL available")
	}

	if strings.Contains(requrl, "?token") {
		requrl += fmt.Sprintf("&namespace=%s&id=%s", namespace, id)
	} else {
		return nil, fmt.Errorf("token missing")
	}

	defer dw.httpCli.CloseIdleConnections()

	l.Debugf("election sending heartbeat %s", requrl)
	resp, err := dw.httpCli.Post(requrl, "", nil)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck
	switch resp.StatusCode / 100 {
	case 2:
		return body, nil
	default:
		return nil, fmt.Errorf("election heartbeat failed: %s", string(body))
	}
}

func (dc *endPoint) heartBeat(data []byte) error {
	requrl, ok := dc.categoryURL[datakit.HeartBeat]
	if !ok {
		return fmt.Errorf("HeartBeat API missing, should not been here")
	}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	if dc.ontest {
		return nil
	}

	resp, err := dc.dw.sendReq(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		err := fmt.Errorf("heart beat resp err: %+#v", resp)
		return err
	}

	return nil
}

func (dw *DataWayCfg) DatawayList() ([]string, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0]
	requrl, ok := dc.categoryURL[datakit.ListDataWay]
	if !ok {
		return nil, fmt.Errorf("dataway list API not available")
	}

	req, err := http.NewRequest("GET", requrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := dw.sendReq(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type dataways struct {
		Content []string `json:"content"`
	}

	var dws dataways
	if err := json.Unmarshal(body, &dws); err != nil {
		l.Errorf(`%s, body: %s`, err, string(body))
		return nil, err
	}

	l.Debugf(`available dataways; %+#v`, dws.Content)
	return dws.Content, nil
}

func (dw *DataWayCfg) HeartBeat() error {
	if dw.httpCli != nil {
		defer dw.httpCli.CloseIdleConnections()
	}

	body := map[string]interface{}{
		"dk_uuid":   dw.Hostname, // 暂用 hostname 代之, 后将弃用该字段
		"heartbeat": time.Now().Unix(),
		"host":      dw.Hostname,
	}

	if dw.httpCli == nil {
		if err := dw.initHttp(); err != nil {
			return err
		}
	}

	bodyByte, err := json.Marshal(body)
	if err != nil {
		err := fmt.Errorf("[error] heartbeat json marshal err: %w", err)
		return err
	}

	for _, dc := range dw.endPoints {
		if err := dc.heartBeat(bodyByte); err != nil {
			return err
		}
	}

	return nil
}

// UpsertObjectLabels , dw api create or update object labels
func (dw *DataWayCfg) UpsertObjectLabels(tkn string, body []byte) (*http.Response, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0]
	requrl, ok := dc.categoryURL[datakit.ObjectLabel]
	if !ok {
		return nil, fmt.Errorf("no object labels URL available")
	}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("delete object label error: %w", err)
	}

	return dw.sendReq(req)
}

// DeleteObjectLabels , dw api delete object labels
func (dw *DataWayCfg) DeleteObjectLabels(tkn string, body []byte) (*http.Response, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0]
	requrl, ok := dc.categoryURL[datakit.ObjectLabel]
	if !ok {
		return nil, fmt.Errorf("no object labels URL available")
	}

	rBody := bytes.NewReader(body)
	req, err := http.NewRequest("DELETE", requrl, rBody)
	if err != nil {
		return nil, fmt.Errorf("delete object label error: %w", err)
	}

	return dw.sendReq(req)
}
