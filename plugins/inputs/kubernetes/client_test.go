package kubernetes

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

func TestNewClient(t *testing.T) {
	const (
		kubeURL            = "172.16.2.41:6443"
		ca                 = "/run/secrets/kubernetes.io/serviceaccount/pki/ca.crt"
		cert               = "/run/secrets/kubernetes.io/serviceaccount/pki/apiserver-kubelet-client.crt"
		key                = "/run/secrets/kubernetes.io/serviceaccount/pki/apiserver-kubelet-client.key"
		insecureSkipVerify = false
		bearerTokenPath    = ""
		bearerToken        = ""
	)

	tlsconfig := net.TlsClientConfig{
		CaCerts:            []string{ca},
		Cert:               cert,
		CertKey:            key,
		InsecureSkipVerify: insecureSkipVerify,
	}

	cli, err := newClientFromTLS(kubeURL, &tlsconfig)
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
