package main

import (
	"fmt"
	"net/http"

	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "this is a blocked-response example")
	zlog.Debugf("request path %s", r.RequestURI)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
