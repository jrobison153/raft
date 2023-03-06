package telemetry

import (
	"google.golang.org/grpc"
	"testing"
)

func TestWhenAddingServerOptionsThenGrpcStreamInterceptorsAreAdded(t *testing.T) {

	metrics := NewPrometheusMetrics()

	var opts []grpc.ServerOption
	opts = metrics.AddServerOptions(opts)

	if len(opts) != 2 {
		t.Errorf("There should have been 2 ServerOption added but %d were added",
			len(opts))
	}
}
