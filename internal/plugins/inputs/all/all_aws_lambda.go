// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build datakit_aws_lambda && with_inputs
// +build datakit_aws_lambda,with_inputs

// Package inputs wraps all inputs implements
package inputs

import (
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ddtrace"
	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/dk"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/opentelemetry"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/statsd"
)
