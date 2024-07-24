// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package extension

import (
	"fmt"
	"os"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/consts"
)

var l = logger.DefaultSLogger("awslambda")

func GetAwsLambdaRuntimeAPI() string {
	return os.Getenv("AWS_LAMBDA_RUNTIME_API")
}

func getBaseURL(awsLambdaRuntimeAPI string) string {
	baseURL := fmt.Sprintf("http://%s/%s", awsLambdaRuntimeAPI, consts.ExtensionRoute)
	return baseURL
}

func SetLogger(log *logger.Logger) {
	l = log
}
