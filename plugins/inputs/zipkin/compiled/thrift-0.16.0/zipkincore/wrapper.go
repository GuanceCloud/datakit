// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package zipkincore provide Zipkin model definitions.
package zipkincore

import "net"

type EndpointJSONApater struct {
	Ipv4        net.IP `json:"ipv4"`
	Port        int16  `json:"port"`
	ServiceName string `json:"service_name"`
	Ipv6        net.IP `json:"ipv6,omitempty"`
}

type AnnotationJSONApater struct {
	Timestamp uint64              `json:"timestamp"`
	Value     string              `json:"value"`
	Host      *EndpointJSONApater `json:"endpoint,omitempty"`
}

type BinaryAnnotationJSONApater struct {
	Key            string              `json:"key"`
	Value          []byte              `json:"value"`
	AnnotationType AnnotationType      `json:"annotation_type"`
	Host           *EndpointJSONApater `json:"endpoint,omitempty"`
}

type SpanJSONApater struct {
	TraceID           uint64                       `json:"traceId"`
	Name              string                       `json:"name"`
	ID                uint64                       `json:"id"`
	ParentID          uint64                       `json:"parentId,omitempty"`
	Annotations       []AnnotationJSONApater       `json:"annotations"`
	BinaryAnnotations []BinaryAnnotationJSONApater `json:"binaryAnnotations"`
	Debug             bool                         `json:"debug,omitempty"`
	Timestamp         uint64                       `json:"timestamp,omitempty"`
	Duration          uint64                       `json:"duration,omitempty"`
	TraceIDHigh       uint64                       `json:"traceIdHigh,omitempty"`
}
