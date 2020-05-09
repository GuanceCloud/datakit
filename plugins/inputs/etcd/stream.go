package etcd

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

var _ETCD_ACTION_LIST = map[string]byte{
	"create":       '0',
	"set":          '0',
	"get":          '0',
	"getRecursive": '0',
}

type stream struct {
	etc *Etcd
	//
	sub *Subscribe
	// http://host:port/metrics
	address string
	//
	startTime time.Time
	// mate data
	points []*influxdb.Point
	// HTTPS TLS
	tlsConfig *tls.Config
}

func newStream(sub *Subscribe, etc *Etcd) *stream {
	return &stream{
		etc: etc,
		sub: sub,
		address: func() string {
			// usage HTTPS
			if sub.CacertFile != "" && sub.CertFile != "" && sub.KeyFile != "" {
				return fmt.Sprintf("https://%s:%d/metrics", sub.EtcdHost, sub.EtcdPort)
			}
			return fmt.Sprintf("http://%s:%d/metrics", sub.EtcdHost, sub.EtcdPort)
		}(),
		startTime: time.Now(),
	}
}

func (s *stream) start(wg *sync.WaitGroup) error {
	defer wg.Done()

	if s.sub.Measurement == "" {
		err := errors.New("invalid measurement")
		log.Printf("E! [Etcd] subscribe %s:%d, error: %s\n", s.sub.EtcdHost, s.sub.EtcdPort, err.Error())
		return err
	}

	// usage HTTPS
	if s.sub.CacertFile != "" && s.sub.CertFile != "" && s.sub.KeyFile != "" {

		tc, err := TLSConfig(s.sub.CacertFile, s.sub.CertFile, s.sub.KeyFile)
		if err != nil {
			log.Printf("E! [Etcd] subscribe %s:%d, build TLSConfig failed: %s\n", s.sub.EtcdHost, s.sub.EtcdPort, err.Error())
			return err
		}

		s.tlsConfig = tc
		log.Printf("I! [Etcd] subscribe %s:%d, build TLSConfig success\n", s.sub.EtcdHost, s.sub.EtcdPort)

	} else {
		log.Printf("I! [Etcd] subscribe %s:%d, usage HTTP connection\n", s.sub.EtcdHost, s.sub.EtcdPort)
	}

	ticker := time.NewTicker(time.Second * s.sub.Cycle)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.exec(); err != nil {
				log.Printf("E! [Etcd] subscribe %s:%d, exec failed: %s\n", s.sub.EtcdHost, s.sub.EtcdPort, err.Error())
			}
		default:
			// nil
		}
	}
}

func (s *stream) exec() error {

	client := &http.Client{}
	client.Timeout = time.Second * 5

	if s.tlsConfig != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: s.tlsConfig,
		}
	}

	resp, err := client.Get(s.address)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	metrics, err := ParseV2(resp.Body)
	if err != nil {
		return err
	}

	if len(metrics) == 0 {
		return errors.New("The metrics is nil")
	}

	var tags = make(map[string]string)
	var fields = make(map[string]interface{}, len(metrics))

	// prometheus to point
	for _, metric := range metrics {

		for k, v := range metric.Tags() {

			// Save 4 fields from prometheus.
			// It's STUPID!

			// prometheus,action=create etcd_debugging_store_writes_total=1 1586857769285050000
			// prometheus,action=set etcd_debugging_store_writes_total=4 1586857769285050000
			// prometheus,action=get etcd_debugging_store_reads_total=1 1586857769285050000
			// prometheus,action=getRecursive etcd_debugging_store_reads_total=2 1586857769285050000
			// ============== TO =============
			// etcd_debugging_store_writes_total_create=1
			// etcd_debugging_store_writes_total_set=4
			// etcd_debugging_store_reads_total_get=1
			// etcd_debugging_store_reads_total_getRecursive=2

			if _, ok := _ETCD_ACTION_LIST[v]; ok {
				for _k, _v := range metric.Fields() {
					fields[_k+"_"+v] = _v
					goto END_ACTION
				}
			}

			if _, ok := etcdCollectList[k]; ok {
				tags[k] = v
			}
		}

		for k, v := range metric.Fields() {
			if _, ok := etcdCollectList[k]; ok {
				fields[k] = v
			}
		}

	END_ACTION:
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
	err = s.etc.ProcessPts(s.points)
	s.points = nil
	return err
}
