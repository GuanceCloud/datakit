// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import "strings"

type Encoding int

const (
	encProtobuf      = "protobuf"
	encProtobufAlias = "v2"
	encJSON          = "json"
	encPBJSON        = "pbjson"

	encLineprotocolAlias = "v1"
	encLineprotocol      = "line-protocol"

	contentTypeJSON      = "application/json"
	contentTypeProtobuf  = "application/protobuf; proto=com.guance.Point"
	contentTypePBJSON    = "application/pbjson; proto=com.guance.Point"
	contentTypeLineproto = "application/line-protocol"
)

const (
	LineProtocol Encoding = iota // encoding in InfluxDB line-protocol
	Protobuf                     // encoding in protobuf
	JSON                         // encoding int simple JSON
	PBJSON                       // encoding in protobuf structured JSON(with better field-type labeled)
)

// EncodingStr convert encoding-string in configure file to
// encoding enum.
//
// Here v1/v2 are alias for lineprotocol and protobuf, this makes
// people easy to switch between lineprotocol and protobuf. For
// json, you should not configure json encoding in production
// environments(json do not classify int and float).
func EncodingStr(s string) Encoding {
	switch strings.ToLower(s) {
	case encProtobuf, encProtobufAlias:
		return Protobuf
	case encJSON:
		return JSON
	case encPBJSON:
		return PBJSON
	case encLineprotocol,
		encLineprotocolAlias:
		return LineProtocol
	default:
		return LineProtocol
	}
}

// HTTPContentType detect HTTP body content encoding according to header Content-Type.
func HTTPContentType(ct string) Encoding {
	switch ct {
	case contentTypeJSON:
		return JSON
	case contentTypePBJSON:
		return PBJSON
	case contentTypeProtobuf:
		return Protobuf
	case contentTypeLineproto:
		return LineProtocol
	default: // default use line-protocol to be compatible with lagacy code
		return LineProtocol
	}
}

// HTTPContentType get correct HTTP Content-Type value on different body encoding.
func (e Encoding) HTTPContentType() string {
	switch e {
	case JSON:
		return contentTypeJSON
	case PBJSON:
		return contentTypePBJSON
	case Protobuf:
		return contentTypeProtobuf
	case LineProtocol:
		return contentTypeLineproto
	default: // default use line-protocol to be compatible with lagacy code
		return contentTypeLineproto
	}
}

func (e Encoding) String() string {
	switch e {
	case JSON:
		return encJSON
	case PBJSON:
		return encPBJSON
	case Protobuf:
		return encProtobuf
	case LineProtocol:
		return encLineprotocol
	default: // default use line-protocol to be compatible with lagacy code
		return encLineprotocol
	}
}
