// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package otlp

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	ContentTypeJSON        = "application/json"
	ContentTypeProtobuf    = "application/x-protobuf"
	ContentTypeAltProtobuf = "application/protobuf"
)

var ErrUnsupportedContentType = errors.New("unsupported OTLP content type")

// NormalizeContentType strips parameters and normalizes a content-type value.
func NormalizeContentType(contentType string) string {
	if contentType == "" {
		return ""
	}

	parts := strings.Split(contentType, ";")

	return strings.ToLower(strings.TrimSpace(parts[0]))
}

// Unmarshal decodes an OTLP payload into msg according to contentType.
func Unmarshal(body []byte, contentType string, msg proto.Message) error {
	switch NormalizeContentType(contentType) {
	case ContentTypeProtobuf, ContentTypeAltProtobuf:
		return proto.Unmarshal(body, msg)
	case ContentTypeJSON:
		return protojson.Unmarshal(body, msg)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedContentType, contentType)
	}
}
