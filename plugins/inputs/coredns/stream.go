package coredns

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

var tagsWhiteList = map[string]byte{"version": '0'}

type stream struct {
	coredns *Coredns
	//
	sub *Subscribe
	// http://host:port/metrics
	address string
	//
	startTime time.Time
	// mate data
	points []*influxdb.Point
}

func newStream(sub *Subscribe, coredns *Coredns) *stream {
	return &stream{
		coredns:   coredns,
		sub:       sub,
		address:   fmt.Sprintf("http://%s:%d/metrics", sub.CorednsHost, sub.CorednsPort),
		startTime: time.Now(),
	}
}

func (s *stream) start(wg *sync.WaitGroup) error {
	defer wg.Done()

	if s.sub.Measurement == "" {
		err := errors.New("invalid measurement")
		log.Printf("E! [CoreDNS] subscribe %s:%d, error: %s\n", s.sub.CorednsHost, s.sub.CorednsPort, err.Error())
		return err
	}

	ticker := time.NewTicker(time.Second * s.sub.Cycle)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.exec(); err != nil {
				log.Printf("E! [CoreDNS] subscribe %s:%d, error: %s\n", s.sub.CorednsHost, s.sub.CorednsPort, err.Error())
			}
		default:
			// nil
		}
	}
}

func (s *stream) exec() error {

	resp, err := http.Get(s.address)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	metrics, err := ParseV2(resp.Body)
	if err != nil {
		return err
	}

	if len(metrics) == 0 {
		return errors.New("the metrics is empty")
	}

	var tags = make(map[string]string)
	var fields = make(map[string]interface{}, len(metrics))

	// prometheus to point
	for _, metric := range metrics {

		for k, v := range metric.Tags() {
			if _, ok := tagsWhiteList[k]; ok {
				tags[k] = v
			}
		}

		for k, v := range metric.Fields() {
			fields[k] = v
		}

	}

	pt, err := influxdb.NewPoint(s.sub.Measurement, tags, fields, metrics[0].Time())
	if err != nil {
		return err
	}

	s.points = []*influxdb.Point{pt}

	return s.flush()
}

func (s *stream) flush() (err error) {
	// fmt.Printf("%v\n", s.points)
	err = s.coredns.ProcessPts(s.points)
	s.points = nil
	return err
}
