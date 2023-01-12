// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"sort"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

func (p *Point) SetTags(tags Tags) error {
	sort.Sort(tags)

	p.tags = tags

	if p.lpPoint != nil {
		fs, err := p.lpPoint.Fields()
		if err != nil {
			return err
		}

		newLPPt, err := influxdb.NewPoint(string(p.Name()), tags.InfluxTags(), fs, p.Time())
		if err != nil {
			return err
		}

		p.lpPoint = newLPPt
	}

	if p.pbPoint != nil {
		p.pbPoint.Tags = tags
	}

	return nil
}

func (p *Point) SetFields(fields Fields) error {
	sort.Sort(fields)

	p.fields = fields

	if p.lpPoint != nil {
		newLPPt, err := influxdb.NewPoint(string(p.Name()), p.Tags().InfluxTags(), fields.InfluxFields(), p.Time())
		if err != nil {
			return err
		}

		p.lpPoint = newLPPt
	}

	if p.pbPoint != nil {
		p.pbPoint.Fields = fields
	}
	return nil
}

// TODO: set-time/set-name
