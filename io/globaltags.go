// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

var (
	globalHostTags = map[string]string{}
	globalEnvTags  = map[string]string{}
)

func SetGlobalHostTags(k, v string) {
	globalHostTags[k] = v
}

func SetGlobalEnvTags(k, v string) {
	globalEnvTags[k] = v
}

func GlobalHostTags() map[string]string {
	return globalHostTags
}

func GlobalEnvTags() map[string]string {
	return globalEnvTags
}

func ClearGlobalTags() {
	globalHostTags = map[string]string{}
	globalEnvTags = map[string]string{}
}
