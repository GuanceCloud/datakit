package kubernetes

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

func TestAAA(t *testing.T) {

	const (
		// kubePath      = "172.16.2.41:30443"
		kubePath      = "172.16.2.41:6443"
		ca            = "/run/secrets/kubernetes.io/serviceaccount/pki/ca.crt"
		cert          = "/run/secrets/kubernetes.io/serviceaccount/pki/apiserver-kubelet-client.crt"
		key           = "/run/secrets/kubernetes.io/serviceaccount/pki/apiserver-kubelet-client.key"
		insercureSkip = false
	)

	tlsconfig := &tls.ClientConfig{
		TLSCA:              ca,
		TLSCert:            cert,
		TLSKey:             key,
		InsecureSkipVerify: insercureSkip,
	}

	conf := createConfigByCert(kubePath, tlsconfig)
	cli, err := newClient(conf, time.Second*5)
	if err != nil {
		t.Fatal(err)
	}

	// p := pods{client: cli}
	// p.Gather()

	list, err := cli.getClusters()
	if err != nil {
		t.Error(err)
	}

	for _, item := range list.Items {
		t.Logf("%#v\n\n", item)
	}
}
