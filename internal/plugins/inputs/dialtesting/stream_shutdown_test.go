// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestStreamWatcherShutdownCancelsRequest(t *testing.T) {
	// Other input tests can leave panic-recovery goroutines reading Exit.
	// Isolate replacement of this process-wide signal from those goroutines.
	if os.Getenv("DATAKIT_TEST_STREAM_SHUTDOWN") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestStreamWatcherShutdownCancelsRequest$", "-test.timeout=60s")
		cmd.Env = append(os.Environ(), "DATAKIT_TEST_STREAM_SHUTDOWN=1")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s", output)
		return
	}

	for _, trigger := range []string{"global_exit", "input_stop"} {
		for _, mode := range []string{"response_headers", "silent_stream", "heartbeat_stream"} {
			t.Run(trigger+"/"+mode, func(t *testing.T) {
				oldExit := datakit.Exit
				datakit.Exit = cliutils.NewSem()
				defer func() { datakit.Exit = oldExit }()

				started := make(chan struct{})
				requestDone := make(chan struct{})
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer close(requestDone)
					if mode != "response_headers" {
						w.Header().Set("Content-Type", "text/event-stream")
						w.(http.Flusher).Flush()
					}
					close(started)
					if mode != "heartbeat_stream" {
						<-r.Context().Done()
						return
					}
					ticker := time.NewTicker(10 * time.Millisecond)
					defer ticker.Stop()
					for {
						select {
						case <-r.Context().Done():
							return
						case <-ticker.C:
							if _, err := w.Write([]byte(": heartbeat\n\n")); err != nil {
								return
							}
							w.(http.Flusher).Flush()
						}
					}
				}))
				defer server.Close()

				ipt := defaultInput()
				ipt.Server = server.URL
				ipt.RegionID = "shutdown-test"
				ipt.cli = server.Client()
				ipt.isServerMode = true
				ipt.startStreamWatcher()
				// Explicit cancellation also releases the server if an assertion fails.
				defer ipt.stopStreamWatcher()
				ipt.streamWatchMu.Lock()
				done := ipt.streamWatchDone
				ipt.streamWatchMu.Unlock()
				require.NotNil(t, done)
				select {
				case <-started:
				case <-time.After(3 * time.Second):
					t.Fatal("stream request did not reach the server")
				}

				if trigger == "global_exit" {
					datakit.Exit.Close()
				} else {
					// Closing the input signal must work without calling Terminate,
					// which would otherwise explicitly cancel the watcher first.
					ipt.semStop.Close()
				}
				select {
				case <-done:
				case <-time.After(3 * time.Second):
					t.Fatal("stream watcher did not exit after shutdown")
				}
				select {
				case <-requestDone:
				case <-time.After(3 * time.Second):
					t.Fatal("shutdown did not cancel the HTTP request")
				}
			})
		}
	}
}
