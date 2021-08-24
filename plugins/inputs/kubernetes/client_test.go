package kubernetes

import (
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	var (
		kubeURL     = "172.16.2.41:6443"
		bearerToken = os.Getenv("K8S_TOKEN")
	)

	cli, err := newClientFromBearerTokenString(kubeURL, bearerToken)
	if err != nil {
		t.Fatal(err)
	}

	list, err := cli.getPods()
	if err != nil {
		t.Error(err)
	}

	for _, item := range list.Items {
		t.Logf("%#v\n\n", item)
	}
}
