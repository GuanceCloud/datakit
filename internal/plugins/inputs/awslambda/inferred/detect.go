// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

// Package inferred provides functionality to detect AWS Lambda event triggers from payloads.
package inferred

import (
	"encoding/json"
	"strings"
)

type Info struct {
	Operation string
	Resource  string
	Service   string
	Tags      map[string]string
}

func Detect(payload []byte) *Info {
	if len(payload) == 0 {
		return nil
	}

	var event map[string]interface{}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil
	}

	switch {
	case hasKey(event, "requestContext") && nestedString(event, "requestContext", "elb", "targetGroupArn") != "":
		return &Info{
			Operation: "aws.alb",
			Resource:  stringValue(event, "path", "/"),
			Service:   "aws-alb",
			Tags:      map[string]string{"trigger": "alb"},
		}
	case hasKey(event, "httpMethod") && hasKey(event, "requestContext"):
		return &Info{
			Operation: "aws.apigateway",
			Resource:  stringValue(event, "path", "/"),
			Service:   "aws-apigateway",
			Tags:      map[string]string{"trigger": "api-gateway-rest"},
		}
	case strings.Contains(stringValue(event, "requestContext.domainName", ""), "lambda-url"):
		return &Info{
			Operation: "aws.lambda.url",
			Resource:  stringValue(event, "rawPath", "/"),
			Service:   "aws-lambda-url",
			Tags:      map[string]string{"trigger": "lambda-function-url"},
		}
	case nestedString(event, "requestContext", "http", "method") != "":
		return &Info{
			Operation: "aws.apigateway.http",
			Resource:  stringValue(event, "rawPath", "/"),
			Service:   "aws-apigateway",
			Tags:      map[string]string{"trigger": "api-gateway-http"},
		}
	case firstRecordSource(event) == "aws:sqs":
		return &Info{
			Operation: "aws.sqs",
			Resource:  firstRecordField(event, "eventSourceARN"),
			Service:   "aws-sqs",
			Tags:      map[string]string{"trigger": "sqs"},
		}
	case firstRecordSource(event) == "aws:sns":
		return &Info{
			Operation: "aws.sns",
			Resource:  firstRecordField(event, "Sns.TopicArn"),
			Service:   "aws-sns",
			Tags:      map[string]string{"trigger": "sns"},
		}
	case firstRecordSource(event) == "aws:s3":
		return &Info{
			Operation: "aws.s3",
			Resource:  firstRecordField(event, "s3.bucket.name"),
			Service:   "aws-s3",
			Tags:      map[string]string{"trigger": "s3"},
		}
	case firstRecordSource(event) == "aws:kinesis":
		return &Info{
			Operation: "aws.kinesis",
			Resource:  firstRecordField(event, "eventSourceARN"),
			Service:   "aws-kinesis",
			Tags:      map[string]string{"trigger": "kinesis"},
		}
	case firstRecordSource(event) == "aws:dynamodb":
		return &Info{
			Operation: "aws.dynamodb",
			Resource:  firstRecordField(event, "eventSourceARN"),
			Service:   "aws-dynamodb",
			Tags:      map[string]string{"trigger": "dynamodb"},
		}
	case firstRecordSource(event) == "aws:msk":
		return &Info{
			Operation: "aws.msk",
			Resource:  firstRecordField(event, "eventSourceArn"),
			Service:   "aws-msk",
			Tags:      map[string]string{"trigger": "msk"},
		}
	case stringValue(event, "eventSource", "") == "aws:kafka" || stringValue(event, "eventSource", "") == "aws:msk":
		return &Info{
			Operation: "aws.msk",
			Resource:  firstKafkaRecordKey(event),
			Service:   "aws-msk",
			Tags:      map[string]string{"trigger": "msk"},
		}
	case hasKey(event, "detail-type") && hasKey(event, "source"):
		return &Info{
			Operation: "aws.eventbridge",
			Resource:  stringValue(event, "source", "eventbridge"),
			Service:   "aws-eventbridge",
			Tags:      map[string]string{"trigger": "eventbridge"},
		}
	case hasKey(event, "Execution") || hasKey(event, "StateMachine"):
		return &Info{
			Operation: "aws.step_functions",
			Resource:  nestedString(event, "StateMachine", "Name"),
			Service:   "aws-step-functions",
			Tags:      map[string]string{"trigger": "step-functions"},
		}
	default:
		return nil
	}
}

func hasKey(m map[string]interface{}, key string) bool {
	if strings.Contains(key, ".") {
		return nestedString(m, strings.Split(key, ".")...) != ""
	}
	_, ok := m[key]
	return ok
}

func stringValue(m map[string]interface{}, key, fallback string) string {
	if strings.Contains(key, ".") {
		if value := nestedString(m, strings.Split(key, ".")...); value != "" {
			return value
		}
		return fallback
	}
	if value, ok := m[key].(string); ok && value != "" {
		return value
	}
	return fallback
}

func nestedString(v interface{}, path ...string) string {
	current := v
	for _, step := range path {
		m, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current, ok = m[step]
		if !ok {
			return ""
		}
	}
	s, _ := current.(string)
	return s
}

func firstRecordSource(m map[string]interface{}) string {
	records, ok := m["Records"].([]interface{})
	if !ok || len(records) == 0 {
		return ""
	}
	record, ok := records[0].(map[string]interface{})
	if !ok {
		return ""
	}
	if source, ok := record["eventSource"].(string); ok {
		return source
	}
	if source, ok := record["EventSource"].(string); ok {
		return source
	}
	return ""
}

func firstRecordField(m map[string]interface{}, path string) string {
	records, ok := m["Records"].([]interface{})
	if !ok || len(records) == 0 {
		return ""
	}
	record, ok := records[0].(map[string]interface{})
	if !ok {
		return ""
	}
	return nestedString(record, strings.Split(path, ".")...)
}

func firstKafkaRecordKey(m map[string]interface{}) string {
	records, ok := m["records"].(map[string]interface{})
	if !ok {
		return ""
	}
	for key := range records {
		return key
	}
	return ""
}
