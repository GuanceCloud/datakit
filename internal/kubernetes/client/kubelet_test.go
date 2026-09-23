// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	statsv1alpha1 "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

func TestKubeletStatsCancellation(t *testing.T) {
	for _, flushHeaders := range []bool{false, true} {
		t.Run(map[bool]string{false: "headers", true: "body"}[flushHeaders], func(t *testing.T) {
			started := make(chan struct{})
			release := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if flushHeaders {
					writer.WriteHeader(http.StatusOK)
					writer.(http.Flusher).Flush()
				}
				close(started)
				select {
				case <-request.Context().Done():
				case <-release:
				}
			}))
			defer server.Close()
			defer close(release)
			client, err := NewKubeletClient(&rest.Config{}, "http", strings.TrimPrefix(server.URL, "http://"))
			require.NoError(t, err)
			provider, ok := client.(interface {
				GetStatsSummaryWithContext(context.Context) (*statsv1alpha1.Summary, error)
			})
			require.True(t, ok, "kubelet stats must support cancellation")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			finished := make(chan error, 1)
			go func() {
				_, requestErr := provider.GetStatsSummaryWithContext(ctx)
				finished <- requestErr
			}()
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("request did not start")
			}
			cancel()
			select {
			case err := <-finished:
				require.True(t, errors.Is(err, context.Canceled), "%v", err)
			case <-time.After(time.Second):
				t.Fatal("cancellation did not interrupt the stats request")
			}
		})
	}
}
