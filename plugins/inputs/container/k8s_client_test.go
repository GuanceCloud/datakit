package container

/* failure of testing
func TestNewClient(t *testing.T) {
	var (
		kubeURL     = "172.16.2.41:6443"
		bearerToken = os.Getenv("K8S_TOKEN")
	)

	cli, err := newK8sClientFromBearerTokenString(kubeURL, bearerToken)
	if err != nil {
		t.Fatal(err)
	}

	list, err := cli.getPods().List(context.Background(), metaV1ListOption)
	if err != nil {
		t.Error(err)
	}
	for _, item := range list.Items {
		s, _ := json.MarshalIndent(item, "", "    ")
		t.Logf("%s\n\n", s)
	}
}
*/
