//go:build ignore
// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":18080", "listen address")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body := fmt.Sprintf("ok %s\n", r.URL.RequestURI())
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		_, _ = w.Write([]byte(body))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok\n"))
	})

	srv := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}
	log.Fatal(srv.ListenAndServe())
}
