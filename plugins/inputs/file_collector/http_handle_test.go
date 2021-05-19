package file_collector

import (
	"net/http"
	"testing"
)

func TestHandle(t *testing.T) {
	fileCollector = newfc()
	http.HandleFunc("/test", Handle)
	if err := http.ListenAndServe(":8888", nil); err != nil {
		l.Fatal(err)
	}

}
