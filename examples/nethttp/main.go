package main

import (
	"net/http"

	"github.com/sebest/xff"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from " + r.RemoteAddr + "\n"))
	})

	xffmw, _ := xff.Default()
	http.ListenAndServe(":3000", xffmw.Handler(handler))
}
