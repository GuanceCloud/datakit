package external

import (
	"context"
	"errors"
	"net"
	"os"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
	"github.com/influxdata/telegraf"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	pluginName           = "external"
	externalConfigSample = `
# [external]
# listen = "/usr/local/cloudcare/DataFlux/datakit/datakit.sock" # domain socket path`
)

var (
	ErrInfluxParsePointFailed = errors.New("parse influx points failed")

	l *zap.SugaredLogger
)

type external struct {
	Listen string `toml:"listen"`
}

type Server struct {
	DataKitServer

	External *external `toml:"external"`

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

		l.Debugf("received %d points", len(pts))
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

func (s *Server) Interval() time.Duration {
	return time.Duration(0)
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

func (s *Server) Run() {
	l = logger.SLogger("external")
	l.Infof("gRPC external started...")

	if _, err := os.Stat(s.External.Listen); err == nil {
		if err := os.Remove(s.External.Listen); err != nil {
			panic(err)
		}
	}

	var err error

	rpcListener, err = net.Listen("unix", s.External.Listen)
	if err != nil {
		l.Errorf("start gRPC server failed: %s", err)
		return
	}

	l.Infof("start gRPC server on %s ok", s.External.Listen)

	s.rpcServer = grpc.NewServer()
	RegisterDataKitServer(s.rpcServer, s)

	go func() {
		if err := s.rpcServer.Serve(rpcListener); err != nil {
			l.Error(err)
		}

		l.Info("gRPC server exit")
	}()

	<-config.Exit.Wait()
	l.Info("stopping gRPC server...")
	s.rpcServer.Stop()

	l.Info("exit")
	return
}
