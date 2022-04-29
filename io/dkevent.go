package io

import (
	"context"
	"regexp"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type DKEvent struct {
	Status   string `json:"status"` // info/warning/error
	Message  string `json:"message"`
	Category string `josn:"category"`
	Logtype  string `josn:"Logtype"`
	feed     func(string, string, []*Point, *Option) error
}

func (e *DKEvent) Tags() map[string]string {
	tags := map[string]string{
		"status":   "info",
		"category": "default",
		"log_type": "",
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
	pt, err := NewPoint("datakit", e.Tags(), e.Fields(), &PointOption{
		Time:     time.Now(),
		Category: datakit.Logging,
	})
	if err != nil {
		log.Errorf("NewPoint: %s", err.Error())
		return
	}

	g := datakit.G("io")
	g.Go(func(ctx context.Context) error {
		if err := Feed("dkevent", datakit.Logging, []*Point{pt}, nil); err != nil {
			log.Errorf("Feed: %s", err.Error())
		}
		return nil
	})
}
