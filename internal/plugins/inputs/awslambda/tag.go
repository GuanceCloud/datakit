// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

import (
	"os"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const (
	// EnvLambdaFunctionName 函数名称.
	EnvLambdaFunctionName = "AWS_LAMBDA_FUNCTION_NAME"
	// EnvLambdaFunctionVersion 函数版本.
	EnvLambdaFunctionVersion = "AWS_LAMBDA_FUNCTION_VERSION"
	// EnvAWSRegion 执行函数的 AWS 区域.
	EnvAWSRegion = "AWS_REGION"
	// EnvLambdaFunctionMemorySize 函数配置的内存大小.
	EnvLambdaFunctionMemorySize = "AWS_LAMBDA_FUNCTION_MEMORY_SIZE"
	// EnvLambdaInitializationType 函数的初始化类型.
	EnvLambdaInitializationType = "AWS_LAMBDA_INITIALIZATION_TYPE"
)

const (
	LambdaFunctionName       = "aws_lambda_function_name"
	LambdaFunctionVersion    = "aws_lambda_function_version"
	AWSRegion                = "aws_region"
	LambdaFunctionMemorySize = "aws_lambda_function_memory_size"
	LambdaInitializationType = "aws_lambda_initialization_type"
	AccountID                = "aws_account_id"
	AWSLogFrom               = "aws_log_from"
)

func (ipt *Input) initTags() {
	tags := make(map[string]string)

	envKeys := []string{
		EnvLambdaFunctionName,
		EnvLambdaFunctionVersion,
		EnvAWSRegion,
		EnvLambdaFunctionMemorySize,
		EnvLambdaInitializationType,
	}
	tagKeys := []string{
		LambdaFunctionName,
		LambdaFunctionVersion,
		AWSRegion,
		LambdaFunctionMemorySize,
		LambdaInitializationType,
	}
	for i, key := range envKeys {
		if value := os.Getenv(key); value != "" {
			tags[tagKeys[i]] = value
		}
	}

	if tags[LambdaFunctionVersion] == "$LATEST" {
		delete(tags, LambdaFunctionVersion)
	}

	datakit.SetGlobalHostTagsByMap(tags)
}
