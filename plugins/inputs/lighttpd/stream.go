package lighttpd

import (
	"fmt"
	"sync"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type stream struct {
	lighttpd *Lighttpd
	//
	sub *Subscribe
	//
	statusURL string
	// v1 or v2
	statusVersion Version
	//
	startTime time.Time
	// mate data
	points []*influxdb.Point
}

func newStream(sub *Subscribe, lt *Lighttpd) *stream {
	var url string
	var v Version

	switch sub.LighttpdVersion {
	case "v1":
		url = fmt.Sprintf("%s?json", sub.LighttpdURL)
		v = v1
	case "v2":
		url = fmt.Sprintf("%s?format=plain", sub.LighttpdURL)
		v = v2
	default:
		// nil
	}

	return &stream{
		lighttpd:      lt,
		sub:           sub,
		statusURL:     url,
		statusVersion: v,
		startTime:     time.Now(),
	}
}

func (s *stream) start(wg *sync.WaitGroup) error {
	defer wg.Done()

	if s.sub.Measurement == "" {
		return fmt.Errorf("invalid measurement")
	}

	ticker := time.NewTicker(time.Second * s.sub.Cycle)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.exec()
		default:
			// nil
		}
	}
}

func (s *stream) exec() error {

	pt, err := LighttpdStatusParse(s.statusURL, s.statusVersion, s.sub.Measurement)
	if err != nil {
		return err
	}

	s.points = []*influxdb.Point{pt}

	return s.flush()
}

func (s *stream) flush() (err error) {
	// fmt.Printf("%v\n", s.points)
	err = s.lighttpd.ProcessPts(s.points)
	s.points = nil

	return err
}
