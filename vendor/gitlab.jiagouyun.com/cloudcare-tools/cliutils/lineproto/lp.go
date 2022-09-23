package lineproto

import (
	"time"

	lp "github.com/influxdata/line-protocol/v2/lineprotocol"
)

type Tag struct {
	Key, Val []byte
}

type Field struct {
	Key []byte
	Val lp.Value
}

type Point struct {
	Measurement []byte
	Tags        []*Tag
	Fields      []*Field
	Time        time.Time
}

func Parse(data []byte, opt *Option) ([]*Point, error) {
	dec := lp.NewDecoderWithBytes(data)

	var pts []*Point

	for dec.Next() {
		pt := &Point{}

		m, err := dec.Measurement()
		if err != nil {
			return pts, err
		}

		pt.Measurement = m

		for {
			k, v, err := dec.NextTag()
			if err != nil {
				return pts, err
			}

			if k == nil {
				break
			}

			pt.Tags = append(pt.Tags, &Tag{Key: k, Val: v})
		}

		for {
			k, v, err := dec.NextField()
			if err != nil {
				return pts, err
			}

			if k == nil {
				break
			}

			pt.Fields = append(pt.Fields, &Field{Key: k, Val: v})
		}

		t, err := dec.Time(opt.PrecisionV2, time.Time{})
		if err != nil {
			return pts, err
		}

		pt.Time = t
		pts = append(pts, pt)
	}

	return pts, nil
}

func Encode(pts []*Point, opt *Option) ([]byte, error) {
	enc := &lp.Encoder{}

	enc.SetPrecision(opt.PrecisionV2)
	for _, pt := range pts {
		enc.StartLineRaw(pt.Measurement)

		for _, t := range pt.Tags {
			enc.AddTagRaw(t.Key, t.Val)
		}

		for _, f := range pt.Fields {
			enc.AddFieldRaw(f.Key, f.Val)
		}

		enc.EndLine(pt.Time)
	}

	return enc.Bytes(), nil
}
