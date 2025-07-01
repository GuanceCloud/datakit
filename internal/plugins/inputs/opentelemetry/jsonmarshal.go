// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"encoding/json"

	jsoniter "github.com/json-iterator/go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type jsonMarshaler interface {
	Marshal(x proto.Message) ([]byte, error)
}

type protojsonMarshaler struct{}

func (j *protojsonMarshaler) Marshal(x proto.Message) ([]byte, error) {
	return protojson.Marshal(x)
}

type gojsonMarshaler struct{}

func (j *gojsonMarshaler) Marshal(x proto.Message) ([]byte, error) {
	return json.Marshal(x)
}

func initJSONIter() jsoniter.API {
	customExtension := jsoniter.DecoderExtension{}

	// 注册所有OTel关键类型的序列化优化
	json := jsoniter.Config{
		EscapeHTML:                    false,
		TagKey:                        "json",
		OnlyTaggedField:               false,
		SortMapKeys:                   false,
		IndentionStep:                 0,
		ValidateJsonRawMessage:        true,
		ObjectFieldMustBeSimpleString: true,
	}.Froze()

	json.RegisterExtension(customExtension)
	return json
}

var otelJSON = initJSONIter()

type jsoniterMarshaler struct{}

func (m *jsoniterMarshaler) Marshal(x proto.Message) ([]byte, error) {
	return otelJSON.Marshal(x)
}
