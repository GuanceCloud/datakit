package external

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
	"github.com/influxdata/telegraf"
	"google.golang.org/grpc"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	pluginName           = "external"
	externalConfigSample = `
# [external]
# [[external.servers]]
# listen = "/usr/local/cloudcare/DataFlux/datakit/datakit.sock" # domain socket path`
)

var (
	ErrInfluxParsePointFailed = errors.New("parse influx points failed")
)

type Config struct {
	Servers []*Server `toml:"Servers"`
}

type Server struct {
	DataKitServer
	Listen    string
	rpcServer *grpc.Server
}

func init() {
	inputs.Add(pluginName, func() inputs.Input {
		x := &Server{}
		return x
	})
}

func (s *Server) Feed(ctx context.Context, req *Request) (*Response, error) {

	resp := &Response{}

	if req.Lines != nil {
		pts, err := influxm.ParsePointsWithPrecision(req.Lines, time.Now().UTC(), req.Precision)
		if err != nil {
			return &Response{Err: err.Error()}, nil
		}

		log.Printf("[S] received %d points", len(pts))
		io.Feed(req.Lines, io.Metric)
		resp.Points = int64(len(pts))
	}

	if req.Objects != nil {
		// TODO
		// XXX: check if valid objects
	}

	return resp, nil
}

// inputs methods
func (s *Server) Catalog() string {
	return pluginName // XXX: use plugin name as the catalog
}

func (s *Server) SampleConfig() string {
	return externalConfigSample
}

func (s *Server) Description() string {
	return "Accept external line protocol from socket."
}

func (s *Server) Gather(telegraf.Accumulator) error { return nil }

var (
	rpcListener net.Listener
)

func (s *Server) Start() {

	if _, err := os.Stat(s.Listen); err == nil {
		if err := os.Remove(s.Listen); err != nil {
			panic(err)
		}
	}

	var err error

	rpcListener, err = net.Listen("unix", s.Listen)
	if err != nil {
		panic(err)
	}

	s.rpcServer = grpc.NewServer()
	RegisterDataKitServer(s.rpcServer, s)
	if err := s.rpcServer.Serve(rpcListener); err != nil {
		panic(err)
	}
}

func (s *Server) Stop() {
	s.rpcServer.Stop()
}
