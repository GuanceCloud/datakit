// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collect

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var (
	datakitLastErrURL string
	feedErrInterval   = time.Second * 30
)

const InputName = "oracle"

type ExternalLastErr struct {
	Input      string `json:"input"`
	Source     string `json:"source"`
	ErrContent string `json:"err_content"`
}

func FeedLastError(errString string) error {
	lastError := &ExternalLastErr{
		Input:      InputName,
		Source:     InputName,
		ErrContent: errString,
	}

	data, err := json.Marshal(lastError)
	if err != nil {
		l.Errorf("FeedLastError json.Marshal failed: %v", err)
		return err
	}

	if err := writeData(data, datakitLastErrURL); err != nil {
		l.Errorf("FeedLastError writeData failed: %v", err)
		return err
	}

	return nil
}

func FeedLastErrorLoop(errString string, ch chan os.Signal) {
	l.Error(errString)

	ticker := time.NewTicker(feedErrInterval)
	defer ticker.Stop()

	for {
		if err := FeedLastError(errString); err != nil {
			l.Error(err)
		}

		select {
		case <-ticker.C:
		case <-ch:
			return
		}
	}
}

func ReportErrorf(format string, args ...interface{}) {
	go func() {
		l.Errorf(format, args...)

		if err := FeedLastError(fmt.Sprintf(format, args...)); err != nil {
			l.Errorf("FeedLastError failed: %v", err)
		}
	}()
}
