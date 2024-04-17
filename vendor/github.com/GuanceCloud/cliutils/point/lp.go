// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/influxdata/influxdb1-client/models"
)

type Precision int

const (
	PrecNS Precision = iota // nano-second
	PrecUS                  // micro-second
	PrecMS                  // milli-second
	PrecS                   // second
	PrecM                   // minute
	PrecH                   // hour
	PrecD                   // day
	PrecW                   // week
)

func (p Precision) String() string {
	switch p {
	case PrecNS:
		return "n"
	case PrecUS:
		return "u"
	case PrecMS:
		return "ms"
	case PrecS:
		return "s"
	case PrecM:
		return "m"
	case PrecH:
		return "h"
	case PrecD:
		return "d"
	case PrecW:
		return "w"
	default:
		return "unknown"
	}
}

func PrecStr(s string) Precision {
	switch s {
	case "n":
		return PrecNS
	case "u":
		return PrecUS
	case "ms":
		return PrecMS
	case "s":
		return PrecS
	case "m":
		return PrecM
	case "h":
		return PrecH
	default:
		return PrecNS
	}
}

var sepRe = regexp.MustCompile(": ")

// simplifyLPError used to simplify original line-protocol parse error, the
// parse error will print origin data payload, for large payload, it's hard
// to read the error message.
func simplifyLPError(err error) error {
	parts := sepRe.Split(err.Error(), -1)
	if len(parts) != 2 { // return origin error
		return err
	}

	return fmt.Errorf("lineproto parse error: %s", parts[1])
}

// parseLPPoints parse line-protocol payload to Point.
func parseLPPoints(data []byte, c *cfg) ([]*Point, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	if c == nil {
		c = GetCfg()
		defer PutCfg(c)
	}

	ptTime := c.t
	if c.t.IsZero() {
		ptTime = time.Now()
	}

	lppts, err := models.ParsePointsWithPrecision(data, ptTime, c.precision.String())
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidLineProtocol, err)
	}

	res := []*Point{}
	chk := checker{cfg: c}

	for _, x := range lppts {
		if x == nil {
			return nil, fmt.Errorf("line point is empty")
		}

		if c.extraTags != nil {
			for _, t := range c.extraTags {
				if !x.HasTag([]byte(t.Key)) {
					x.AddTag(t.Key, t.GetS())
				}
			}
		}

		pt := FromModelsLP(x)
		if pt == nil {
			continue
		}

		if c.keySorted {
			kvs := KVs(pt.pt.Fields)
			sort.Sort(kvs)
			pt.pt.Fields = kvs
		}

		if c.callback != nil {
			newPoint, err := c.callback(pt)
			if err != nil {
				return nil, err
			}

			if newPoint == nil {
				return nil, fmt.Errorf("no point")
			}
		}

		pt = chk.check(pt)
		pt.pt.Warns = chk.warns
		chk.reset()

		// re-sort again: check may update pt.pt.Fields
		if c.keySorted {
			kvs := KVs(pt.pt.Fields)
			sort.Sort(kvs)
			pt.pt.Fields = kvs
		}

		res = append(res, pt)
	}

	return res, nil
}
