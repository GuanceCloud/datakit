package sender

import (
	"fmt"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type Point interface{}

type Writer interface {
	Write(string, []*influxdb.Point) error
}

type Sinker struct {
	Store map[string]Writer
}

func (d *Sinker) Write(category string, pts []*influxdb.Point) error {
	writer, ok := d.Store["dataway"]
	if !ok {
		return fmt.Errorf("invalid category: %s, found no sinker", category)
	}
	return writer.Write(category, pts)
}
