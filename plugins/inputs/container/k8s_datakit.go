// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	kubev1guancebeta1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/typed/guance/v1beta1"
)

//nolint
type datakitCRD struct {
	client k8sClientX
	items  []kubev1guancebeta1.DataKit
}

//nolint
func newDatakitCRD(client k8sClientX) *datakitCRD {
	return &datakitCRD{
		client: client,
	}
}
