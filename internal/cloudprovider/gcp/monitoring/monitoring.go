// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitoring

import (
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
)

const RuntimeName = runtime.CloudMonitoringRuntime

type Config = runtime.CloudMonitoringConfig

func NewRuntime(cfg Config, client *http.Client) (runtime.ContainerRuntime, error) {
	return runtime.NewCloudMonitoringRuntime(cfg, client)
}
