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
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/compact"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/endpoint"
)

// DialtestingSender used for dialtesting collector.
type DialtestingSender struct {
	ep *endpoint.EndPoint
}

type DialtestingSenderOpt struct {
	HTTPTimeout time.Duration
	HTTPProxy   string
}

type BrowserScreenshotUpload struct {
	FileName    string
	ContentType string
	Data        []byte
	TaskID      string
	RunID       string
	StepSeq     string
}

type BrowserScreenshotUploadResult struct {
	ScreenshotID   string `json:"screenshot_id"`
	ScreenshotDate string `json:"screenshot_date"`
	FileName       string `json:"file_name"`
	FileSize       int64  `json:"file_size"`
	SHA256         string `json:"sha256"`
}

func (d *DialtestingSender) Init(opt *DialtestingSenderOpt) error {
	d.ep = &endpoint.EndPoint{}
	if opt != nil {
		endpoint.WithHTTPTimeout(opt.HTTPTimeout)(d.ep)
		endpoint.WithProxy(opt.HTTPProxy)(d.ep)
	}
	return d.ep.SetupHTTP()
}

func (d *DialtestingSender) WriteData(url string, pts []*point.Point) error {
	// TODO: can not set content encoding here, default use line-protocol

	// return write error or build error
	var writeError error
	w := compact.GetWriter(
		compact.WithPoints(pts),
		compact.WithDynamicURL(url),
		compact.WithCategory(point.DynamicDWCategory),
		compact.WithHTTPEncoding(point.LineProtocol),
		compact.WithBodyCallback(func(w *compact.Writer, b *compact.Body) error {
			err := d.ep.WritePointData(w, b)
			if err != nil {
				writeError = err
			}

			return err
		}),
		compact.WithHTTPHeader("X-Sub-Category", "dialtesting"))
	defer compact.PutWriter(w)

	if d.ep == nil {
		return fmt.Errorf("endpoint is not set correctly")
	}

	buildErr := w.BuildPointsBody()

	if buildErr != nil {
		return buildErr
	}

	return writeError
}

func (d *DialtestingSender) UploadBrowserScreenshot(
	url string,
	req *BrowserScreenshotUpload,
) (*BrowserScreenshotUploadResult, error) {
	if d.ep == nil {
		return nil, fmt.Errorf("endpoint is not set correctly")
	}
	if req == nil {
		return nil, fmt.Errorf("browser screenshot upload request is nil")
	}
	if len(req.Data) == 0 {
		return nil, fmt.Errorf("browser screenshot file is empty")
	}
	if req.TaskID == "" || req.RunID == "" || req.StepSeq == "" {
		return nil, fmt.Errorf("task_id, run_id and step_seq are required")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileName := filepath.Base(req.FileName)
	if fileName == "." || fileName == string(filepath.Separator) || fileName == "" {
		fileName = "screenshot"
	}

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, escapeQuotes(fileName)))
	if req.ContentType != "" {
		header.Set("Content-Type", req.ContentType)
	}

	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(req.Data); err != nil {
		return nil, err
	}

	if err := writer.WriteField("task_id", req.TaskID); err != nil {
		return nil, err
	}
	if err := writer.WriteField("run_id", req.RunID); err != nil {
		return nil, err
	}
	if err := writer.WriteField("step_seq", req.StepSeq); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := d.ep.SendReq(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("upload browser screenshot failed(status: %d): %s", resp.StatusCode, string(respBody))
	}

	result, err := decodeBrowserScreenshotUploadResult(respBody)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON body content(%s): %w", respBody, err)
	}

	return result, nil
}

func decodeBrowserScreenshotUploadResult(respBody []byte) (*BrowserScreenshotUploadResult, error) {
	result := &BrowserScreenshotUploadResult{}
	if err := json.Unmarshal(respBody, result); err != nil {
		return nil, err
	}
	if result.ScreenshotID != "" || result.FileName != "" || result.FileSize > 0 {
		return result, nil
	}

	var wrapped struct {
		Content BrowserScreenshotUploadResult `json:"content"`
	}
	if err := json.Unmarshal(respBody, &wrapped); err != nil {
		return nil, err
	}
	return &wrapped.Content, nil
}

func escapeQuotes(s string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(s)
}

// CheckToken checks if token is valid based on the specified scheme and host.
func (d *DialtestingSender) CheckToken(token, scheme, host string) (bool, error) {
	if d.ep == nil {
		return false, fmt.Errorf("no endpoint available")
	}

	reqURL := fmt.Sprintf("%s://%s%s/%s", scheme, host, datakit.TokenCheck, token)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := d.ep.SendReq(req)
	if err != nil {
		return false, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return false, err
	}

	defer resp.Body.Close() //nolint:errcheck

	result := checkTokenResult{}

	if resp.StatusCode == 200 {
		return true, nil
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("invalid JSON body content(%s): %w", body, err)
	}

	if result.Code == 200 || len(result.ErrorCode) == 0 {
		return true, nil
	} else {
		return false, nil
	}
}
