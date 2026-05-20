package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCopyHeader(t *testing.T) {
	src := http.Header{}
	src.Add("Content-Type", "application/json")
	src.Add("X-Custom", "value1")
	src.Add("X-Custom", "value2")

	dst := http.Header{}
	copyHeader(dst, src)

	if dst.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", dst.Get("Content-Type"))
	}
	if vals := dst["X-Custom"]; len(vals) != 2 {
		t.Errorf("expected 2 X-Custom values, got %d", len(vals))
	}
}

func TestCopyHeaderEmpty(t *testing.T) {
	dst := http.Header{}
	copyHeader(dst, http.Header{})
	if len(dst) != 0 {
		t.Errorf("expected empty dst, got %v", dst)
	}
}

func TestProcessURL(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Upstream", "yes")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello from upstream"))
	}))
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodGet, upstream.URL, nil)
	rr := httptest.NewRecorder()

	processURL(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "hello from upstream" {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
	if rr.Header().Get("X-Upstream") != "yes" {
		t.Errorf("expected X-Upstream header to be forwarded")
	}
}

func TestProcessURLNon200(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodGet, upstream.URL, nil)
	rr := httptest.NewRecorder()

	processURL(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
	if rr.Body.String() != "not found" {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
}
