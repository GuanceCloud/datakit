package kubernetes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

const body = `
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 15143
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
`

func TestPod(t *testing.T) {
	io.SetTest()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	var cases = []struct {
		podList *corev1.PodList
	}{
		{
			podList: &corev1.PodList{
				Items: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_name",
							Namespace: "test_namespace",
							Annotations: map[string]string{
								"datakit/prom.instances": `
									[[inputs.prom]]
									  url = "$IP"
									  source = "prom"
									  metric_types = ["counter", "gauge"]
									  measurement_prefix = ""
									  interval = "10s"
									  [inputs.prom.tags]
									  # some_tag = "some_value"
									  # more_tag = "some_other_value"
									  `,
							},
						},
						Status: corev1.PodStatus{
							PodIP: ts.URL,
						},
					},
				},
			},
		},
	}

	p := &pod{discovery: NewDiscovery()}
	done := make(chan struct{})

	for _, tc := range cases {
		p.run(tc.podList)
	}

	<-done

}
