// +build linux

package containerd

import (
	"errors"
	// "fmt"
	"log"
	"sync"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type stream struct {
	cont *Containerd
	//
	sub *Subscribe
	// get all ids metrics
	isAll bool
	// id cache
	ids map[string]byte
	// mate data
	points []*influxdb.Point
}

func newStream(sub *Subscribe, cont *Containerd) *stream {
	return &stream{
		cont:  cont,
		sub:   sub,
		isAll: len(sub.IDList) == 1 && sub.IDList[0] == "*",
		ids: func() map[string]byte {
			m := make(map[string]byte)
			for _, v := range sub.IDList {
				m[v] = '0'
			}
			return m
		}(),
	}

}

func (s *stream) start(wg *sync.WaitGroup) error {
	defer wg.Done()

	if s.sub.Measurement == "" {
		err := errors.New("invalid measurement")
		log.Printf("E! [Ccontainerd] subscribe namespace: %s, error: %s\n", s.sub.Namespace, err.Error())
		return err
	}

	if s.isAll {
		log.Printf("I! [Ccontainerd] subscribe namespace: %s, collect all id\n", s.sub.Namespace)
	} else {
		log.Printf("I! [Ccontainerd] subscribe namespace: %s, collect len %d id\n", s.sub.Namespace, len(s.ids))
	}

	ticker := time.NewTicker(time.Second * s.sub.Cycle)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.exec(); err != nil {
				log.Printf("E! [Ccontainerd] subscribe namespace: %s, error: %s\n", s.sub.Namespace, err.Error())
			}
		default:
			// nil
		}
	}
}

func (s *stream) exec() error {
	err := s.processMetrics()
	if err != nil {
		s.points = nil
		return err
	}
	return s.flush()
}

func (s *stream) flush() (err error) {
	// fmt.Printf("%v\n", s.points)
	err = s.cont.ProcessPts(s.points)
	s.points = nil
	return err
}
