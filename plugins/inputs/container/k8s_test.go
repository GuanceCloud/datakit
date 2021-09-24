package container

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestGetContainerPodName(t *testing.T) {
	fakeResp := `
{
	"kind": "not-set",
	"apiVersion": "not-set",
	"items": [
		{
			"metadata": {
				"name": "abc123",
				"namespace": "not-set",
				"uid": "id-not-set",
				"labels": null
			},
			"status": {
				"phase": "not-set",
				"startTime": "not-set",
				"containerStatuses": [
					{
						"containerID": "docker://faked-container-id",
						"restartCount": 0,
						"ready": true
					}
				]
			}
		}
	]
}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, fakeResp)
	}))

	defer ts.Close()

	k := Kubernetes{
		URL: ts.URL,
	}

	if err := k.Init(); err != nil {
		t.Error(err)
	}

	cases := []struct {
		id, expected string
	}{
		{
			id:       "faked-container-id",
			expected: "abc123",
		},
	}

	for _, tc := range cases {
		name, err := k.GetContainerPodName(containerIDPrefix + tc.id)
		if err != nil {
			t.Error(err)
		}

		tu.Equals(t, tc.expected, name)
	}
}

func TestGetDeploymentFromPodName(t *testing.T) {
	cases := []struct {
		podName, deploymentName string
	}{
		{
			"corestone-76b5fb8bd-lbxc6",
			"corestone",
		},
		{
			"nsqd-7c49ff9c77-w85mb",
			"nsqd",
		},
		{
			"kodo-inner-5df4fb4897-csqdz",
			"kodo-inner",
		},
		{
			"invalid-12345678",
			"invalid-12345678",
		},
		{
			"invalid",
			"invalid",
		},
	}

	for _, tc := range cases {
		output := getDeploymentFromPodName(tc.podName)

		tu.Assert(t, output == tc.deploymentName,
			"\nexpect: %s\n   got: %s",
			tc.deploymentName, output)
	}
}
