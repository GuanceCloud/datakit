// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	pt "github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

func (d *dialer) pointsFeed(urlStr string) {
	d.seqNumber++
	startTime := time.Now()
	regionName := d.regionName()
	tags, fields := d.task.GetResults()

	if d.task.Class() == dt.ClassHeadless {
		d.processBrowserScreenshots(fields)
	}

	if status, ok := tags["status"]; ok {
		taskCheckCostSummary.WithLabelValues(regionName, d.class, status).Observe(float64(time.Since(startTime)) / float64(time.Second))
	}

	d.enrichPointResult(tags, fields, regionName)

	opt := append(pt.DefaultLoggingOptions(), pt.WithTime(d.dialingTime))
	data := pt.NewPoint(d.task.MetricName(),
		append(pt.NewTags(tags), pt.NewKVs(fields)...), opt...)

	dialWorker.addPoints(&jobData{
		url:        urlStr,
		pt:         data,
		regionName: regionName,
		class:      d.class,
	})
}

func (d *dialer) enrichPointResult(
	tags map[string]string,
	fields map[string]interface{},
	regionName string,
) {
	for k, v := range d.tags {
		if d.measurementInfo != nil && d.measurementInfo.Tags != nil {
			if _, ok := d.measurementInfo.Tags[k]; !ok {
				continue
			}
		}

		if k == LabelOwnerTag {
			tags[k] = v
			continue
		}

		if _, ok := tags[k]; !ok {
			tags[k] = v
		} else {
			l.Debugf("ignore dialer tag %s: %s", k, v)
		}
	}

	fields["seq_number"] = d.seqNumber
	fields["task_id"] = d.task.GetExternalID()
	tags["datakit_version"] = datakit.Version
	tags["node_name"] = regionName
	nodeID := d.regionID()
	if nodeID != "" {
		tags["node_id"] = nodeID
	}
	if d.task.Class() == dt.ClassNetPath {
		if tags["source_name"] == "" && regionName != "" {
			tags["source_name"] = regionName
		}
		if nodeID != "" {
			tags["source_host"] = nodeID
		} else if regionName != "" {
			tags["source_host"] = regionName
		}
		if pathKey := makeNetPathDialtestingPathKey(nodeID, d.task.GetExternalID()); pathKey != "" {
			tags["path_key"] = pathKey
		}
	}

	// df tags
	for k, v := range d.dfTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		} else {
			l.Debugf("ignore df tag %s: %s", k, v)
		}
	}

	// custom tags
	for k, v := range d.ipt.Tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		} else {
			l.Warnf("duplicate tag, ignore custom tag %s", k)
		}
	}

	triggerType := d.triggerType
	if triggerType == "" {
		triggerType = triggerTypeScheduled
	}
	tags["trigger_type"] = triggerType
	if triggerType == triggerTypeManual && d.runBatchID != "" {
		tags["run_batch_id"] = d.runBatchID
	} else {
		delete(tags, "run_batch_id")
	}
}

type browserScreenshotObject struct {
	ID   string `json:"id"`
	Date string `json:"date"`
	File string `json:"file"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

func (d *dialer) processBrowserScreenshots(fields map[string]interface{}) {
	if fields == nil {
		return
	}

	stepsText, ok := fields["steps"].(string)
	if !ok || stepsText == "" {
		return
	}

	dec := json.NewDecoder(bytes.NewBufferString(stepsText))
	dec.UseNumber()

	var steps []map[string]interface{}
	if err := dec.Decode(&steps); err != nil {
		l.Warnf("decode browser steps failed: %s", err.Error())
		return
	}

	uploadURL, err := d.browserScreenshotUploadURL()
	if err != nil {
		l.Warnf("build browser screenshot upload url failed: %s", err.Error())
		fields["screenshot_upload_error"] = err.Error()
	}

	runID, _ := fields["browser_run_id"].(string)
	uploadCount := 0
	uploadErrors := []string{}

	for _, step := range steps {
		screenshotPath, ok := step["screenshot"].(string)
		if !ok || screenshotPath == "" {
			continue
		}

		delete(step, "screenshot")
		stepSeq := browserStepSeq(step["seq"])
		if err != nil {
			uploadErrors = append(uploadErrors, err.Error())
			step["screenshot_upload_error"] = err.Error()
			cleanupBrowserScreenshot(screenshotPath)
			continue
		}

		result, contentType, uploadErr := d.uploadBrowserScreenshot(uploadURL, screenshotPath, runID, stepSeq)
		cleanupBrowserScreenshot(screenshotPath)
		if uploadErr != nil {
			l.Warnf("upload browser screenshot failed: %s", uploadErr.Error())
			uploadErrors = append(uploadErrors, uploadErr.Error())
			step["screenshot_upload_error"] = uploadErr.Error()
			continue
		}

		step["screenshot"] = browserScreenshotObject{
			ID:   result.ScreenshotID,
			Date: result.ScreenshotDate,
			File: result.FileName,
			Size: result.FileSize,
			Type: contentType,
		}
		uploadCount++
	}

	if len(uploadErrors) > 0 {
		fields["screenshot_upload_error"] = uploadErrors[0]
	}
	if uploadCount > 0 {
		fields["has_screenshot"] = true
	}

	updated, marshalErr := json.Marshal(steps)
	if marshalErr != nil {
		l.Warnf("encode browser steps failed: %s", marshalErr.Error())
		return
	}
	fields["steps"] = string(updated)
}

func cleanupBrowserScreenshot(path string) {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		l.Warnf("remove browser screenshot failed: %s", err.Error())
	}
	// The runner stores each run in its own directory. Remove it when the last
	// screenshot has gone; a non-empty directory is intentionally left alone.
	_ = os.Remove(filepath.Dir(path))
}

func (d *dialer) browserScreenshotUploadURL() (string, error) {
	u, err := url.Parse(d.task.PostURLStr())
	if err != nil {
		return "", err
	}
	u.Path = datakit.BrowserScreenshotUpload
	return u.String(), nil
}

func (d *dialer) uploadBrowserScreenshot(
	uploadURL string,
	screenshotPath string,
	runID string,
	stepSeq string,
) (*dataway.BrowserScreenshotUploadResult, string, error) {
	if dialWorker == nil || dialWorker.sender == nil {
		return nil, "", fmt.Errorf("dialtesting sender is nil")
	}
	if runID == "" {
		return nil, "", fmt.Errorf("browser_run_id is empty")
	}
	if stepSeq == "" {
		return nil, "", fmt.Errorf("step seq is empty")
	}

	data, err := os.ReadFile(screenshotPath) //nolint:gosec
	if err != nil {
		return nil, "", err
	}

	contentType := http.DetectContentType(data)
	result, err := dialWorker.sender.uploadBrowserScreenshot(uploadURL, &dataway.BrowserScreenshotUpload{
		FileName:    screenshotPath,
		ContentType: contentType,
		Data:        data,
		TaskID:      d.task.GetExternalID(),
		RunID:       runID,
		StepSeq:     stepSeq,
	})
	if err != nil {
		return nil, "", err
	}

	return result, contentType, nil
}

func browserStepSeq(v interface{}) string {
	switch x := v.(type) {
	case json.Number:
		return x.String()
	case float64:
		return strconv.FormatInt(int64(x), 10)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case string:
		return x
	default:
		return ""
	}
}
