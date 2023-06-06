// Copyright 2022 The jackal Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpServer

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HttpServer struct {
	port     int
	srv      *http.Server
	mux      *http.ServeMux
	logger   kitlog.Logger
	handlers map[string]http.HandlerFunc
}

func NewHTTPServer(port int, logger kitlog.Logger) *HttpServer {
	return &HttpServer{port: port, logger: logger, handlers: make(map[string]http.HandlerFunc)}
}

func (h *HttpServer) Register(endpoint string, handler func(http.ResponseWriter, *http.Request)) {
	h.handlers[endpoint] = http.HandlerFunc(handler)
}

func (h *HttpServer) Start(_ context.Context) error {
	h.mux = http.NewServeMux()
	h.mux.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{EnableOpenMetrics: true},
	))
	h.mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	h.mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	h.mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	h.mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	h.mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	h.mux.Handle("/healthz", http.HandlerFunc(h.healthCheck))

	level.Debug(h.logger).Log("msg", fmt.Sprintf("registered %d handlers", len(h.handlers)))
	for k, v := range h.handlers {
		level.Debug(h.logger).Log("msg", "new handler ", v)
		h.mux.Handle(k, v)
	}

	h.srv = &http.Server{Handler: h.mux}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", h.port))
	if err != nil {
		return err
	}
	go func() {
		if err := h.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			level.Error(h.logger).Log("msg", "failed to serve HTTP", "err", err)
		}
	}()
	level.Info(h.logger).Log("msg", "HTTP server listening", "port", h.port)
	return nil
}

func (h *HttpServer) Stop(ctx context.Context) error {
	if err := h.srv.Shutdown(ctx); err != nil {
		return err
	}
	level.Info(h.logger).Log("msg", "closed HTTP server", "port", h.port)
	return nil
}

func (h *HttpServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	w.WriteHeader(http.StatusOK)
	return
}
