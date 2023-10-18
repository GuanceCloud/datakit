// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package newrelic handle New Relic APM traces.
package newrelic

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

// type meta struct {
// 	sql      string `json:"sql"`
// 	host     string `json:"host"`
// 	info     string `json:"port_path_or_id"`
// 	database string `json:"database_name"`
// 	duration int64  `json:"exclusive_duration_millis"`
// }

type meta map[string]interface{}

// for keys: sql, host, port_path_or_id, database_name, exclusive_duration_millis.
func (x meta) stringValue(key string) string {
	if s, ok := x[key]; ok {
		if v, ok := s.(string); ok {
			return v
		}
	}

	return ""
}

func (x meta) float64Value(key string) float64 { // nolint:unused
	if f, ok := x[key]; ok {
		if v, ok := f.(float64); ok {
			return v
		}
	}

	return 0
}

type segment []interface{}

// micro sec.
func (x segment) startElapsed() int64 {
	if len(x) > 0 {
		if start, ok := x[0].(float64); ok {
			return int64(start * 1000)
		}
	}

	return 0
}

// micro sec.
func (x segment) endElapsed() int64 {
	if len(x) > 1 {
		if end, ok := x[1].(float64); ok {
			return int64(end * 1000)
		}
	}

	return 0
}

func (x segment) resource() string {
	if len(x) > 2 {
		if res, ok := x[2].(string); ok {
			return res
		}
	}

	return ""
}

func (x segment) meta() meta {
	if len(x) > 3 {
		if meta, ok := x[3].(map[string]interface{}); ok {
			return meta
		}
	}

	return meta{}
}

func (x segment) children() []segment {
	if len(x) > 4 {
		if ss, ok := x[4].([]interface{}); ok {
			var segs []segment
			for _, v := range ss {
				if s, ok := v.([]interface{}); ok {
					segs = append(segs, s)
				}
			}

			return segs
		}
	}

	return []segment{}
}

func (x segment) class() string {
	if len(x) > 5 {
		if c, ok := x[5].(string); ok {
			return c
		}
	}

	return ""
}

func (x segment) method() string {
	if len(x) > 6 {
		if m, ok := x[6].(string); ok {
			return m
		}
	}

	return ""
}

func (x segment) parameters() string { // nolint:unused
	if len(x) > 7 {
		if m, ok := x[7].(string); ok {
			return m
		}
	}

	return ""
}

type transaction []interface{}

func (x transaction) agentRunID() string {
	if len(x) > 0 {
		return x[0].(string)
	}

	return ""
}

// micro sec.
func (x transaction) start() int64 {
	if len(x) > 1 {
		if ss, ok := x[1].([]interface{}); ok {
			if len(ss) > 0 {
				if ss, ok = ss[0].([]interface{}); ok {
					if len(ss) > 0 {
						if v, ok := ss[0].(float64); ok {
							return int64(v * 1000)
						}
					}
				}
			}
		}
	}

	return 0
}

// micro sec.
func (x transaction) duration() int64 {
	if len(x) > 1 {
		if ss, ok := x[1].([]interface{}); ok {
			if len(ss) > 0 {
				if ss, ok = ss[0].([]interface{}); ok {
					if len(ss) > 1 {
						if v, ok := ss[1].(float64); ok {
							return int64(v * 1000)
						}
					}
				}
			}
		}
	}

	return 0
}

func (x transaction) module() string { // nolint: unused
	if len(x) > 1 {
		if ss, ok := x[1].([]interface{}); ok {
			if len(ss) > 0 {
				if ss, ok = ss[0].([]interface{}); ok {
					if len(ss) > 2 {
						if v, ok := ss[2].(string); ok {
							return v
						}
					}
				}
			}
		}
	}

	return ""
}

func (x transaction) url() string {
	if len(x) > 1 {
		if ss, ok := x[1].([]interface{}); ok {
			if len(ss) > 0 {
				if ss, ok = ss[0].([]interface{}); ok {
					if len(ss) > 3 {
						if v, ok := ss[3].(string); ok {
							return v
						}
					}
				}
			}
		}
	}

	return ""
}

func (x transaction) root() segment {
	if len(x) > 1 {
		if ss, ok := x[1].([]interface{}); ok {
			if len(ss) > 0 {
				if ss, ok = ss[0].([]interface{}); ok {
					if len(ss) > 4 {
						if ss, ok = ss[4].([]interface{}); ok {
							if len(ss) > 3 {
								if ss, ok = ss[3].([]interface{}); ok {
									return ss
								}
							}
						}
					}
				}
			}
		}
	}

	return segment{}
}

func (x transaction) id() string {
	if len(x) > 1 {
		if ss, ok := x[1].([]interface{}); ok {
			if len(ss) > 0 {
				if ss, ok = ss[0].([]interface{}); ok {
					if len(ss) > 5 {
						if v, ok := ss[5].(string); ok {
							return v
						}
					}
				}
			}
		}
	}

	return ""
}

func (x transaction) attributes() map[string]interface{} { // nolint:unused
	if len(x) > 1 {
		if ss, ok := x[1].([]interface{}); ok {
			if len(ss) > 0 {
				if ss, ok = ss[0].([]interface{}); ok {
					if len(ss) > 4 {
						if ss, ok = ss[4].([]interface{}); ok {
							if len(ss) > 4 {
								if v, ok := ss[4].(map[string]interface{}); ok {
									return v
								}
							}
						}
					}
				}
			}
		}
	}

	return map[string]interface{}{}
}

func randHexID(l int) string {
	buf := make([]byte, l)
	rand.Read(buf) // nolint:errcheck,gosec

	return strings.ToUpper((hex.EncodeToString(buf)))
}

func makeRootSpan(idLength int, service string, transaction *transaction) *itrace.DatakitSpan {
	span := &itrace.DatakitSpan{
		TraceID:    transaction.id(),
		ParentID:   "0",
		SpanID:     randHexID(idLength),
		Service:    service,
		Resource:   transaction.url(),
		Operation:  transaction.url(),
		Source:     inputName,
		SpanType:   itrace.SPAN_TYPE_ENTRY,
		SourceType: itrace.SPAN_SOURCE_WEB,
		Start:      transaction.start() * int64(time.Microsecond),
		Duration:   transaction.duration() * int64(time.Microsecond),
		Status:     itrace.STATUS_OK,
	}

	if uri, err := url.ParseRequestURI(transaction.url()); err == nil {
		span.Tags = map[string]string{itrace.TAG_HTTP_URL: uri.String()}
	}
	if len(tags) != 0 {
		if span.Tags == nil {
			span.Tags = make(map[string]string)
		}
		for k, v := range tags {
			span.Tags[k] = v
		}
	}

	if buf, err := json.Marshal(transaction.root()); err == nil {
		span.Content = string(buf)
	} else {
		log.Debug(err.Error())
	}

	return span
}

func makeChildrenSpan(service string, rootStart int64, idLength int, traceID, parentID string, children []segment, out *itrace.DatakitTrace) {
	for _, child := range children {
		span := &itrace.DatakitSpan{
			TraceID:    traceID,
			ParentID:   parentID,
			SpanID:     randHexID(idLength),
			Service:    service,
			Resource:   child.resource(),
			Operation:  fmt.Sprintf("%s:%s", child.class(), child.method()),
			Source:     inputName,
			SpanType:   itrace.SPAN_TYPE_LOCAL,
			SourceType: itrace.SPAN_SOURCE_WEB,
			Start:      (rootStart + child.startElapsed()) * int64(time.Microsecond),
			Duration:   (child.endElapsed() - child.startElapsed()) * int64(time.Microsecond),
			Status:     itrace.STATUS_OK,
		}
		if child.method() == "InvokeService" {
			if len(child.children()) != 0 {
				if uri, err := url.Parse(child.children()[0].meta().stringValue("uri")); err == nil {
					service = uri.Host
				}
			}
		}
		if sql := child.meta().stringValue("sql"); sql != "" {
			span.Resource = sql
		}
		if buf, err := json.Marshal(child); err == nil {
			span.Content = string(buf)
		} else {
			log.Debug(err.Error())
		}
		*out = append(*out, span)

		if len(child.children()) != 0 {
			makeChildrenSpan(service, rootStart, idLength, traceID, span.SpanID, child.children(), out)
		}
	}
}

func transformToDkTrace(transaction *transaction) itrace.DatakitTrace {
	var (
		l       = len(transaction.id())
		id      = transaction.agentRunID()
		service string
		ok      bool
	)
	if id != "" {
		if service, ok = compound[cmdConnect].(*connect).findAppNameBy(id); !ok {
			service = id
		}
	} else {
		service = transaction.url()
	}

	root := makeRootSpan(l, service, transaction)
	trace := &itrace.DatakitTrace{root}
	makeChildrenSpan(service, transaction.start(), len(root.TraceID), root.TraceID, root.SpanID, transaction.root().children(), trace)

	return *trace
}
