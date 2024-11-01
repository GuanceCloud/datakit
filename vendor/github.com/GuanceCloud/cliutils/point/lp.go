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
	PrecNS  Precision = iota // nano-second
	PrecUS                   // micro-second
	PrecMS                   // milli-second
	PrecS                    // second
	PrecM                    // minute
	PrecH                    // hour
	PrecD                    // day
	PrecW                    // week
	PrecDyn                  // dynamic precision
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

	ptTime := time.Unix(0, c.timestamp)
	if c.timestamp == 0 {
		ptTime = time.Now()
	}

	// NOTE: always parse point with precision ns, the caller should
	// adjust the time according to specific precision setting.
	lppts, err := models.ParsePointsWithPrecision(data, ptTime, "ns")
	if err != nil {
		return nil, fmt.Errorf("%w: %s. Origin data: %q",
			ErrInvalidLineProtocol, err, data)
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
