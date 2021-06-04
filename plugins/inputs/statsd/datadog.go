package statsd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	eventInfo    = "info"
	eventWarning = "warning"
	eventError   = "error"
	eventSuccess = "success"

	priorityLow    = "low"
	priorityNormal = "normal"
)

var (
	uncommenter = strings.NewReplacer("\\n", "\n")
)

func (p *parser) parseEventMsg(now time.Time, message, host string) error {
	// _e{title.length,text.length}:title|text
	//  [
	//   |d:date_happened
	//   |p:priority
	//   |h:hostname
	//   |t:alert_type
	//   |s:source_type_nam
	//   |#tag1,tag2
	//  ]
	//
	//
	// tag is key:value
	messageRaw := strings.SplitN(message, ":", 2)
	if len(messageRaw) < 2 || len(messageRaw[0]) < 7 || len(messageRaw[1]) < 3 {
		return fmt.Errorf("invalid message format")
	}
	header := messageRaw[0]
	message = messageRaw[1]

	rawLen := strings.SplitN(header[3:], ",", 2)
	if len(rawLen) != 2 {
		return fmt.Errorf("invalid message format")
	}

	titleLen, err := strconv.ParseInt(rawLen[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid message format, could not parse title.length: '%s'", rawLen[0])
	}
	if len(rawLen[1]) < 1 {
		return fmt.Errorf("invalid message format, could not parse text.length: '%s'", rawLen[0])
	}
	textLen, err := strconv.ParseInt(rawLen[1][:len(rawLen[1])-1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid message format, could not parse text.length: '%s'", rawLen[0])
	}
	if titleLen+textLen+1 > int64(len(message)) {
		return fmt.Errorf("invalid message format, title.length and text.length exceed total message length")
	}

	rawTitle := message[:titleLen]
	rawText := message[titleLen+1 : titleLen+1+textLen]
	message = message[titleLen+1+textLen:]

	if len(rawTitle) == 0 || len(rawText) == 0 {
		return fmt.Errorf("invalid event message format: empty 'title' or 'text' field")
	}

	m := &metric{
		name: rawTitle,
		fields: map[string]interface{}{
			"alert_type": eventInfo, // default event type
			"text":       uncommenter.Replace(rawText),
			"priority":   priorityNormal,
			"ts":         now,
		},
		tags: map[string]string{},
	}

	if host != "" {
		m.tags["source"] = host
	}

	if len(message) < 2 {
		p.cache = append(p.cache, m)
		return nil
	}

	rawMetadataFields := strings.Split(message[1:], "|")
	for i := range rawMetadataFields {
		if len(rawMetadataFields[i]) < 2 {
			return errors.New("too short metadata field")
		}
		switch rawMetadataFields[i][:2] {
		case "d:":
			ts, err := strconv.ParseInt(rawMetadataFields[i][2:], 10, 64)
			if err != nil {
				continue
			}
			m.fields["ts"] = ts
		case "p:":
			switch rawMetadataFields[i][2:] {
			case priorityLow:
				m.fields["priority"] = priorityLow
			case priorityNormal: // we already used this as a default
			default:
				continue
			}
		case "h:":
			m.tags["source"] = rawMetadataFields[i][2:]
		case "t:":
			switch rawMetadataFields[i][2:] {
			case eventError, eventWarning, eventSuccess, eventInfo:
				m.fields["alert_type"] = rawMetadataFields[i][2:] // already set for info
			default:
				continue
			}
		case "k:":
			m.tags["aggregation_key"] = rawMetadataFields[i][2:]
		case "s:":
			m.fields["source_type_name"] = rawMetadataFields[i][2:]
		default:
			if rawMetadataFields[i][0] == '#' {
				parseDataDogTags(m.tags, rawMetadataFields[i][1:])
			} else {
				return fmt.Errorf("unknown metadata type: '%s'", rawMetadataFields[i])
			}
		}
	}
	// Use source tag because host is reserved tag key in Telegraf.
	// In datadog the host tag and `h:` are interchangable, so we have to chech for the host tag.
	if host, ok := m.tags["host"]; ok {
		delete(m.tags, "host")
		m.tags["source"] = host
	}

	p.cache = append(p.cache, m)

	return nil
}

func parseDataDogTags(tags map[string]string, message string) {
	if len(message) == 0 {
		return
	}

	start, i := 0, 0
	var k string
	var inVal bool // check if we are parsing the value part of the tag
	for i = range message {
		if message[i] == ',' {
			if k == "" {
				k = message[start:i]
				tags[k] = "true" // this is because influx doesn't support empty tags
				start = i + 1
				continue
			}
			v := message[start:i]
			if v == "" {
				v = "true"
			}
			tags[k] = v
			start = i + 1
			k, inVal = "", false // reset state vars
		} else if message[i] == ':' && !inVal {
			k = message[start:i]
			start = i + 1
			inVal = true
		}
	}
	if k == "" && start < i+1 {
		tags[message[start:i+1]] = "true"
	}
	// grab the last value
	if k != "" {
		if start < i+1 {
			tags[k] = message[start : i+1]
			return
		}
		tags[k] = "true"
	}

	return
}
