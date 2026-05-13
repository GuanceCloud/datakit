// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

// Package lambdatrace provides tracing functionality for AWS Lambda invocations.
package lambdatrace

import (
	"os"
	"strconv"
)

const (
	envDDTraceAWSServiceRepresentationEnabled = "DD_TRACE_AWS_SERVICE_REPRESENTATION_ENABLED"
	envLambdaFunctionName                     = "AWS_LAMBDA_FUNCTION_NAME"
	envLambdaFunctionVersion                  = "AWS_LAMBDA_FUNCTION_VERSION"
	defaultLambdaSpanService                  = "aws.lambda"
	defaultLambdaSpanResource                 = "aws.lambda"
)

func lambdaInvocationServiceAndResource() (service, resource string) {
	resource = lambdaCanonicalResourceName()
	service = resource

	if !traceAWSServiceRepresentationEnabled() {
		service = defaultLambdaSpanService
	}

	if service == "" {
		service = defaultLambdaSpanService
	}

	if resource == "" {
		resource = defaultLambdaSpanResource
	}

	return service, resource
}

func lambdaCanonicalResourceName() string {
	name := os.Getenv(envLambdaFunctionName)
	if name == "" {
		return defaultLambdaSpanResource
	}

	version := os.Getenv(envLambdaFunctionVersion)
	if version != "" && version != "$LATEST" {
		return name + ":" + version
	}

	return name
}

func traceAWSServiceRepresentationEnabled() bool {
	value, ok := os.LookupEnv(envDDTraceAWSServiceRepresentationEnabled)
	if !ok || value == "" {
		return true
	}

	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return true
	}

	return enabled
}
