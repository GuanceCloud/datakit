package aggregate

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

// AggregatorConfigure is the top-level aggregator configure on single workspace.
type AggregatorConfigure struct {
	DefaultWindow    time.Duration    `toml:"default_window" json:"default_window"`
	AggregateRules   []*AggregateRule `toml:"aggregate_rules" json:"aggregate_rules"`
	DefaultAction    Action           `toml:"action" json:"action"`
	DeleteRulesPoint bool             `toml:"delete_rules_point" json:"delete_rules_point"`

	hash  uint64
	raw   []byte
	calcs map[uint64]Calculator
}

func (ac *AggregatorConfigure) UnmarshalTOML(data interface{}) error {
	type rawAggregatorConfigure struct {
		DefaultWindow    any              `json:"default_window"`
		AggregateRules   []*AggregateRule `json:"aggregate_rules"`
		DefaultAction    Action           `json:"action"`
		DeleteRulesPoint bool             `json:"delete_rules_point"`
	}

	var raw rawAggregatorConfigure
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(j, &raw); err != nil {
		return err
	}

	switch v := raw.DefaultWindow.(type) {
	case nil:
	case string:
		if v != "" {
			d, err := time.ParseDuration(v)
			if err != nil {
				return err
			}
			ac.DefaultWindow = d
		}
	case float64:
		ac.DefaultWindow = time.Duration(int64(v))
	default:
		return fmt.Errorf("unsupported default_window type %T", v)
	}

	ac.AggregateRules = raw.AggregateRules
	ac.DefaultAction = raw.DefaultAction
	ac.DeleteRulesPoint = raw.DeleteRulesPoint

	return nil
}

func (ac *AggregatorConfigure) doHash() {
	if j, err := json.Marshal(ac.hashView()); err != nil {
		return
	} else {
		ac.raw = j
		ac.hash = xxhash.Sum64(j)
	}
}

func (ac *AggregatorConfigure) hashView() any {
	type aggregateRuleView struct {
		Name       string                            `json:"name"`
		Selector   *RuleSelector                     `json:"select"`
		Groupby    []string                          `json:"group_by"`
		Algorithms map[string]*AggregationAlgoConfig `json:"algorithms"`
	}

	type aggregatorConfigureView struct {
		DefaultWindow    time.Duration        `json:"default_window"`
		AggregateRules   []*aggregateRuleView `json:"aggregate_rules"`
		DefaultAction    Action               `json:"action"`
		DeleteRulesPoint bool                 `json:"delete_rules_point"`
	}

	view := &aggregatorConfigureView{
		DefaultWindow:    ac.DefaultWindow,
		DefaultAction:    ac.DefaultAction,
		DeleteRulesPoint: ac.DeleteRulesPoint,
	}

	for _, rule := range ac.AggregateRules {
		if rule == nil {
			view.AggregateRules = append(view.AggregateRules, nil)
			continue
		}

		view.AggregateRules = append(view.AggregateRules, &aggregateRuleView{
			Name:       rule.Name,
			Selector:   rule.Selector,
			Groupby:    rule.Groupby,
			Algorithms: rule.Algorithms,
		})
	}

	return view
}

// AggregateRule configured a specific aggregate rule.
type AggregateRule struct {
	Name       string                            `toml:"name" json:"name"`
	Selector   *RuleSelector                     `toml:"select" json:"select"`
	Groupby    []string                          `toml:"group_by" json:"group_by"`
	Algorithms map[string]*AggregationAlgoConfig `toml:"algorithms" json:"algorithms"`

	aggregationOpts map[string]*AggregationAlgo
}

type AggregationAlgoConfig struct {
	Method        string                `toml:"method" json:"method"`
	SourceField   string                `toml:"source_field" json:"source_field"`
	Window        int64                 `toml:"window" json:"window"`
	AddTags       map[string]string     `toml:"add_tags" json:"add_tags"`
	HistogramOpts *HistogramOptions     `toml:"histogram_opts" json:"histogram_opts"`
	ExpoOpts      *ExpoHistogramOptions `toml:"expo_opts" json:"expo_opts"`
	QuantileOpts  *QuantileOptions      `toml:"quantile_opts" json:"quantile_opts"`
}

func (cfg *AggregationAlgoConfig) ToAggregationAlgo() *AggregationAlgo {
	if cfg == nil {
		return nil
	}

	algo := &AggregationAlgo{
		Method:      cfg.Method,
		SourceField: cfg.SourceField,
		Window:      cfg.Window,
	}

	if len(cfg.AddTags) > 0 {
		algo.AddTags = make(map[string]string, len(cfg.AddTags))
		for k, v := range cfg.AddTags {
			algo.AddTags[k] = v
		}
	}

	switch {
	case cfg.HistogramOpts != nil:
		algo.Options = &AggregationAlgo_HistogramOpts{HistogramOpts: cfg.HistogramOpts}
	case cfg.ExpoOpts != nil:
		algo.Options = &AggregationAlgo_ExpoOpts{ExpoOpts: cfg.ExpoOpts}
	case cfg.QuantileOpts != nil:
		algo.Options = &AggregationAlgo_QuantileOpts{QuantileOpts: cfg.QuantileOpts}
	}

	return algo
}

// Setup initializes the aggregator configuration, validates rules, and prepares calculators.
func (ac *AggregatorConfigure) Setup() error {
	switch ac.DefaultAction {
	case "", ActionPassThrough, ActionDrop:
	default:
		return fmt.Errorf("invalid action: %s", ac.DefaultAction)
	}

	for _, ar := range ac.AggregateRules {
		if ar == nil {
			return fmt.Errorf("aggregate rule is nil")
		}
		if ar.Selector == nil {
			return fmt.Errorf("aggregate rule %q missing selector", ar.Name)
		}
		if err := ar.Selector.Setup(); err != nil {
			return err
		}

		algorithms, err := ar.setupAlgorithms()
		if err != nil {
			return fmt.Errorf("aggregate rule %q: %w", ar.Name, err)
		}

		for key, algo := range algorithms {
			if err := validateAggregationAlgo(key, algo); err != nil {
				return fmt.Errorf("aggregate rule %q: %w", ar.Name, err)
			}
			if algo.Window <= 10 {
				algo.Window = int64(ac.DefaultWindow)
			}
		}
		ar.aggregationOpts = algorithms

		sort.Strings(ar.Groupby)
	}

	ac.doHash()
	ac.calcs = map[uint64]Calculator{}

	return nil
}

func validateAggregationAlgo(key string, algo *AggregationAlgo) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("algorithm name is empty")
	}
	if algo == nil {
		return fmt.Errorf("algorithm %q is nil", key)
	}

	method := NormalizeAlgoMethod(algo.Method)
	switch method {
	case SUM, AVG, COUNT, MIN, MAX, HISTOGRAM, STDEV, QUANTILES, COUNT_DISTINCT, LAST, FIRST:
	case EXPO_HISTOGRAM:
		return fmt.Errorf("algorithm %q: method %q is not supported", key, algo.Method)
	case METHOD_UNSPECIFIED:
		return fmt.Errorf("algorithm %q missing method", key)
	default:
		return fmt.Errorf("algorithm %q has unknown method %q", key, algo.Method)
	}

	if method == QUANTILES {
		opt, ok := algo.Options.(*AggregationAlgo_QuantileOpts)
		if !ok || opt == nil || opt.QuantileOpts == nil || len(opt.QuantileOpts.Percentiles) == 0 {
			return fmt.Errorf("algorithm %q: quantiles requires quantile_opts.percentiles", key)
		}
		for _, percentile := range opt.QuantileOpts.Percentiles {
			if percentile < 0 || percentile > 1 {
				return fmt.Errorf("algorithm %q: percentile %v out of range [0,1]", key, percentile)
			}
		}
	}

	return nil
}

func (ar *AggregateRule) setupAlgorithms() (map[string]*AggregationAlgo, error) {
	if len(ar.Algorithms) == 0 {
		return nil, nil
	}

	res := make(map[string]*AggregationAlgo, len(ar.Algorithms))
	for key, cfg := range ar.Algorithms {
		res[key] = cfg.ToAggregationAlgo()
	}

	return res, nil
}

// SelectPoints selects points from the input slice based on aggregate rules.
func (ac *AggregatorConfigure) SelectPoints(pts []*point.Point) (groups [][]*point.Point) {
	for _, ar := range ac.AggregateRules {
		groups = append(groups, ar.SelectPoints(pts))
	}
	return
}
