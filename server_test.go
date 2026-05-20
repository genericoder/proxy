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

// TestProcessURL_UpstreamUnreachable verifies that when the upstream host cannot
// be reached, processURL returns a 502 Bad Gateway instead of calling os.Exit.
func TestProcessURL_UpstreamUnreachable(t *testing.T) {
	// Use a URL that will never succeed — nothing is listening on this port.
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1", nil)
	rr := httptest.NewRecorder()

	processURL(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rr.Code)
	}
}

// TestProcessURL_HopByHopHeadersNotForwarded verifies that hop-by-hop headers
// (e.g. Transfer-Encoding) set by the upstream are not propagated to the client.
// Go's HTTP client automatically strips hop-by-hop headers before they reach
// response.Header, so the proxy must not forward them even if copyHeader is naive.
func TestProcessURL_HopByHopHeadersNotForwarded(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Attempt to set a hop-by-hop header; Go's HTTP layer will strip it
		// from the response seen by http.Get on the client side.
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("X-Real-Header", "should-pass")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("body"))
	}))
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodGet, upstream.URL, nil)
	rr := httptest.NewRecorder()

	processURL(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	// Transfer-Encoding is a hop-by-hop header; Go's HTTP client strips it,
	// so it must not appear in the proxied response.
	if te := rr.Header().Get("Transfer-Encoding"); te != "" {
		t.Errorf("hop-by-hop Transfer-Encoding must not be forwarded to client, got %q", te)
	}
	// Non-hop-by-hop headers must still be forwarded.
	if rr.Header().Get("X-Real-Header") != "should-pass" {
		t.Errorf("expected X-Real-Header to be forwarded")
	}
}

// TestProcessURL_MultipleResponseHeadersForwarded verifies that all non-hop-by-hop
// response headers from the upstream — including multi-value headers — are
// correctly proxied to the client.
func TestProcessURL_MultipleResponseHeadersForwarded(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("X-Service", "proxy-test")
		// Add multiple values for the same header key.
		w.Header().Add("X-Trace-Id", "abc123")
		w.Header().Add("X-Trace-Id", "def456")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodGet, upstream.URL, nil)
	rr := httptest.NewRecorder()

	processURL(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("expected Content-Type text/plain, got %q", rr.Header().Get("Content-Type"))
	}
	if rr.Header().Get("X-Service") != "proxy-test" {
		t.Errorf("expected X-Service proxy-test, got %q", rr.Header().Get("X-Service"))
	}
	// Both values of the multi-value header must be forwarded.
	traceIDs := rr.Header()["X-Trace-Id"]
	if len(traceIDs) != 2 {
		t.Errorf("expected 2 X-Trace-Id values, got %d: %v", len(traceIDs), traceIDs)
	}
}
