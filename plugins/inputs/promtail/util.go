// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promtail

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"strings"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/grafana/loki/pkg/util/unmarshal"
	unmarshal_legacy "github.com/grafana/loki/pkg/util/unmarshal/legacy"
	"github.com/prometheus/prometheus/pkg/labels"
	promql_parser "github.com/prometheus/prometheus/promql/parser"
)

var (
	contentType     = http.CanonicalHeaderKey("Content-Type")
	contentEnc      = http.CanonicalHeaderKey("Content-Encoding")
	applicationJSON = "application/json"
)

func (i *Input) parseRequest(r *http.Request) (*logproto.PushRequest, error) {
	var body io.Reader
	contentEncoding := r.Header.Get(contentEnc)
	switch contentEncoding {
	case "":
		body = r.Body
	case "snappy":
		body = r.Body
	case "gzip":
		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close() //nolint:errcheck
		body = gzipReader
	case "deflate":
		flateReader := flate.NewReader(r.Body)
		defer flateReader.Close() //nolint:errcheck
		body = flateReader
	default:
		return nil, fmt.Errorf("Content-Encoding %q is not supported", contentEncoding)
	}

	var req logproto.PushRequest
	contentType := r.Header.Get(contentType)
	contentType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}

	switch contentType {
	case applicationJSON:
		var err error
		if i.Legacy {
			err = unmarshal_legacy.DecodePushRequest(body, &req)
		} else {
			err = unmarshal.DecodePushRequest(body, &req)
		}
		if err != nil {
			return nil, err
		}
	default:
		// When no content-type header is set or when it is set to
		// `application/x-protobuf`: expect snappy compression.
		if err := util.ParseProtoReader(r.Context(), body, int(r.ContentLength), math.MaxInt32, &req, util.RawSnappy); err != nil {
			return nil, err
		}
	}
	return &req, nil
}

func getSource(req *http.Request) string {
	source := req.URL.Query().Get("source")
	if source != "" {
		return source
	}
	return "default"
}

func getPipelinePath(req *http.Request) string {
	return req.URL.Query().Get("pipeline")
}

func getCustomTags(req *http.Request) map[string]string {
	tagStr := req.URL.Query().Get("tags")
	return parseTagStr(tagStr)
}

func parseTagStr(tagStr string) map[string]string {
	tags := make(map[string]string)
	if !strings.Contains(tagStr, "=") {
		return tags
	}
	parts := strings.Split(tagStr, ",")
	for _, p := range parts {
		kv := strings.Split(p, "=")
		if len(kv) != 2 {
			l.Warnf("skip invalid custom tag: %s", p)
			continue
		}
		tags[kv[0]] = kv[1]
	}
	return tags
}

// ParseLabels parses labels from a string using logql parser.
func parseLabels(lbs string) (labels.Labels, error) {
	ls, err := promql_parser.ParseMetric(lbs)
	if err != nil {
		return nil, err
	}
	return ls, nil
}
