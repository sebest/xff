package main

import (
	"net/http"

	"github.com/sebest/xff"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from " + r.RemoteAddr + "\n"))
	})

	http.ListenAndServe(":8080", xff.Handler(handler))
}
