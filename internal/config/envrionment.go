// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import "os"

func IsECSFargate() bool {
	return os.Getenv("ENV_ECS_FARGATE") == "on"
}

func IsDockerRuntime() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func IsKubernetes() bool {
	return os.Getenv("KUBERNETES_SERVICE_PORT") != ""
}

func ECSFargateBaseURIV4() string {
	return os.Getenv("ECS_CONTAINER_METADATA_URI_V4")
}

func ECSFargateBaseURIV3() string {
	return os.Getenv("ECS_CONTAINER_METADATA_URI")
}
