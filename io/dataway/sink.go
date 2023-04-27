// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"fmt"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/parser"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

type Sinker struct {
	Categories      []string `toml:"categories" json:"categories"`
	Filters         []string `toml:"filters" json:"filters"`
	URL             string   `toml:"url" json:"url"`
	Proxy           string   `toml:"proxy" json:"proxy"`
	TokenDeprecated string   `toml:"token,omitempty" json:"token,omitempty"`

	conditions parser.WhereConditions
	ep         *endPoint
	cats       []point.Category
}

var sinkerAPIs = []point.Category{
	point.MetricDeprecated,
	point.Metric,
	point.Network,
	point.KeyEvent,
	point.Object,
	point.CustomObject,
	point.Logging,
	point.Tracing,
	point.RUM,
	point.Security,
	point.Profiling,
}

func (s *Sinker) String() string {
	// TODO: make it more human readable.
	return fmt.Sprintf("[categories: %s][filters: %s][URL: %s][proxy: %s]",
		strings.Join(s.Categories, ","),
		strings.Join(s.Filters, ","),
		s.URL, s.Proxy,
	)
}

func (s *Sinker) expectedCategory(cat point.Category) bool {
	for _, c := range s.cats {
		if c == cat {
			return true
		}
	}

	return false
}

func (s *Sinker) sink(cat point.Category, pts []*dkpt.Point) (remainIndices []int, err error) {
	var (
		sinkPts []*dkpt.Point
		fok     bool
	)

	defer func() {
		sinkCounterVec.WithLabelValues(cat.String()).Inc()
		if len(remainIndices) > 0 {
			notSinkPtsVec.WithLabelValues(cat.String()).Add(float64(len(remainIndices)))
		}
	}()

	if !s.expectedCategory(cat) {
		log.Debugf("category %q not expected for my categories %v", cat, s.Categories)
		for i := range pts {
			remainIndices = append(remainIndices, i)
		}
		return
	}

	// Unconditional: all points send to the sinker
	if len(s.conditions) == 0 {
		sinkPts = pts
		goto __write
	}

	for i, pt := range pts {
		fok, err = filter.CheckPointFiltered(s.conditions, cat.URL(), pt)
		if err != nil {
			log.Warnf("pt.Fields: %s, ignored", err.Error())

			remainIndices = append(remainIndices, i) // filter failed point NOT sinked
			continue
		}

		if fok {
			sinkPts = append(sinkPts, pt)
		} else {
			remainIndices = append(remainIndices, i)
		}
	}

	if len(sinkPts) == 0 {
		return
	}

__write:
	if err = s.write(cat.URL(), sinkPts); err != nil {
		return
	}

	log.Debugf("after sinker %d points on %v remain %d points",
		len(pts), s.Categories, len(remainIndices))

	return remainIndices, nil
}

func (s *Sinker) write(category string, pts []*dkpt.Point) error {
	return s.ep.writePoints(
		&writer{
			isSinker: true,
			category: category,
			pts:      pts,
		})
}

func (s *Sinker) Setup() error {
	if err := s.setupFilters(); err != nil {
		return err
	}

	var apis []string
	for _, x := range sinkerAPIs {
		apis = append(apis, x.URL())
	}

	ep, err := newEndpoint(s.URL, withAPIs(apis), withProxy(s.Proxy)) // no proxy allowed
	if err != nil {
		return err
	}

	// use deprected token field if token not exist in s.URL
	if ep.token == "" && s.TokenDeprecated != "" {
		ep.token = s.TokenDeprecated
	}

	log.Infof("set endpoint %s on sinker %s", ep, s)
	s.ep = ep

	for _, c := range s.Categories {
		s.cats = append(s.cats, point.CatAlias(c))
	}

	return nil
}

func (s *Sinker) setupFilters() error {
	if len(s.Filters) == 0 {
		return nil
	}

	// parse filter conditions
	if conditions, err := filter.GetConds(s.Filters); err != nil {
		return fmt.Errorf("GetConds(%+#v): %w", s.Filters, err)
	} else {
		s.conditions = conditions
	}

	return nil
}
