package io

import (
	"context"
	"errors"
	"net"
	"os"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
	"google.golang.org/grpc"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

var (
	ErrInfluxParsePointFailed = errors.New("parse influx points failed")

	rpcListener net.Listener
)

type Server struct {
	DataKitServer

	rpcServer *grpc.Server
}

func (s *Server) Send(ctx context.Context, req *Request) (*Response, error) {

	resp := &Response{}

	if req.Lines != nil {
		pts, err := influxm.ParsePointsWithPrecision(req.Lines, time.Now().UTC(), req.Precision)
		if err != nil {
			return &Response{Err: err.Error()}, nil
		}

		l.Debugf("received %d points from %s", len(pts), req.Name)
		Feed(req.Lines, Metric)
		resp.Points = int64(len(pts))
	}

	if req.Objects != nil {
		// TODO
		// XXX: check if valid objects
	}

	return resp, nil
}

func GRPCServer() {

	if !inited {
		panic(initErr)
	}

	s := &Server{}

	l.Infof("gRPC starting...")

	if _, err := os.Stat(config.GRPCDomainSock); err == nil {
		if err := os.Remove(config.GRPCDomainSock); err != nil {
			panic(err)
		}
	}

	var err error

	rpcListener, err = net.Listen("unix", config.GRPCDomainSock)
	if err != nil {
		l.Errorf("start gRPC server failed: %s", err)
		return
	}

	l.Infof("start gRPC server on %s ok", config.GRPCDomainSock)

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
