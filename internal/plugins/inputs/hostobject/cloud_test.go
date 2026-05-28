// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	"github.com/stretchr/testify/assert"
)

func TestMetaGet(t *testing.T) {
	cases := []struct {
		body, expect string
	}{
		{
			body: `multi-
lin-
data`,
			expect: `multi- lin- data`,
		},

		{
			body:   `中文balabala`,
			expect: `中文balabala`,
		},

		{
			body:   `¡™£¢∞§¶•ªº–≠‘«“æ…÷≥≤`,
			expect: `¡™£¢∞§¶•ªº–≠‘«“æ…÷≥≤`,
		},

		{
			body:   `~!@#$%^&*()_+-=|}{\][":';?><,./`,
			expect: `~!@#$%^&*()_+-=|}{\][":';?><,./`,
		},

		{
			body:   `abc`,
			expect: `abc`,
		},
	}

	tags := models.Tags{models.NewTag([]byte("a"), []byte(`~!@#$%^&*()_+=-|}{\][":';?><,./`))}

	for _, tc := range cases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, tc.body)
		}))

		x := metaGet(ts.URL)

		assert.Equal(t, tc.expect, x)

		ts.Close()

		pt1, err := models.NewPoint("test", tags,
			map[string]interface{}{"extra_cloud_meta": x}, time.Now())
		if err != nil {
			t.Error(err)
		}

		pts, err := models.ParsePointsWithPrecision([]byte(pt1.String()), time.Now(), "n")
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, 1, len(pts))

		assert.Equal(t, pt1.String(), pts[0].String())

		t.Logf("pt: %s", pt1.String())
	}
}

const (
	AWSAuthHeader  = "X-aws-ec2-metadata-token"
	AWSTTLHeader   = "X-aws-ec2-metadata-token-ttl-seconds"
	AWSMaxTokenTTL = 21600 * time.Second
)

func TestMetaGetV2(t *testing.T) {
	expectToken := "test-token"
	t.Run("AWS IMDSv2 Success", func(t *testing.T) {
		tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			ttlHeader := r.Header.Get(AWSTTLHeader)
			if ttlHeader == "" {
				http.Error(w, "Missing TTL header", http.StatusBadRequest)
				return
			}
			fmt.Fprint(w, expectToken)
		}))
		defer tokenServer.Close()

		metaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get(AWSAuthHeader)
			if token != expectToken {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			fmt.Fprint(w, "meta-data-response")
		}))
		defer metaServer.Close()

		authConfig := AuthConfig{
			Enable:      true,
			TokenURL:    tokenServer.URL,
			AuthHeader:  AWSAuthHeader,
			TTLHeader:   AWSTTLHeader,
			MaxTokenTTL: AWSMaxTokenTTL,
		}

		res := metaGetV2(metaServer.URL, authConfig)
		assert.Equal(t, "meta-data-response", res)
	})

	t.Run("AWS IMDS 401", func(t *testing.T) {
		tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			ttlHeader := r.Header.Get(AWSTTLHeader)
			if ttlHeader == "" {
				http.Error(w, "Missing TTL header", http.StatusBadRequest)
				return
			}
			fmt.Fprint(w, "")
		}))
		defer tokenServer.Close()

		metaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get(AWSAuthHeader)
			if token != expectToken {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			fmt.Fprint(w, "meta-data-response")
		}))
		defer metaServer.Close()

		authConfig := AuthConfig{
			Enable:      true,
			TokenURL:    tokenServer.URL,
			AuthHeader:  AWSAuthHeader,
			TTLHeader:   AWSTTLHeader,
			MaxTokenTTL: AWSMaxTokenTTL,
		}

		res := metaGetV2(metaServer.URL, authConfig)
		assert.Equal(t, Unavailable, res)
	})
}

func TestKsyunSync(t *testing.T) {
	ts := newKsyunMetadataServer(t)
	defer ts.Close()

	got, err := (&ksyun{baseURL: ts.URL}).Sync()
	assert.NoError(t, err)

	assert.Equal(t, Ksyun, got["cloud_provider"])
	assert.Equal(t, "9554895b-968d-4905-a032-0b8f179230dc", got["instance_id"])
	assert.Equal(t, "vm10-67-0-136", got["instance_name"])
	assert.Equal(t, "KEC", got["instance_type"])
	assert.Equal(t, "10.67.0.136", got["private_ip"])
	assert.Equal(t, "cn-beijing-6", got["region"])
	assert.Equal(t, "cn-beijing-6a", got["zone_id"])
	assert.Equal(t,
		"4b579780-81bc-427c-aee9-bb02943d5a8e dc4c927a-b33d-4c23-b00b-9298580cef8d",
		got["security_group_id"])
	assert.Equal(t, Unavailable, got["description"])
	assert.Equal(t, Unavailable, got["instance_charge_type"])
	assert.Equal(t, Unavailable, got["instance_network_type"])
	assert.Equal(t, Unavailable, got["instance_status"])
	assert.Equal(t, Unavailable, got["project_id"])
}

func TestSyncCloudInfoKsyunWithCustomMetaURL(t *testing.T) {
	ts := newKsyunMetadataServer(t)
	defer ts.Close()

	ipt := defaultInput()
	ipt.CloudMetaURL = map[string]string{Ksyun: ts.URL}

	got, err := ipt.SyncCloudInfo(Ksyun)
	assert.NoError(t, err)

	assert.Equal(t, Ksyun, got["cloud_provider"])
	assert.Equal(t, "9554895b-968d-4905-a032-0b8f179230dc", got["instance_id"])
	assert.Equal(t, "cn-beijing-6", got["region"])
	assert.Equal(t, "cn-beijing-6a", got["zone_id"])
}

func TestMatchCloudProviderKsyunUsesSingleProbe(t *testing.T) {
	var (
		mu    sync.Mutex
		paths []string
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()

		if r.URL.Path == "/instance-id" {
			fmt.Fprint(w, "9554895b-968d-4905-a032-0b8f179230dc")
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	ipt := defaultInput()
	ipt.CloudMetaURL = map[string]string{Ksyun: ts.URL}

	assert.True(t, ipt.matchCloudProvider(Ksyun))

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []string{"/instance-id"}, paths)
}

func newKsyunMetadataServer(t *testing.T) *httptest.Server {
	t.Helper()

	responses := map[string]string{
		"/instance-id":       "9554895b-968d-4905-a032-0b8f179230dc",
		"/hostname":          "vm10-67-0-136",
		"/host-type":         "KEC",
		"/local-ipv4":        "10.67.0.136",
		"/placement/region":  "cn-beijing-6",
		"/placement/zone":    "cn-beijing-6a",
		"/securitygroup-ids": "4b579780-81bc-427c-aee9-bb02943d5a8e\ndc4c927a-b33d-4c23-b00b-9298580cef8d",
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, ok := responses[r.URL.Path]; ok {
			fmt.Fprint(w, body)
			return
		}

		http.NotFound(w, r)
	}))
}

// func TestSyncCloudInfo(t *testing.T) {
// 	ipt := defaultInput()
// 	ipt.EnableCloudAWSIMDSv2 = true

// 	res, _ := ipt.SyncCloudInfo(AWS)
// 	fmt.Println(res)
// }

// func handleTokenRequest(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPut {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	ttlStr := r.Header.Get(AWSTTLHeader)
// 	if ttlStr == "" {
// 		http.Error(w, "Missing X-aws-ec2-metadata-token-ttl-seconds header", http.StatusBadRequest)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("test-token"))
// }

// func handleMetaDataRequest(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	token := r.Header.Get(AWSAuthHeader)
// 	if token == "" {
// 		http.Error(w, "Missing X-aws-ec2-metadata-token header", http.StatusBadRequest)
// 		return
// 	}

// 	if token != "test-token" {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("success"))
// }

// func TestAWSServer(t *testing.T) {
// 	addr := "http://127.0.0.1:7654"
// 	// http.HandleFunc("/latest/api/token", handleTokenRequest)
// 	http.HandleFunc("/latest/meta", handleMetaDataRequest)

// 	log.Println(addr)
// 	if err := http.ListenAndServe(":7654", nil); err != nil {
// 		log.Fatalf("Failed to start server: %v", err)
// 	}
// 	fmt.Println("test")
// }
