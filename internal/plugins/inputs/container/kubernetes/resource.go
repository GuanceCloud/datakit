// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
)

type resourceType struct {
	name       string
	namespaced bool
	nodeLocal  bool
}

type resourceConstructor func(k8sClient) resource

var resources = map[resourceType]resourceConstructor{}

type resource interface {
	getMetadata(ctx context.Context, ns, fieldSelector string) (metadata, error)
	hasNext() bool
}

type metadata interface {
	transformMetric() pointKVs
	transformObject() pointKVs
}

func registerResource(name string, namespaced, nodeLocal bool, rt resourceConstructor) {
	resources[resourceType{name, namespaced, nodeLocal}] = rt
}
