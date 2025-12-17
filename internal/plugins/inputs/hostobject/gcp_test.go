// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var gcpInstance = &gcp{}

func testGetGCP() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for GCP metadata flavor header
		if r.Header.Get("Metadata-Flavor") != "Google" {
			http.Error(w, "Missing Metadata-Flavor header", http.StatusForbidden)
			return
		}

		switch r.URL.Path {
		case "/computeMetadata/v1/instance/id":
			fmt.Fprint(w, "1234567890123456789")
		case "/computeMetadata/v1/instance/name":
			fmt.Fprint(w, "gcp-test-instance")
		case "/computeMetadata/v1/instance/machine-type":
			fmt.Fprint(w, "projects/123456789/machineTypes/n1-standard-1")
		case "/computeMetadata/v1/instance/zone":
			fmt.Fprint(w, "projects/123456789/zones/us-central1-a")
		case "/computeMetadata/v1/instance/network-interfaces/0/ip":
			fmt.Fprint(w, "10.128.0.2")
		case "/computeMetadata/v1/project/project-id":
			fmt.Fprint(w, "my-gcp-project")
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
	gcpInstance.baseURL = ts.URL + "/computeMetadata/v1"
	return ts
}

func TestGCP_InstanceID(t *testing.T) {
	ts := testGetGCP()
	defer ts.Close()

	id := gcpInstance.InstanceID()
	assert.Equal(t, "1234567890123456789", id, "GCP_InstanceID")
}

func TestGCP_InstanceName(t *testing.T) {
	ts := testGetGCP()
	defer ts.Close()

	name := gcpInstance.InstanceName()
	assert.Equal(t, "gcp-test-instance", name, "GCP_InstanceName")
}

func TestGCP_InstanceType(t *testing.T) {
	ts := testGetGCP()
	defer ts.Close()

	instanceType := gcpInstance.InstanceType()
	assert.Equal(t, "n1-standard-1", instanceType, "GCP_InstanceType")
}

func TestGCP_ZoneID(t *testing.T) {
	ts := testGetGCP()
	defer ts.Close()

	zoneID := gcpInstance.ZoneID()
	assert.Equal(t, "us-central1-a", zoneID, "GCP_ZoneID")
}

func TestGCP_Region(t *testing.T) {
	ts := testGetGCP()
	defer ts.Close()

	region := gcpInstance.Region()
	assert.Equal(t, "us-central1", region, "GCP_Region")
}

func TestGCP_PrivateIP(t *testing.T) {
	ts := testGetGCP()
	defer ts.Close()

	privateIP := gcpInstance.PrivateIP()
	assert.Equal(t, "10.128.0.2", privateIP, "GCP_PrivateIP")
}

func TestGCP_ProjectID(t *testing.T) {
	ts := testGetGCP()
	defer ts.Close()

	projectID := gcpInstance.ProjectID()
	assert.Equal(t, "my-gcp-project", projectID, "GCP_ProjectID")
}

func TestGCP_UnsupportedFields(t *testing.T) {
	// These fields are not supported by GCP and should return Unavailable
	assert.Equal(t, Unavailable, gcpInstance.Description(), "GCP_Description")
	assert.Equal(t, Unavailable, gcpInstance.InstanceChargeType(), "GCP_InstanceChargeType")
	assert.Equal(t, Unavailable, gcpInstance.InstanceNetworkType(), "GCP_InstanceNetworkType")
	assert.Equal(t, Unavailable, gcpInstance.InstanceStatus(), "GCP_InstanceStatus")
	assert.Equal(t, Unavailable, gcpInstance.SecurityGroupID(), "GCP_SecurityGroupID")
}

func TestGCP_WithoutHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't check header, just return error
		http.Error(w, "Missing Metadata-Flavor header", http.StatusForbidden)
	}))
	defer ts.Close()

	gcpInstance.baseURL = ts.URL + "/computeMetadata/v1"
	result := metadataGetByHeaderGCP(ts.URL + "/computeMetadata/v1/instance/id")
	assert.Equal(t, Unavailable, result, "GCP_WithoutHeader")
}

func TestGCP_Sync(t *testing.T) {
	ts := testGetGCP()
	defer ts.Close()

	result, err := gcpInstance.Sync()
	assert.NoError(t, err, "GCP_Sync should not return error")
	assert.NotNil(t, result, "GCP_Sync result should not be nil")

	assert.Equal(t, "gcp", result["cloud_provider"], "GCP_Sync cloud_provider")
	assert.Equal(t, "1234567890123456789", result["instance_id"], "GCP_Sync instance_id")
	assert.Equal(t, "gcp-test-instance", result["instance_name"], "GCP_Sync instance_name")
	assert.Equal(t, "n1-standard-1", result["instance_type"], "GCP_Sync instance_type")
	assert.Equal(t, "us-central1-a", result["zone_id"], "GCP_Sync zone_id")
	assert.Equal(t, "us-central1", result["region"], "GCP_Sync region")
	assert.Equal(t, "10.128.0.2", result["private_ip"], "GCP_Sync private_ip")
	assert.Equal(t, "my-gcp-project", result["project_id"], "GCP_Sync project_id")

	// Unsupported fields should be Unavailable
	assert.Equal(t, Unavailable, result["description"], "GCP_Sync description")
	assert.Equal(t, Unavailable, result["instance_charge_type"], "GCP_Sync instance_charge_type")
	assert.Equal(t, Unavailable, result["instance_network_type"], "GCP_Sync instance_network_type")
	assert.Equal(t, Unavailable, result["instance_status"], "GCP_Sync instance_status")
	assert.Equal(t, Unavailable, result["security_group_id"], "GCP_Sync security_group_id")
}

func TestGCP_RegionFromDifferentZones(t *testing.T) {
	testCases := []struct {
		name     string
		zone     string
		expected string
	}{
		{"us-central1-a", "projects/123456789/zones/us-central1-a", "us-central1"},
		{"us-east1-b", "projects/123456789/zones/us-east1-b", "us-east1"},
		{"asia-east1-c", "projects/123456789/zones/asia-east1-c", "asia-east1"},
		{"europe-west1-a", "projects/123456789/zones/europe-west1-a", "europe-west1"},
		{"us-west2-a", "projects/123456789/zones/us-west2-a", "us-west2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Metadata-Flavor") != "Google" {
					http.Error(w, "Missing Metadata-Flavor header", http.StatusForbidden)
					return
				}
				if r.URL.Path == "/computeMetadata/v1/instance/zone" {
					fmt.Fprint(w, tc.zone)
				}
			}))
			defer ts.Close()

			gcpInstance.baseURL = ts.URL + "/computeMetadata/v1"
			region := gcpInstance.Region()
			assert.Equal(t, tc.expected, region, "GCP_RegionFromZone: zone=%s", tc.zone)
		})
	}
}

func TestGCP_InstanceTypeParsing(t *testing.T) {
	testCases := []struct {
		name        string
		machineType string
		expected    string
	}{
		{"standard", "projects/123456789/machineTypes/n1-standard-1", "n1-standard-1"},
		{"custom", "projects/123456789/machineTypes/custom-4-8192", "custom-4-8192"},
		{"e2", "projects/123456789/machineTypes/e2-medium", "e2-medium"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Metadata-Flavor") != "Google" {
					http.Error(w, "Missing Metadata-Flavor header", http.StatusForbidden)
					return
				}
				if r.URL.Path == "/computeMetadata/v1/instance/machine-type" {
					fmt.Fprint(w, tc.machineType)
				}
			}))
			defer ts.Close()

			gcpInstance.baseURL = ts.URL + "/computeMetadata/v1"
			instanceType := gcpInstance.InstanceType()
			assert.Equal(t, tc.expected, instanceType, "GCP_InstanceType: machineType=%s", tc.machineType)
		})
	}
}
