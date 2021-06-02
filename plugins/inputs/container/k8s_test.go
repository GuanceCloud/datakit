package container

import (
	"fmt"
	"testing"
)

func TestK8s(t *testing.T) {
	const kubeletURL = "http://127.0.0.1:10255"

	k := Kubernetes{
		URL:               kubeletURL,
		BearerTokenString: "NULL",
	}
	if err := k.Init(); err != nil {
		t.Fatal(err)
	}

	summaryApi, err := k.GetSummaryMetrics()
	if err != nil {
		t.Fatal(err)
	}

	pt, err := buildNodeMetrics(summaryApi)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%s\n\n", pt.String())

	pts, err := buildPodMetrics(summaryApi)
	if err != nil {
		t.Fatal(err)
	}

	for _, pt := range pts {
		fmt.Printf("%s\n\n", pt.String())
	}
}
