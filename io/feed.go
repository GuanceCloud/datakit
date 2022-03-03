package io

import (
	"fmt"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var l = logger.DefaultSLogger("event")

// NamedFeed Deprecated.
func NamedFeed(data []byte, category, name string) error {
	pts, err := lp.ParsePoints(data, nil)
	if err != nil {
		return err
	}

	x := []*Point{}
	for _, pt := range pts {
		x = append(x, &Point{Point: pt})
	}

	return defaultIO.DoFeed(x, category, name, nil)
}

type Measurement interface {
	LineProto() (*Point, error)
	Info() *MeasurementInfo
}

// NamedFeedEx Deprecated.
func NamedFeedEx(name, category, metric string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) error {
	var ts time.Time
	if len(t) > 0 {
		ts = t[0]
	} else {
		ts = time.Now().UTC()
	}

	pt, err := lp.MakeLineProtoPoint(metric, tags, fields,
		&lp.Option{
			ExtraTags: extraTags,
			Strict:    true,
			Time:      ts,
			Precision: "n",
		})
	if err != nil {
		return err
	}

	return defaultIO.DoFeed([]*Point{{pt}}, category, name, nil)
}

func Feed(name, category string, pts []*Point, opt *Option) error {
	if len(pts) == 0 {
		return fmt.Errorf("no points")
	}

	return defaultIO.DoFeed(pts, category, name, opt)
}

func FeedLastError(inputName string, err string) {
	select {
	case defaultIO.inLastErr <- &lastError{
		from: inputName,
		err:  err,
		ts:   time.Now(),
	}:
		addReporter(Reporter{Status: "warning", Message: fmt.Sprintf("%s %s", inputName, err), Category: "input"})
	case <-datakit.Exit.Wait():
		log.Warnf("%s feed last error skipped on global exit", inputName)
	}
}

func SelfError(err string) {
	FeedLastError(datakit.DatakitInputName, err)
}

func FeedEventLog(reporter *Reporter) {
	measurement := getReporterMeasurement(reporter)
	err := FeedMeasurement("datakit", datakit.Logging, []Measurement{measurement}, nil)
	if err != nil {
		l.Errorf("send datakit logging error: %s", err.Error())
	}
}

type ReporterMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func getReporterMeasurement(reporter *Reporter) ReporterMeasurement {
	now := time.Now()
	m := ReporterMeasurement{
		name: "datakit",
		ts:   now,
	}

	m.tags = reporter.Tags()
	m.fields = reporter.Fields()
	return m
}

func FeedMeasurement(name, category string, measurements []Measurement, opt *Option) error {
	if len(measurements) == 0 {
		return fmt.Errorf("no points")
	}

	pts, err := GetPointsFromMeasurement(measurements)
	if err != nil {
		return err
	}

	return Feed(name, category, pts, opt)
}

func GetPointsFromMeasurement(measurements []Measurement) ([]*Point, error) {
	var pts []*Point
	for _, m := range measurements {
		if pt, err := m.LineProto(); err != nil {
			return nil, err
		} else {
			pts = append(pts, pt)
		}
	}
	return pts, nil
}

func (e ReporterMeasurement) LineProto() (*Point, error) {
	return MakePoint(e.name, e.tags, e.fields, e.ts)
}

func (e ReporterMeasurement) Info() *MeasurementInfo {
	return &MeasurementInfo{}
}

type MeasurementInfo struct {
	Name   string
	Desc   string
	Type   string
	Fields map[string]interface{}
	Tags   map[string]interface{}
}
