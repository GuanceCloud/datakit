// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package textparser

import (
	"encoding/json"
	"fmt"
)

type dockerjson struct {
	Time   string `json:"time"`
	Log    string `json:"log"`
	Stream string `json:"stream"`
}

func ParseDockerJSONLog(log []byte, msg *LogMessage) error {
	var d dockerjson

	if err := json.Unmarshal(log, &d); err != nil {
		return fmt.Errorf("could to parsed docker-json log %w", err)
	}

	// Parse timestamp
	// msg.Timestamp, err = time.Parse(timeFormatIn, d.Time)
	// if err != nil {
	// 	return fmt.Errorf("unexpected timestamp format %q: %w", timeFormatIn, err)
	// }

	// Parse stream type
	msg.Stream = LogStreamType(d.Stream)
	if msg.Stream != Stdout && msg.Stream != Stderr {
		return fmt.Errorf("unexpected stream type %q", msg.Stream)
	}

	if len(d.Log) > 0 {
		msg.IsPartial = d.Log[len(d.Log)-1] != '\n'
		msg.Log = []byte(d.Log)
	}

	return nil
}
