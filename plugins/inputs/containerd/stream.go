// +build linux

package containerd

import (
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

func (s *stream) start(wg *sync.WaitGroup) {
	defer wg.Done()

	if s.sub.Measurement == "" {
		log.Printf("E! [Containerd] subscribe namespace '%s', error: invalid measurement\n", s.sub.Namespace)
		return
	}

	if s.isAll {
		log.Printf("I! [Containerd] subscribe namespace '%s' start, collect all id\n", s.sub.Namespace)
	} else {
		log.Printf("I! [Containerd] subscribe namespace '%s' start, collect len %d id\n", s.sub.Namespace, len(s.ids))
	}

	ticker := time.NewTicker(time.Second * s.sub.Cycle)
	defer ticker.Stop()

	for {
		select {
		case <-s.cont.ctx.Done():
			log.Printf("I! [Containerd] subscribe namespace '%s' stop\n", s.sub.Namespace)

		case <-ticker.C:
			if err := s.exec(); err != nil {
				log.Printf("E! [Containerd] subscribe namespace '%s', error: %s\n", s.sub.Namespace, err.Error())
			}
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
