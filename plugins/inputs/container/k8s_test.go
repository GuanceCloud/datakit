package container

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestGetContainerPodName(t *testing.T) {
	mock := Kubernetes{
		URL: "http://127.0.0.1:10255",
	}

	if err := mock.Init(); err != nil {
		t.Fatal(err)
	}

	var cases = []struct {
		id string
	}{
		{
			"cf063df157fed1976dfff6b47e7ba55a1ce36385715b859bd03d4acc9b92690c",
		},
	}

	for idx, tc := range cases {
		name, err := mock.GetContainerPodName(containerIDPrefix + tc.id)
		if err != nil {
			t.Error(err)
		}
		t.Logf("[%d] container_id:%s pod_name:%s\n", idx, tc.id, name)
	}
}

func TestGetDeploymentFromPodName(t *testing.T) {
	var cases = []struct {
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
