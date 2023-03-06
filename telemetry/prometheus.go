package telemetry

import (
	"fmt"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

type PrometheusMetrics struct {
	isRunning bool
}

func NewPrometheusMetrics() *PrometheusMetrics {

	return &PrometheusMetrics{isRunning: false}
}

func (metrics *PrometheusMetrics) AddServerOptions(opts []grpc.ServerOption) []grpc.ServerOption {

	opts = append(opts, grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor))
	opts = append(opts, grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor))

	return opts
}

func (metrics *PrometheusMetrics) EnableMetrics(server *grpc.Server) {

	grpc_prometheus.Register(server)

	promRegistry := prometheus.NewRegistry()

	grpcMetrics := grpc_prometheus.NewServerMetrics()
	promRegistry.MustRegister(grpcMetrics)
	grpcMetrics.InitializeMetrics(server)

	const promServerPort = 9090

	promHttpServer := &http.Server{
		Handler: promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{}),
		Addr:    fmt.Sprintf("0.0.0.0:%d", promServerPort)}

	go func() {

		log.Printf("Prometheus server starting on port %d", promServerPort)

		if err := promHttpServer.ListenAndServe(); err != nil {
			log.Fatal("Unable to start Prometheus http server. Metrics will not be available")
		}
	}()
}
