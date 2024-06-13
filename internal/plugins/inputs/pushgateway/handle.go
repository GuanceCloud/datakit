// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pushgateway

import (
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/golang/snappy"
	"github.com/prometheus/common/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
)

const base64Suffix = "@base64"

type processor func(io.Reader, map[string]string) error

func pushHandle(
	jobBase64Encoded bool,
	text processor,
	protobuf processor,
) func(http.ResponseWriter, *http.Request, ...interface{}) (interface{}, error) {
	handler := func(w http.ResponseWriter, r *http.Request, _ ...interface{}) (interface{}, error) {
		// parse job
		job := param(r.Context(), "job")
		if jobBase64Encoded {
			var err error
			if job, err = decodeBase64(job); err != nil {
				err = fmt.Errorf("invalid base64 encoding in job name %q: %w", job, err)
				return nil, uhttp.NewErr(err, http.StatusBadRequest)
			}
		}
		if job == "" {
			err := fmt.Errorf("job name is required")
			return nil, uhttp.NewErr(err, http.StatusBadRequest)
		}

		// parse labels
		labelsString := param(r.Context(), "labels")
		labels, err := splitLabels(labelsString)
		if err != nil {
			return nil, uhttp.NewErr(err, http.StatusBadRequest)
		}
		labels["job"] = job

		// default use text processor
		process := text
		// type is protobuf
		ctMediatype, ctParams, ctErr := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if ctErr == nil &&
			ctMediatype == "application/vnd.google.protobuf" &&
			ctParams["encoding"] == "delimited" &&
			ctParams["proto"] == "io.prometheus.client.MetricFamily" {
			process = protobuf
		}

		// decoding
		defer r.Body.Close() //nolint:errcheck
		switch contentEncoding := r.Header.Get("Content-Encoding"); strings.ToLower(contentEncoding) {
		case "gzip":
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				return nil, uhttp.NewErr(err, http.StatusBadRequest)
			}
			defer gr.Close() //nolint:errcheck
			r.Body = gr
		case "snappy":
			r.Body = io.NopCloser(snappy.NewReader(r.Body))
		default:
			// Do nothing
		}

		if err := process(r.Body, labels); err != nil {
			log.Debugf("failed to parse text for %s, err %s", r.RemoteAddr, err)
			return nil, uhttp.NewErr(err, http.StatusBadRequest)
		}

		return nil, nil
	}

	return handler
}

func param(ctx context.Context, p string) string {
	if v := ctx.Value(httpapi.Param(p)); v != nil {
		return v.(string)
	}
	return ""
}

// decodeBase64 decodes the provided string using the “Base 64 Encoding with URL
// and Filename Safe Alphabet” (RFC 4648). Padding characters (i.e. trailing
// '=') are ignored.
func decodeBase64(s string) (string, error) {
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
	return string(b), err
}

// splitLabels splits a labels string into a label map mapping names to values.
func splitLabels(labels string) (map[string]string, error) {
	result := map[string]string{}
	if len(labels) <= 1 {
		return result, nil
	}
	components := strings.Split(labels[1:], "/")
	if len(components)%2 != 0 {
		return nil, fmt.Errorf("odd number of components in label string %q", labels)
	}

	for i := 0; i < len(components)-1; i += 2 {
		name, value := components[i], components[i+1]
		trimmedName := strings.TrimSuffix(name, base64Suffix)
		if !model.LabelNameRE.MatchString(trimmedName) ||
			strings.HasPrefix(trimmedName, model.ReservedLabelPrefix) {
			return nil, fmt.Errorf("improper label name %q", trimmedName)
		}
		if name == trimmedName {
			result[name] = value
			continue
		}
		decodedValue, err := decodeBase64(value)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 encoding for label %s=%q: %w", trimmedName, value, err)
		}
		result[trimmedName] = decodedValue
	}
	return result, nil
}
