// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"google.golang.org/protobuf/proto"
)

const (
	replayAssetAppIDField       = "appid"
	replayAssetLegacyAppIDField = "app_id"
	replayAssetTagsField        = "tags"
	replayAssetCheckBodyMaxSize = MiB // Check requests contain only JSON metadata.
)

type replayAssetUploadFailure struct {
	File   string `json:"file"`
	Reason string `json:"reason"`
}

type replayAssetUploadResp struct {
	Success  []string                   `json:"success"`
	Failures []replayAssetUploadFailure `json:"failures"`
}

type replayAssetCheckReq struct {
	AppID       string            `json:"appid"`
	LegacyAppID string            `json:"app_id"`
	Files       []string          `json:"files"`
	Tags        map[string]string `json:"tags"`
}

func (r replayAssetCheckReq) appID() string {
	if r.AppID != "" {
		return r.AppID
	}

	return r.LegacyAppID
}

func replayAssetAppID(values map[string][]string) string {
	if appID := firstReplayAssetValue(values[replayAssetAppIDField]); appID != "" {
		return appID
	}

	return firstReplayAssetValue(values[replayAssetLegacyAppIDField])
}

func firstReplayAssetValue(values []string) string {
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func parseReplayAssetTags(values []string) (map[string]string, error) {
	if len(values) == 0 || values[0] == "" {
		return nil, nil
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("multiple %s fields", replayAssetTagsField)
	}

	tags := make(map[string]string)
	if err := json.Unmarshal([]byte(values[0]), &tags); err != nil {
		return nil, fmt.Errorf("parse %s: %w", replayAssetTagsField, err)
	}

	return tags, nil
}

func replayAssetSinkTags(appID string, linkedTags map[string]string) map[string]string {
	tags := make(map[string]string, len(linkedTags)+3)
	for k, v := range linkedTags {
		tags[k] = v
	}

	tags["category"] = "session_replay"
	tags["service"] = "replay_assets"
	tags[replayAssetLegacyAppIDField] = appID

	return tags
}

func replayAssetSinkHeaders(dw *dataway.Dataway, appID string, linkedTags map[string]string) map[string]string {
	headerValue := dw.SinkHeaderValueFromTags(replayAssetSinkTags(appID, linkedTags))
	if headerValue == "" {
		return nil
	}

	return map[string]string{dw.SinkHeaderKey(): headerValue}
}

func addReplayAssetTags(formValues map[string]*ValuesSlice, linkedTags map[string]string) {
	for k, v := range linkedTags {
		switch k {
		case "category", "service", replayAssetAppIDField, replayAssetLegacyAppIDField:
			continue
		default:
			formValues[k] = &ValuesSlice{Values: []string{v}}
		}
	}
}

func buildReplayAssetFormValues(values map[string][]string, appID string, linkedTags map[string]string) map[string]*ValuesSlice {
	formValues := make(map[string]*ValuesSlice, len(values)+len(linkedTags)+1)
	for k, v := range values {
		formValues[k] = &ValuesSlice{Values: v}
	}

	formValues[replayAssetLegacyAppIDField] = &ValuesSlice{Values: []string{appID}}
	addReplayAssetTags(formValues, linkedTags)

	for k, v := range datakit.GlobalHostTags() {
		if _, ok := formValues[k]; !ok {
			formValues[k] = &ValuesSlice{Values: []string{v}}
		}
	}

	return formValues
}

func readReplayAssetRequestBody(req *http.Request, maxSize int64) ([]byte, error) {
	if req.ContentLength > maxSize {
		return nil, uhttp.Errorf(httpapi.ErrBadReq,
			"request body size [%d] exceeds the limit [%d]", req.ContentLength, maxSize)
	}

	body, err := io.ReadAll(io.LimitReader(req.Body, maxSize+1))
	if err != nil {
		return nil, uhttp.Errorf(httpapi.ErrHTTPReadErr, "%s", err)
	}
	if int64(len(body)) > maxSize {
		return nil, uhttp.Errorf(httpapi.ErrBadReq, "request body exceeds the limit [%d]", maxSize)
	}

	return body, nil
}

func (ipt *Input) handleReplayAssetsUpload(_ http.ResponseWriter, req *http.Request, _ ...interface{}) (interface{}, error) {
	if ipt.replayDiskQueue == nil {
		return nil, uhttp.Errorf(httpapi.ErrInvalidAPIHandler, "replay disk queue is nil")
	}

	body, err := readReplayAssetRequestBody(req, ReplayBodyMaxSize)
	if err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(bytes.NewReader(body))
	if err := req.ParseMultipartForm(ReplayBodyMaxSize); err != nil {
		return nil, uhttp.Errorf(httpapi.ErrBadReq, "parse multipart form: %s", err)
	}

	appID := replayAssetAppID(req.MultipartForm.Value)
	if appID == "" {
		return nil, uhttp.Errorf(httpapi.ErrBadReq, "missing %s", replayAssetAppIDField)
	}

	linkedTags, err := parseReplayAssetTags(req.MultipartForm.Value[replayAssetTagsField])
	if err != nil {
		return nil, uhttp.Errorf(httpapi.ErrBadReq, "%s", err)
	}

	files := req.MultipartForm.File["files"]
	if len(files) == 0 {
		return nil, uhttp.Errorf(httpapi.ErrBadReq, "missing files")
	}

	reqPB := &RequestPB{
		Header:     copyReplayRequestHeaders(req.Header),
		Body:       body,
		FormValues: buildReplayAssetFormValues(req.MultipartForm.Value, appID, linkedTags),
		APIPath:    datakit.SessionReplayAssetUpload,
	}

	pbData, err := proto.Marshal(reqPB)
	if err != nil {
		return nil, uhttp.Errorf(httpapi.ErrInvalidData, "marshal request: %s", err)
	}

	resp := replayAssetUploadResp{
		Success: make([]string, 0, len(files)),
	}
	for _, fh := range files {
		resp.Success = append(resp.Success, fh.Filename)
	}

	if err := ipt.replayDiskQueue.Put(pbData); err != nil {
		resp.Failures = make([]replayAssetUploadFailure, 0, len(files))
		resp.Success = nil
		for _, fh := range files {
			resp.Failures = append(resp.Failures, replayAssetUploadFailure{
				File:   fh.Filename,
				Reason: err.Error(),
			})
		}
	}

	return resp, nil
}

func (ipt *Input) handleReplayAssetsCheck(_ http.ResponseWriter, req *http.Request, _ ...interface{}) (interface{}, error) {
	dw := config.Cfg.Dataway
	if dw == nil {
		return nil, uhttp.Errorf(httpapi.ErrInvalidAPIHandler, "dataway is nil")
	}

	body, err := readReplayAssetRequestBody(req, replayAssetCheckBodyMaxSize)
	if err != nil {
		return nil, err
	}

	var payload replayAssetCheckReq
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, uhttp.Errorf(httpapi.ErrInvalidJSON, "json.Unmarshal: %s", err)
	}
	if payload.appID() == "" {
		return nil, uhttp.Errorf(httpapi.ErrBadReq, "missing %s", replayAssetAppIDField)
	}
	if len(payload.Files) == 0 {
		return nil, uhttp.Errorf(httpapi.ErrBadReq, "missing files")
	}

	resp, err := dw.CheckSessionReplayAssets(body, replayAssetSinkHeaders(dw, payload.appID(), payload.Tags))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	j, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, uhttp.NewErr(fmt.Errorf("%s", string(j)), resp.StatusCode)
	}

	return uhttp.RawJSONBody(j), nil
}

func (ipt *Input) handleReplayAssetsGet(w http.ResponseWriter, req *http.Request) {
	dw := config.Cfg.Dataway
	if dw == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	query := req.URL.Query()
	if query.Get("workspace_uuid") == "" || query.Get("app_id") == "" || query.Get("file") == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := dw.GetSessionReplayAsset(url.Values{
		"workspace_uuid": []string{query.Get("workspace_uuid")},
		"app_id":         []string{query.Get("app_id")},
		"file":           []string{query.Get("file")},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Errorf("copy replay asset response: %s", err)
	}
}
