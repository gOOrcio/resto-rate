package utils

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/prometheus/client_golang/prometheus"
)

type RPCMetrics struct {
	reqs     *prometheus.CounterVec
	inflight *prometheus.GaugeVec
	latency  *prometheus.HistogramVec
}

func NewRPCMetrics(reg prometheus.Registerer) *RPCMetrics {
	m := &RPCMetrics{
		reqs: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rpc_requests_total",
				Help: "Total RPC requests by method and status.",
			},
			[]string{"method", "status"},
		),
		inflight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "rpc_inflight_requests",
				Help: "In-flight RPC requests by method.",
			},
			[]string{"method"},
		),
		latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "rpc_request_duration_seconds",
				Help:    "RPC request duration by method.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),
	}
	reg.MustRegister(m.reqs, m.inflight, m.latency)
	return m
}

func (m *RPCMetrics) ConnectInterceptor() connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			method := req.Spec().Procedure

			m.inflight.WithLabelValues(method).Inc()
			start := time.Now()
			resp, err := next(ctx, req)
			dur := time.Since(start).Seconds()
			m.latency.WithLabelValues(method).Observe(dur)

			status := "OK"
			if err != nil {
				status = connect.CodeOf(err).String()
			}
			m.reqs.WithLabelValues(method, status).Inc()
			m.inflight.WithLabelValues(method).Dec()

			return resp, err
		}
	})
}
