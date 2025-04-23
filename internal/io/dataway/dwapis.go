// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

type checkTokenResult struct {
	Code      int    `json:"code"`
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
}

func (dw *Dataway) UsageTrace(body []byte) error {
	if len(dw.eps) == 0 {
		return fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	requrl, ok := ep.categoryURL[datakit.UsageTrace]
	if !ok {
		return fmt.Errorf("no workspace query URL available")
	}

	l.Debugf("NewRequest: %s", requrl)
	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	resp, err := ep.sendReq(req)
	if err != nil {
		return err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	defer resp.Body.Close() //nolint:errcheck
	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("usage trace refresh ok")
		return nil
	default:
		return fmt.Errorf("usage trace refresh failed(status: %d): %s", resp.StatusCode, string(respBody))
	}
}

func (dw *Dataway) WorkspaceQuery(body []byte) (*http.Response, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	requrl, ok := ep.categoryURL[datakit.Workspace]
	if !ok {
		return nil, fmt.Errorf("no workspace query URL available")
	}

	l.Debugf("NewRequest: %s", requrl)
	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	return ep.sendReq(req)
}

func (dw *Dataway) DQLQuery(body []byte) (*http.Response, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	requrl, ok := ep.categoryURL[datakit.QueryRaw]
	if !ok {
		return nil, fmt.Errorf("no DQL query URL available")
	}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	return dw.eps[0].sendReq(req)
}

func (dw *Dataway) Election(namespace, id string, reqBody io.Reader) ([]byte, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]

	requrl, ok := ep.categoryURL[datakit.Election]
	if !ok {
		return nil, fmt.Errorf("no election URL available")
	}

	if strings.Contains(requrl, "?token") {
		requrl += fmt.Sprintf("&namespace=%s&id=%s", namespace, id)
	} else {
		return nil, fmt.Errorf("token missing")
	}

	l.Debugf("election sending %s", requrl)

	req, err := http.NewRequest("POST", requrl, reqBody)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	resp, err := ep.sendReq(req)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	if resp == nil {
		return nil, errRequestTerminated
	}

	body, err := io.ReadAll(resp.Body)
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

func (dw *Dataway) ElectionHeartbeat(namespace, id string, reqBody io.Reader) ([]byte, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]

	requrl, ok := ep.categoryURL[datakit.ElectionHeartbeat]
	if !ok {
		return nil, fmt.Errorf("no election URL available")
	}

	if strings.Contains(requrl, "?token") {
		requrl += fmt.Sprintf("&namespace=%s&id=%s", namespace, id)
	} else {
		return nil, fmt.Errorf("token missing")
	}

	l.Debugf("election sending heartbeat %s", requrl)

	req, err := http.NewRequest("POST", requrl, reqBody)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	resp, err := ep.sendReq(req)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	if resp == nil {
		return nil, errRequestTerminated
	}

	body, err := io.ReadAll(resp.Body)
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

func (dw *Dataway) DatawayList() ([]string, int, error) {
	if len(dw.eps) == 0 {
		return nil, datawayListIntervalDefault, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	requrl, ok := ep.categoryURL[datakit.ListDataWay]
	if !ok {
		return nil, datawayListIntervalDefault, fmt.Errorf("dataway list API not available")
	}

	req, err := http.NewRequest("GET", requrl, nil)
	if err != nil {
		return nil, datawayListIntervalDefault, err
	}

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	resp, err := ep.sendReq(req)
	if err != nil {
		return nil, datawayListIntervalDefault, err
	}

	if resp == nil {
		return nil, -1, errRequestTerminated
	}

	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, datawayListIntervalDefault, err
	}

	type dataways struct {
		Content struct {
			DatawayList []string `json:"dataway_list"`
			Interval    int      `json:"interval"`
		} `json:"content"`
	}

	var dws dataways
	if err := json.Unmarshal(body, &dws); err != nil {
		l.Errorf(`%s, body: %s`, err, string(body))
		return nil, datawayListIntervalDefault, err
	}

	l.Debugf(`available dataways; %+#v,body: %s`, dws.Content, string(body))
	return dws.Content.DatawayList, dws.Content.Interval, nil
}

// UpsertObjectLabels used to create or update object labels.
func (dw *Dataway) UpsertObjectLabels(tkn string, body []byte) (*http.Response, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	requrl, ok := ep.categoryURL[datakit.ObjectLabel]
	if !ok {
		return nil, fmt.Errorf("no object labels URL available")
	}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("delete object label error: %w", err)
	}

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	return ep.sendReq(req)
}

// DeleteObjectLabels used to delete object labels.
func (dw *Dataway) DeleteObjectLabels(tkn string, body []byte) (*http.Response, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	requrl, ok := ep.categoryURL[datakit.ObjectLabel]
	if !ok {
		return nil, fmt.Errorf("no object labels URL available")
	}

	rBody := bytes.NewReader(body)
	req, err := http.NewRequest("DELETE", requrl, rBody)
	if err != nil {
		return nil, fmt.Errorf("delete object label error: %w", err)
	}

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	return ep.sendReq(req)
}

// GetEnvVariable used to get env variable.
func (dw *Dataway) GetEnvVariable(param map[string]string) (*http.Response, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	requrl, ok := ep.categoryURL[datakit.EnvVariable]
	if !ok {
		return nil, fmt.Errorf("no env_variable URL available")
	}

	if strings.Contains(requrl, "?token") {
		for k, v := range param {
			requrl += fmt.Sprintf("&%s=%s", k, v)
		}
	} else {
		return nil, fmt.Errorf("token missing")
	}

	req, err := http.NewRequest("GET", requrl, nil)
	if err != nil {
		return nil, fmt.Errorf("delete object label error: %w", err)
	}

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	return ep.sendReq(req)
}

func (dw *Dataway) GetEndpoints() []*endPoint {
	return dw.eps
}

func (dw *Dataway) UploadLog(r io.Reader, hostName string) (*http.Response, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	reqURL, ok := ep.categoryURL[datakit.LogUpload]
	if !ok {
		return nil, fmt.Errorf("no file upload URL available")
	}

	req, err := http.NewRequest("POST", reqURL, r)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	req.Header.Add("Host-Name", hostName)
	return ep.sendReq(req)
}

func (dw *Dataway) Pull(args string) ([]byte, error) {
	if dw.ClientsCount() == 0 {
		return nil, fmt.Errorf("dataway URL not set")
	}

	return dw.eps[0].datakitPull(args)
}

type ntpResp struct {
	TimestampSec int64 `json:"timestamp_sec"`
}

// TimeDiff implement ntp time sync interface.
func (dw *Dataway) TimeDiff() int64 {
	if d, err := dw.doTimeDiff(); err != nil {
		l.Errorf("doTimeDiff: %s", err.Error())
		return 0
	} else {
		return d
	}
}

func (dw *Dataway) doTimeDiff() (int64, error) {
	if len(dw.eps) == 0 {
		return 0, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	requrl, ok := ep.categoryURL[datakit.NTPSync]
	if !ok {
		return 0, fmt.Errorf("url %s not available", datakit.NTPSync)
	}

	l.Debugf("NewRequest: %s", requrl)
	req, err := http.NewRequest(http.MethodGet, requrl, nil)
	if err != nil {
		return 0, fmt.Errorf("http.NewRequest: %w", err)
	}

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	resp, err := ep.sendReq(req)
	if err != nil {
		return 0, fmt.Errorf("doSendReq: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("io.readAll: %w", err)
	}

	defer resp.Body.Close() //nolint:errcheck
	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("ntp ok")

		var nr ntpResp

		if err := json.Unmarshal(respBody, &nr); err != nil {
			l.Errorf("Unmarshal: %s", string(respBody))

			return 0, fmt.Errorf(`json.Unmarshal: %w`, err)
		}

		return nr.TimestampSec - time.Now().Unix(), nil

	default:
		return 0, fmt.Errorf("ntp failed(status: %d): %s", resp.StatusCode, string(respBody))
	}
}

func (dw *Dataway) RemoteJob(bts []byte) (resp *http.Response, err error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	reqURL, ok := ep.categoryURL[datakit.RemoteJob]
	if !ok {
		return nil, fmt.Errorf("no file upload URL available")
	}
	r := bytes.NewReader(bts)
	req, err := http.NewRequest("POST", reqURL, r)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	return ep.sendReq(req)
}
