package aggregate

import (
	"fmt"

	"github.com/GuanceCloud/cliutils"
	fp "github.com/GuanceCloud/cliutils/filter"
	"github.com/GuanceCloud/cliutils/point"
)

// RuleSelector selects measurements and fields among points.
type RuleSelector struct {
	Category     string   `toml:"category" json:"category"`
	Measurements []string `toml:"measurements" json:"measurements"`
	MetricName   []string `toml:"metric_name" json:"metric_name"`
	Condition    string   `toml:"condition" json:"condition"`

	measurementsWhitelist []*cliutils.WhiteListItem
	fieldsWhitelist       []*cliutils.WhiteListItem
	conds                 fp.WhereConditions
}

// Setup initializes the rule selector with validation and prepares whitelists.
func (rs *RuleSelector) Setup() error {
	switch point.CatString(rs.Category) { // category required
	case point.Metric,
		point.Network,
		point.KeyEvent,
		point.Object,
		point.ObjectChange,
		point.CustomObject,
		point.Logging,
		point.Tracing,
		point.RUM,
		point.Security,
		point.Profiling,
		point.DialTesting,
		point.AgentLLM:
	case point.UnknownCategory,
		point.DynamicDWCategory,
		point.MetricDeprecated,
		point.ExecutionLog,
		point.LLM:
		return fmt.Errorf("invalid category: %s", rs.Category)
	default:
		return fmt.Errorf("invalid category: %s", rs.Category)
	}

	if rs.Condition != "" {
		ast, err := fp.GetConds(rs.Condition)
		if err != nil {
			return err
		}
		rs.conds = ast
	}

	if len(rs.Measurements) > 0 {
		for _, m := range rs.Measurements {
			rs.measurementsWhitelist = append(rs.measurementsWhitelist, cliutils.NewWhiteListItem(m))
		}
	}

	if len(rs.MetricName) > 0 {
		for _, f := range rs.MetricName {
			rs.fieldsWhitelist = append(rs.fieldsWhitelist, cliutils.NewWhiteListItem(f))
		}
	}

	return nil
}

func (rs *RuleSelector) doSelect(groupby []string, pts []*point.Point) (res []*point.Point) {
	ptwrapper := &ptWrap{}

	for _, pt := range pts {
		ptname := pt.Name()

		if len(rs.measurementsWhitelist) > 0 {
			if !cliutils.WhiteListMatched(ptname, rs.measurementsWhitelist) {
				continue
			}
		}

		if len(rs.conds) > 0 {
			ptwrapper.Point = pt
			if x := rs.conds.Eval(ptwrapper); x < 0 {
				continue
			}
		}

		forkedPts := rs.selectKVS(false, pt)
		if len(forkedPts) > 0 {
			for _, tagKey := range groupby {
				if v := pt.GetTag(tagKey); v != "" {
					for i := range forkedPts {
						forkedPts[i].SetTag(tagKey, v)
					}
				}

				if v := pt.Get(tagKey); v != nil {
					for i := range forkedPts {
						forkedPts[i].Add(tagKey, v)
					}
				}
			}

			res = append(res, forkedPts...)
		}
	}

	return res
}

func (rs *RuleSelector) selectKVS(delKey bool, pt *point.Point) []*point.Point {
	var pts []*point.Point
	if len(rs.fieldsWhitelist) > 0 {
		for _, kv := range pt.KVs() {
			if !cliutils.WhiteListMatched(kv.Key, rs.fieldsWhitelist) {
				continue
			}

			if kv.IsTag {
				continue
			}
			var kvs point.KVs
			switch v := kv.Val.(type) {
			case *point.Field_F:
				kvs = kvs.Add(kv.Key, v.F)
			case *point.Field_I:
				kvs = kvs.Add(kv.Key, float64(v.I))
			case *point.Field_U:
				kvs = kvs.Add(kv.Key, float64(v.U))
			case *point.Field_S:
				kvs = kvs.Add(kv.Key, v.S)
			case *point.Field_D:
				kvs = kvs.Add(kv.Key, v.D)
			case *point.Field_B:
				kvs = kvs.Add(kv.Key, v.B)
			}
			if len(kvs) > 0 {
				if delKey {
					pt.Del(kv.Key)
				}
				pts = append(pts, point.NewPoint(pt.Name(), kvs, point.WithTime(pt.Time())))
			}
		}
	}

	return pts
}
