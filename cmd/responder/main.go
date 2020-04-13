package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type responderType struct {
	counterRequestsTotal int
}

func main() {
	responder := responderType{
		counterRequestsTotal: 0,
	}
	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	http.HandleFunc("/ip", responder.ipHandler)
	http.HandleFunc("/reset", responder.resetHandler)
	http.HandleFunc("/stats", responder.statsHandler)
	_ = srv.ListenAndServe()
}

func (r *responderType) ipHandler(w http.ResponseWriter, req *http.Request) {
	r.counterRequestsTotal++
	_, _ = fmt.Fprintf(w, strings.Split(req.RemoteAddr, ":")[0])
}

func (r *responderType) resetHandler(w http.ResponseWriter, req *http.Request) {
	r.counterRequestsTotal = 0
	_, _ = fmt.Fprintf(w, "ok")
}

func (r *responderType) statsHandler(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "Total: %d\n", r.counterRequestsTotal)
}
