package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAppHealthy(t *testing.T) {
	client := &http.Client{Timeout: 200 * time.Millisecond}

	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ok.Close()
	if !appHealthy(client, ok.URL) {
		t.Error("200 response: want healthy")
	}

	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second)
	}))
	defer slow.Close()
	if appHealthy(client, slow.URL) {
		t.Error("timed-out response: want unhealthy")
	}

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bad.Close()
	if appHealthy(client, bad.URL) {
		t.Error("500 response: want unhealthy")
	}
}
