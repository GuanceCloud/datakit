// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import "time"

func (p *Point) SetName(name string) {
	p.pt.Name = name
}

func (p *Point) SetTime(t time.Time) {
	p.pt.Time = t.UnixNano()
}
