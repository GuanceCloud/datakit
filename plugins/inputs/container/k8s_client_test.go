// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

/*
import (
	"context"
	"encoding/json"
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewClient(t *testing.T) {
	var (
		kubeURL     = "1.15.22.133:6443"
		bearerToken = os.Getenv("K8S_TOKEN")
	)

	cli, err := newK8sClientFromBearerTokenString(kubeURL, bearerToken)
	if err != nil {
		t.Fatal(err)
	}

	// metav1.ListOptions{LabelSelector: "app=nginx"}
	list, err := cli.getDataKits().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Error(err)
	}

	for _, item := range list.Items {
		s, _ := json.MarshalIndent(item, "", "    ")
		t.Logf("%s\n\n", s)
	}
}
*/
