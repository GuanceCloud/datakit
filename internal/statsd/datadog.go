// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

// this is adapted from datadog's apache licensed version at
// https://github.com/DataDog/datadog-agent/blob/fcfc74f106ab1bd6991dfc6a7061c558d934158a/pkg/dogstatsd/parser.go#L173

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

const (
	priorityNormal              = "normal"
	eventInfo                   = "info"
	eventMeasurementName        = "dogstatsd_event"
	serviceCheckMeasurementName = "dogstatsd_service_check"
)

var uncommenter = strings.NewReplacer("\\n", "\n")

func (col *Collector) parseServiceCheck(now time.Time, message string) error {
	// syntax: _sc|<CHECK_NAME>|<STATUS>|d:<TIMESTAMP>|m:<MESSAGE>|#<TAGS>

	arr := strings.Split(message[4:] /* skip `_sc|` */, "|")

	if len(arr) < 2 {
		col.opts.l.Warnf("invalid service check message format: %s", message)
		return fmt.Errorf("invalid message format")
	}

	var (
		kvs       point.KVs
		checkName = arr[0]
		status    = arr[1]
		ts        = now.UnixNano()
	)

	kvs = kvs.AddTag("check_name", checkName)

	switch status {
	case "0":
		kvs = kvs.Add("status", "ok")
	case "1":
		kvs = kvs.Add("status", "warn")
	case "2":
		kvs = kvs.Add("status", "critical")
	case "3":
		kvs = kvs.Add("status", "unknown")
	}

	for _, elem := range arr[2:] {
		switch {
		case strings.HasPrefix(elem, "d:"):
			if x, err := strconv.ParseInt(elem[2:], 10, 64); err == nil {
				ts = time.Unix(x, 0).UnixNano() // d: unit is second
			} else {
				col.opts.l.Warn("invalid timestamp: %q", elem)
			}
		case strings.HasPrefix(elem, "m:"):
			kvs = kvs.Add("message", elem[2:])
		case strings.HasPrefix(elem, "#"):
			tags := map[string]string{}
			col.parseDataDogTags(tags, elem[1:])
			for k, v := range tags {
				kvs = kvs.AddTag(k, v)
			}
		}
	}

	col.loggingPts = append(col.loggingPts, point.NewPoint(serviceCheckMeasurementName, kvs, point.WithTimestamp(ts)))

	return nil
}

// parseEventMessage parse datadog event payload.
//
// example payload:
//  _e{title.length,text.length}:title|text|d:date_happened|p:priority|h:hostname|t:alert_type|s:source_type_nam|#tag1,tag2

func (col *Collector) parseEventMessage(now time.Time, message string, defaultHostname string) error {
	var (
		arr = strings.SplitN(message, ":", 2)
		ts  = now.UnixNano()
	)

	if len(arr) < 2 || len(arr[0]) < 7 || len(arr[1]) < 3 {
		col.opts.l.Warnf("invalid event message format: %s", message)
		return fmt.Errorf("invalid message format")
	}

	header := arr[0]
	message = arr[1]

	// parse title and text len
	rawLen := strings.SplitN(header[3:], ",", 2)
	if len(rawLen) != 2 {
		col.opts.l.Warnf("invalid message format: %s, header should contains 2 length", message)
		return fmt.Errorf("invalid message format")
	}

	titleLen, err := strconv.ParseInt(rawLen[0], 10, 64)
	if err != nil {
		col.opts.l.Warnf("invalid message format: %s", message)
		return fmt.Errorf("invalid message format, could not parse title.length: '%s'", rawLen[0])
	}

	if len(rawLen[1]) < 1 {
		col.opts.l.Warnf("invalid message format: %s", message)
		return fmt.Errorf("invalid message format, could not parse text.length: '%s'", rawLen[0])
	}

	textLen, err := strconv.ParseInt(rawLen[1][:len(rawLen[1])-1], 10, 64)
	if err != nil {
		col.opts.l.Warnf("invalid message format: %s", message)
		return fmt.Errorf("invalid message format, could not parse text.length: '%s'", rawLen[0])
	}

	if titleLen+textLen+1 > int64(len(message)) {
		col.opts.l.Warnf("invalid message format: %s", message)
		return fmt.Errorf("invalid message format, title.length and text.length exceed total message length")
	}

	var kvs point.KVs

	rawTitle := message[:titleLen]
	rawText := message[titleLen+1 : titleLen+1+textLen]
	message = message[titleLen+1+textLen:]

	if len(rawTitle) == 0 || len(rawText) == 0 {
		col.opts.l.Warnf("invalid message format: %s", message)
		return fmt.Errorf("invalid event message format: empty 'title' or 'text' field")
	}

	kvs = kvs.Add("title", rawTitle).
		Add("message", uncommenter.Replace(rawText)).
		Add("status", eventInfo).
		AddTag("priority", priorityNormal)

	if defaultHostname != "" {
		kvs = kvs.Add("host", defaultHostname)
	}

	if len(message) < 2 {
		col.loggingPts = append(col.loggingPts, point.NewPoint(eventMeasurementName, kvs, point.WithTimestamp(ts)))
		return nil
	}

	rawMetadataFields := strings.Split(message[1:], "|")
	for i := range rawMetadataFields {
		if len(rawMetadataFields[i]) < 2 {
			col.opts.l.Warnf("invalid message format: %s, rawMetadataFields[%d]: %+#v", message, i, rawMetadataFields[i])
			return errors.New("too short metadata field")
		}
		switch rawMetadataFields[i][:2] {
		case "d:":
			if x, err := strconv.ParseInt(rawMetadataFields[i][2:], 10, 64); err == nil {
				ts = time.Unix(x, 0).UnixNano() // d: unit is second
			} else {
				col.opts.l.Warn("invalid timestamp: %q", rawMetadataFields[i][2:])
			}
		case "p:":
			switch rawMetadataFields[i][2:] {
			case priorityNormal: // default set
			default:
				kvs = kvs.SetTag("priority", rawMetadataFields[i][2:])
			}
		case "h:":
			kvs = kvs.SetTag("host", rawMetadataFields[i][2:])
		case "t:":
			switch rawMetadataFields[i][2:] {
			case eventInfo: // default set
			default:
				kvs = kvs.Set("status", rawMetadataFields[i][2:])
			}
		case "k:":
			kvs = kvs.SetTag("aggregation_key", rawMetadataFields[i][2:])
		case "s:":
			kvs = kvs.SetTag("source_type_name", rawMetadataFields[i][2:])
		default:
			if rawMetadataFields[i][0] == '#' {
				tags := map[string]string{}
				col.parseDataDogTags(tags, rawMetadataFields[i][1:])
				for k, v := range tags {
					kvs = kvs.AddTag(k, v)
				}
			} else {
				col.opts.l.Warnf("invalid message format: %s", message)
				return fmt.Errorf("unknown metadata type: '%s'", rawMetadataFields[i])
			}
		}
	}

	col.loggingPts = append(col.loggingPts, point.NewPoint(eventMeasurementName, kvs, point.WithTimestamp(ts)))

	return nil
}

func (col *Collector) parseDataDogTags(tags map[string]string, message string) {
	if len(message) == 0 {
		return
	}

	col.opts.l.Debugf("parse dd tags: %s", message)

	arr := strings.Split(message, ",")
	for _, elem := range arr {
		kv := strings.SplitN(elem, ":", 2)
		if len(kv) != 2 {
			tags[kv[0]] = "" // for tag that missing value, we default set to ""
		} else {
			tags[kv[0]] = kv[1]
		}
	}
}
