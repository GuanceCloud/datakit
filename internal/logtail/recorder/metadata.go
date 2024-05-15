// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package recorder

import "fmt"

type MetaData struct {
	Source string `json:"source"`
	Offset int64  `json:"offset"`
}

func (m *MetaData) String() string {
	return fmt.Sprintf("source: %s, offset: %d", m.Source, m.Offset)
}

func (m *MetaData) DeepCopy() MetaData {
	return MetaData{
		Source: m.Source,
		Offset: m.Offset,
	}
}
