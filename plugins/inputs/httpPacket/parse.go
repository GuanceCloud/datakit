package httpPacket

import (
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

func ParsePoints(data []byte, prec string) ([]*influxdb.Point, error) {
	points, err := influxm.ParsePointsWithPrecision(data, time.Now().UTC(), prec)
	if err != nil {
		return nil, err
	}

	pts := []*influxdb.Point{}
	for _, pt := range points {
		measurement := string(pt.Name())
		tags := map[string]string{}

		for _, tag := range pt.Tags() {
			tags[string(tag.Key)] = string(tag.Value)
		}

		fields, _ := pt.Fields()
		pt, err := influxdb.NewPoint(
			measurement,
			tags,
			fields,
			pt.Time(),
		)

		if err != nil {
			return nil, err
		}

		pts = append(pts, pt)
	}

	return pts, nil
}
