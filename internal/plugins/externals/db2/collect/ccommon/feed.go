// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ccommon

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
)

var (
	DatakitLastErrURL string
	feedErrInterval   = time.Second * 30
)

type ExternalLastErr struct {
	Input      string `json:"input"`
	Source     string `json:"source"`
	ErrContent string `json:"err_content"`
}

func FeedLastError(inputName string, l *logger.Logger, errString string) error {
	lastError := &ExternalLastErr{
		Input:      inputName,
		Source:     inputName,
		ErrContent: errString,
	}

	data, err := json.Marshal(lastError)
	if err != nil {
		l.Errorf("FeedLastError json.Marshal failed: %v", err)
		return err
	}

	if err := WriteData(l, data, DatakitLastErrURL); err != nil {
		l.Errorf("FeedLastError WriteData failed: %v", err)
		return err
	}

	return nil
}

func FeedLastErrorLoop(inputName string, l *logger.Logger, errString string, ch chan os.Signal) {
	l.Error(errString)

	ticker := time.NewTicker(feedErrInterval)
	defer ticker.Stop()

	for {
		if err := FeedLastError(inputName, l, errString); err != nil {
			l.Error(err)
		}

		select {
		case <-ticker.C:
		case <-ch:
			return
		}
	}
}

func ReportErrorf(inputName string, l *logger.Logger, format string, args ...interface{}) {
	go func() {
		l.Errorf(format, args...)

		if err := FeedLastError(inputName, l, fmt.Sprintf(format, args...)); err != nil {
			l.Errorf("FeedLastError failed: %v", err)
		}
	}()
}

type DFEvent struct {
	Name        string
	EventPrefix string
	Tags        map[string]string

	Date    int64  `json:"date"`
	Status  string `json:"df_status"`
	Title   string `json:"df_title"`
	Message string `json:"df_message"`
}

const (
	dfMonitor      = "monitor"
	EventPrefixDb2 = "[Db2] " //nolint:stylecheck
)

func FeedEvent(l *logger.Logger, event *DFEvent, urlPath string) {
	opts := point.CommonLoggingOptions()

	fields := make(map[string]interface{})
	fields["date"] = event.Date
	fields["df_source"] = dfMonitor
	fields["df_status"] = event.Status
	fields["df_title"] = event.EventPrefix + event.Title + " : " + event.Message

	pt := point.NewPointV2(event.Name,
		append(point.NewTags(event.Tags), point.NewKVs(fields)...),
		opts...)

	if err := WriteData(l, []byte(pt.LineProto()), urlPath); err != nil {
		return
	}
}
