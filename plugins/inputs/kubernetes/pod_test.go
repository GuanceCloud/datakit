package kubernetes

import (
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

/* test: fail
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
	cases := []struct {
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
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_name",
							Namespace: "test_namespace",
							// 无效的 url 字段名
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

func TestComplatePromConfig(t *testing.T) {
	io.SetTest()

	tc := struct {
		podList    *corev1.PodList
		configList []string
	}{
		podList: &corev1.PodList{
			Items: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name_01",
						Namespace: "test_namespace_01",
						Annotations: map[string]string{
							"datakit/prom.instances": `
[[inputs.prom]]
  url = "$IP"
  source = "prom"
  metric_types = ["counter", "gauge"]
  measurement_prefix = ""
  interval = "10s"
  [inputs.prom.tags]
  name = "$PODNAME"
  namespace = "$NAMESPACE"
`,
							"datakit/prom.instances.ip_index": "1",
						},
					},
					Status: corev1.PodStatus{
						PodIP: "http://dummy_ip_01",
						PodIPs: []corev1.PodIP{
							{IP: "http://dummy_ip_01_index_00"},
							{IP: "http://dummy_ip_01_index_01"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name_02",
						Namespace: "test_namespace_02",
						Annotations: map[string]string{
							"datakit/prom.instances": `
[[inputs.prom]]
  url = "$IP"
  source = "prom"
  metric_types = ["counter", "gauge"]
  measurement_prefix = ""
  interval = "10s"
  [inputs.prom.tags]
  name = "$PODNAME"
  namespace = "$NAMESPACE"
`,
							"datakit/prom.instances.ip_index": "2",
						},
					},
					Status: corev1.PodStatus{
						PodIP: "http://dummy_ip_02",
						PodIPs: []corev1.PodIP{
							{IP: "http://dummy_ip_02_index_00"},
							{IP: "http://dummy_ip_02_index_01"},
							{IP: "http://dummy_ip_02_index_02"},
						},
					},
				},
				// fail
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name_03",
						Namespace: "test_namespace_03",
						Annotations: map[string]string{
							"datakit/prom.instances": `
[[inputs.prom]]
  url = "$IP"
  source = "prom"
  metric_types = ["counter", "gauge"]
  measurement_prefix = ""
  interval = "10s"
  [inputs.prom.tags]
  name = "$PODNAME"
  namespace = "$NAMESPACE"
`,
							"datakit/prom.instances.ip_index": "3",
						},
					},
					Status: corev1.PodStatus{
						PodIP: "http://dummy_ip_03",
						PodIPs: []corev1.PodIP{
							{IP: "http://dummy_ip_03_index_00"},
							{IP: "http://dummy_ip_03_index_01"},
							{IP: "http://dummy_ip_03_index_02"},
						},
					},
				},
			},
		},

		configList: []string{
			`
[[inputs.prom]]
  url = "http://dummy_ip_01_index_01"
  source = "prom"
  metric_types = ["counter", "gauge"]
  measurement_prefix = ""
  interval = "10s"
  [inputs.prom.tags]
  name = "test_name_01"
  namespace = "test_namespace_01"
`,
			`
[[inputs.prom]]
  url = "http://dummy_ip_02_index_02"
  source = "prom"
  metric_types = ["counter", "gauge"]
  measurement_prefix = ""
  interval = "10s"
  [inputs.prom.tags]
  name = "test_name_02"
  namespace = "test_namespace_02"
`,

			// fail
			"",
		},
	}

	for idx, obj := range tc.podList.Items {
		config, ok := obj.Annotations[annotationPromExport]
		if !ok {
			continue
		}

		config = complatePromConfig(config, &tc.podList.Items[idx])

		if (tc.configList[idx] != "") && !assert.Equal(t, config, tc.configList[idx]) {
			t.Fatal(config)
		}
	}
} */
