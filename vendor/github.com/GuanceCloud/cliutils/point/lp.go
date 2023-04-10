// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"fmt"
	"regexp"
	"time"

	"github.com/influxdata/influxdb1-client/models"
)

type Precision int

const (
	NS Precision = iota // nano-second
	US                  // micro-second
	MS                  // milli-second
	S                   // second
	M                   // minute
	H                   // hour

	// XXX: not used.
	D // day
	W // week
)

func (p Precision) String() string {
	switch p {
	case NS:
		return "n"
	case US:
		return "u"
	case MS:
		return "ms"
	case S:
		return "s"
	case M:
		return "m"
	case H:
		return "h"
	case D:
		return "d"
	case W:
		return "w"
	default:
		return "unknown"
	}
}

func PrecStr(s string) Precision {
	switch s {
	case "n":
		return NS
	case "u":
		return US
	case "ms":
		return MS
	case "s":
		return S
	case "m":
		return M
	case "h":
		return H
	default:
		return NS
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
	}

	defer PutCfg(c)

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
				if !x.HasTag(t.Key) {
					x.AddTag(string(t.Key), string(t.GetD()))
				}
			}
		}

		pt := FromModelsLP(x)

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
		pt.warns = chk.warns
		chk.reset()

		res = append(res, pt)
	}

	return res, nil
}
