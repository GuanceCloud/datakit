// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"regexp"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

type DKEvent struct {
	Status   string `json:"status"` // info/warning/error
	Message  string `json:"message"`
	Category string `josn:"category"`
	Logtype  string `josn:"Logtype"`
}

func (e *DKEvent) Tags() map[string]string {
	tags := map[string]string{
		"status":   "info",
		"category": "default",
		"log_type": "",
		"service":  "datakit",
	}

	if e.Status != "" {
		tags["status"] = e.Status
	}

	if e.Category != "" {
		tags["category"] = e.Category
	}

	if e.Logtype != "" {
		tags["log_type"] = e.Logtype
	}

	return tags
}

func (e *DKEvent) Fields() map[string]interface{} {
	fields := map[string]interface{}{
		"message": "",
	}

	if len(e.Message) > 0 {
		fields["message"] = e.escape(e.Message)
	}

	return fields
}

var tokenRegex = regexp.MustCompile(`token=(\w+)`)

func (e *DKEvent) escape(message string) string {
	escapedMessage := tokenRegex.ReplaceAllString(message, "token=xxxxxx")
	return escapedMessage
}

func FeedEventLog(e *DKEvent) {
	e.feedLog()
}

func (e *DKEvent) feedLog() {
	pt, err := point.NewPoint("datakit", e.Tags(), e.Fields(), &point.PointOption{
		Time:     time.Now(),
		Category: datakit.Logging,
	})
	if err != nil {
		log.Errorf("NewPoint: %s", err.Error())
		return
	}

	if err := Feed("dkevent", datakit.Logging, []*point.Point{pt}, nil); err != nil {
		log.Errorf("Feed: %s", err.Error())
	}
}
