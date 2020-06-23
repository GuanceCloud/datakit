// +build !solaris

package tailf

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hpcloud/tail"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

// const defaultWatchMethod = "inotify"

type stream struct {
	tailf  *Tailf
	sub    *Subscribe
	offset int64
	// mate data
	points []*influxdb.Point
}

func newStream(sub *Subscribe, tailf *Tailf) *stream {
	return &stream{
		tailf:  tailf,
		sub:    sub,
		offset: 0,
	}
}

func (s *stream) start(wg *sync.WaitGroup) {
	defer wg.Done()

	if s.sub.File == config.Cfg.MainCfg.Log {
		log.Printf("E! [Tailf] subscribe file: %s, error: cannot collect datakit log !\n", s.sub.File)
		return

	}

	if s.sub.Measurement == "" {
		log.Printf("E! [Tailf] subscribe file: %s, error: invalid measurement\n", s.sub.File)
		return
	}

	var poll bool
	if s.sub.WatchMethod == "poll" {
		poll = true
	}

	var seek *tail.SeekInfo

	if !s.sub.Pipe && !s.sub.FormBeginning {
		seek = &tail.SeekInfo{
			Whence: 0,
			Offset: s.offset,
		}
		log.Printf("I! [Tailf] subscribe file: %s, using offset %d\n", s.sub.File, s.offset)
	} else {
		seek = &tail.SeekInfo{
			Whence: 2,
			Offset: 0,
		}
	}

	tailer, err := tail.TailFile(s.sub.File,
		tail.Config{
			ReOpen:    true,
			Follow:    true,
			Location:  seek,
			MustExist: true,
			Poll:      poll,
			Pipe:      s.sub.Pipe,
			Logger:    tail.DiscardingLogger,
		})
	if err != nil {
		log.Printf("E! [Tailf] subscribe file: %s, failed to open file: %s\n", s.sub.File, err.Error())
		return
	}

	go s.catchStop(tailer)

	if err := s.exec(tailer); err != nil {
		log.Printf("E! [Tailf] subscribe file: %s, exec failed: %s\n", s.sub.File, err.Error())
		return
	}

}

func (s *stream) exec(tailer *tail.Tail) error {

	var fields = make(map[string]interface{})

	for line := range tailer.Lines {
		if line.Err != nil {
			log.Printf("E! [Tailf] subscribe file: %s, tailing error: %s\n", s.sub.File, line.Err.Error())
			continue
		}

		text := strings.TrimRight(line.Text, "\r")
		fields["__content"] = text

		pt, err := influxdb.NewPoint(s.sub.Measurement, nil, fields, time.Now())
		if err != nil {
			log.Printf("E! [Tailf] subscribe file: %s, new points error: %s\n", s.sub.File, line.Err.Error())
			continue
		}

		s.points = []*influxdb.Point{pt}

		if err := s.flush(); err != nil {
			log.Printf("E! [Tailf] subscribe file: %s, data flush error: %s\n", s.sub.File, err.Error())
		}

	}

	log.Printf("I! [Tailf] subscribe file: %s, tailf stop\n", s.sub.File)
	if err := tailer.Err(); err != nil {
		log.Printf("E! [Tailf] subscribe file: %s, tailing error: %s\n", s.sub.File, err.Error())
		return err
	}

	return nil
}

func (s *stream) catchStop(tailer *tail.Tail) {

	if _, ok := <-s.tailf.ctx.Done(); ok {
		if !s.sub.Pipe && !s.sub.FormBeginning {
			offset, err := tailer.Tell()
			if err == nil {
				log.Printf("I! [Tailf] subscribe file: %s, recording offset %d\n", s.sub.File, offset)
			} else {
				log.Printf("I! [Tailf] subscribe file: %s, recording offset %s\n", s.sub.File, err.Error())
			}
			s.offset = offset
		}
		err := tailer.Stop()
		if err != nil {
			log.Printf("I! [Tailf] subscribe file: %s stop\n", s.sub.File)
		}
		return
	}

}

func (s *stream) flush() (err error) {
	// fmt.Printf("%v\n", s.points)
	err = s.tailf.ProcessPts(s.points)
	s.points = nil
	return err
}
