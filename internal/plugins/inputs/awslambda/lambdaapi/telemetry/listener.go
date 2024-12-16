// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/consts"
)

// Listener Used to listen to the Telemetry API.
type Listener struct {
	httpServer *http.Server
	eventsChan chan []*Event
}

func NewTelemetryListener() *Listener {
	return &Listener{
		httpServer: nil,
		eventsChan: make(chan []*Event, 1000),
	}
}

func (s *Listener) GetPullChan() <-chan []*Event {
	return s.eventsChan
}

func listenOnAddress() string {
	envAwsLocal, ok := os.LookupEnv("AWS_SAM_LOCAL")
	var addr string
	if ok && envAwsLocal == "true" {
		addr = ":" + consts.TelemetryPort
	} else {
		addr = "sandbox:" + consts.TelemetryPort
	}

	return addr
}

// Start Starts the server in a goroutine where the log events will be sent.
func (s *Listener) Start() (string, error) {
	address := listenOnAddress()
	l.Infof("[listener:Start] Starting on address: %q", address)
	s.httpServer = &http.Server{Addr: address}
	http.HandleFunc("/", s.Handler)
	go func() {
		err := s.httpServer.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			l.Error("[listener:goroutine] Unexpected stop on Http Server:", err)
			s.Shutdown()
		} else {
			l.Info("[listener:goroutine] Http Server closed:", err)
		}
	}()
	return fmt.Sprintf("http://%s/", address), nil
}

func (s *Listener) Handler(w http.ResponseWriter, r *http.Request) {
	err := s.HandlerTelemetry(w, r)
	if err != nil {
		return
	}
}

// HandlerTelemetry handles the requests coming from the Telemetry API.
// Everytime Telemetry API sends log events, this function will read them from the response body
// and put into a synchronous queue to be dispatched later.
// Logging or printing besides the error cases below is not recommended if you have subscribed to
// receive extension logs. Otherwise, logging here will cause Telemetry API to send new logs for
// the printed lines which may create an infinite loop.
func (s *Listener) HandlerTelemetry(_ http.ResponseWriter, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		l.Error("[listener:http_handler] Error reading body:", err)
		return err
	}

	if l.Level() <= zap.DebugLevel {
		l.Debug("telemetry body\n", string(body))
	}

	// Parse and put the log messages into the queue.
	var events []*Event
	err = json.Unmarshal(body, &events)
	if err != nil {
		l.Error("[listener:http_handler] Error unmarshalling body(%q):", string(body), err)
		return err
	}

	l.Debugf("send %d events to eventsChan", len(events))
	s.eventsChan <- events

	return nil
}

// Shutdown Terminates the HTTP server listening for logs.
func (s *Listener) Shutdown() {
	l.Info("[listener:http_handler] Shutting down telemetry listener")
	if s.eventsChan != nil {
		close(s.eventsChan)
		s.eventsChan = nil
	}

	if s.httpServer != nil {
		err := s.httpServer.Shutdown(context.Background())
		if err != nil {
			l.Error("[listener:Shutdown] Failed to shutdown http server gracefully:", err)
		} else {
			s.httpServer = nil
		}
	}
}

func (s *Listener) WaitShutdown(ctx context.Context) {
	<-ctx.Done()
	s.Shutdown()
}
