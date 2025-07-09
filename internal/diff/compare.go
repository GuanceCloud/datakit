// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diff

import (
	"reflect"

	"sigs.k8s.io/yaml"
)

const defaultContextLines = 4

func Compare(oldVal, newVal interface{}) (equal bool, difftext string) {
	if reflect.DeepEqual(oldVal, newVal) {
		return true, ""
	}

	oldText, _ := yaml.Marshal(oldVal)
	newText, _ := yaml.Marshal(newVal)

	return false, LineDiffWithContextLines(string(oldText), string(newText), defaultContextLines)
}
