# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
make build        # compiles server.go to ./build/server

# Run
make run          # runs the built binary

# Clean
make clean        # removes ./build/

# Test
go test -v        # run all tests
go test -run TestCopyHeader  # run a single test by name
```

## Architecture

This is a minimal single-file HTTP reverse proxy written in Go (`server.go`).

- **`processURL`** — the sole HTTP handler, registered at `/`. It forwards the incoming request URL as an outbound `http.Get`, then proxies the response status, headers, and body back to the client.
- **`copyHeader`** — helper that copies all header key/value pairs from one `http.Header` to another.
- The server listens on the port defined by the `PORT` environment variable, defaulting to `:19998`.

Tests (`server_test.go`) use `net/http/httptest` — a real fake upstream server is spun up per test, no mocks.
