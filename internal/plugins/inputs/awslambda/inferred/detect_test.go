// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package inferred

import "testing"

func TestDetectKnownTriggers(t *testing.T) {
	tests := []struct {
		name      string
		payload   string
		operation string
	}{
		{"api-gateway-rest", `{"httpMethod":"GET","path":"/orders","requestContext":{"requestId":"1"}}`, "aws.apigateway"},
		{"lambda-function-url", `{"rawPath":"/orders","requestContext":{"domainName":"abc.lambda-url.us-east-1.on.aws","http":{"method":"GET"}}}`, "aws.lambda.url"},
		{"api-gateway-http", `{"rawPath":"/orders","requestContext":{"http":{"method":"GET"}}}`, "aws.apigateway.http"},
		{"alb", `{"httpMethod":"GET","path":"/alb","requestContext":{"elb":{"targetGroupArn":"arn"}}}`, "aws.alb"},
		{"sqs", `{"Records":[{"eventSource":"aws:sqs","eventSourceARN":"arn:aws:sqs:queue"}]}`, "aws.sqs"},
		{"sns", `{"Records":[{"EventSource":"aws:sns","Sns":{"TopicArn":"arn:aws:sns:topic"}}]}`, "aws.sns"},
		{"eventbridge", `{"source":"custom.app","detail-type":"app.event"}`, "aws.eventbridge"},
		{"s3", `{"Records":[{"eventSource":"aws:s3","s3":{"bucket":{"name":"bucket-a"}}}]}`, "aws.s3"},
		{"msk", `{"eventSource":"aws:kafka","records":{"topic-0":[{"topic":"topic"}]}}`, "aws.msk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := Detect([]byte(tt.payload))
			if info == nil {
				t.Fatalf("expected inferred span info")
			}
			if info.Operation != tt.operation {
				t.Fatalf("got %q want %q", info.Operation, tt.operation)
			}
		})
	}
}
