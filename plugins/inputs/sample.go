package inputs

import (
	"errors"
	"fmt"
	"strconv"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	DefSampleHandler SampleFunc = func(getSamplePoint Converter, rate, scope int) bool {
		samplePoint, err := getSamplePoint.Convert()
		if err != nil {
			return false
		} else {
			return (samplePoint % scope) < rate
		}
	}
	DefConvertHandler = func(point *io.Point, key string) ConvertFunc {
		return func() (int, error) {
			tags := point.Tags()
			for k, v := range tags {
				if key == k {
					return strconv.Atoi(v)
				}
			}
			fields, err := point.Fields()
			if err != nil {
				return 0, err
			}
			for k, v := range fields {
				if key == k {
					if id, ok := v.(int); !ok {
						return 0, errors.New("type assertion failed, value is not int type")
					} else {
						return id, nil
					}
				}
			}

			return 0, fmt.Errorf("can not find value of %s", key)
		}
	}
)

type Converter interface {
	Convert() (int, error)
}

type ConvertFunc func() (int, error)

func (this ConvertFunc) Convert() (int, error) {
	return this()
}

type Sampler interface {
	Sample(getSamplePoint Converter, rate, scope int) bool
}

type SampleFunc func(getSamplePoint Converter, rate, scope int) bool

func (this SampleFunc) Sample(getSamplePoint Converter, rate, scope int) bool {
	return this(getSamplePoint, rate, scope)
}

type TraceSampleConfig struct {
	Rate           int               `toml:"rate"`
	Scope          int               `toml:"scope"`
	IgnoreList     map[string]string `toml:ignore_list`
	SampleHandler  Sampler
	Key            string
	ConvertHandler func(point *io.Point, key string) ConvertFunc
}

func (this *TraceSampleConfig) TraceSample(points []*io.Point) []*io.Point {
	if len(this.IgnoreList) != 0 {
		points = this.ignoreFilter(points)
	}
	if this.SampleHandler != nil && this.Key != "" && this.ConvertHandler != nil {

	}

	return points
}
