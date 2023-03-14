// Package grpc provides various controllers that allow clients to communicate with the Raft
// nodes
package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/jrobison153/raft/api"
	"github.com/jrobison153/raft/policy/client"
	"github.com/jrobison153/raft/server"
	"github.com/jrobison153/raft/telemetry"
	"google.golang.org/grpc"
	"log"
	"net"
)

// RaftServer serves the client API via the gRPC protocol
type RaftServer struct {
	api.PersisterServer
	server.LifeCycler

	server       *grpc.Server
	policyClient client.Persister
}

func New(policyClient client.Persister) *RaftServer {

	return &RaftServer{
		policyClient: policyClient,
	}
}

var (
	ErrPutRetryable      = errors.New("unable to save data, the operation is safe to retry")
	ErrGetGeneralFailure = errors.New("unable to get data")
	ErrGetEmptyKey       = errors.New("invalid key, key must not be the empty string")
)

// Start starts the gRPC server listening on the specified port. Any error encountered during start
// will be fatal.
func (clientApi *RaftServer) Start(port uint32) {

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	metrics := telemetry.NewPrometheusMetrics()

	var opts []grpc.ServerOption
	opts = metrics.AddServerOptions(opts)

	grpcServer := grpc.NewServer(opts...)

	clientApi.server = grpcServer

	api.RegisterPersisterServer(grpcServer, clientApi)

	metrics.EnableMetrics(clientApi.server)

	log.Printf("client API server starting on port %d\n", port)

	err = clientApi.server.Serve(listener)

	if err != nil {
		log.Fatalf("failed to start clientApi %s", err)
	}
}

func (clientApi *RaftServer) Stop() {

	clientApi.server.Stop()
	log.Println("client API `server stopped")
}

func (clientApi *RaftServer) Health(context.Context, *api.Empty) (*api.HealthResponse, error) {

	response := &api.HealthResponse{
		Status: api.HealthStatusCodes_OK,
	}

	return response, nil
}

// PutItem is a gRPC controller function used to store, "Put", an Item to the Raft cluster.
func (clientApi *RaftServer) PutItem(ctx context.Context, item *api.Item) (*api.PutItemResponse, error) {

	replicationCh, preReplicationErr := clientApi.policyClient.Put(item.Data)

	var putErr error

	var response *api.PutItemResponse

	if preReplicationErr == nil {

		wasReplicationSuccessful := <-replicationCh

		if wasReplicationSuccessful {
			response = createReplSuccessResponse()
		} else {

			response = createReplFailResponse()
			putErr = ErrPutRetryable
		}
	} else {

		response = createPreReplFailureResponse()
		putErr = ErrPutRetryable
	}

	return response, putErr
}

// GetItem is a gRPC controller function used to retrieve, "Get", an Item from the Raft cluster.
func (clientApi *RaftServer) GetItem(ctx context.Context, item *api.Item) (*api.GetItemResponse, error) {

	var responseErr error
	var response *api.GetItemResponse

	data, err := clientApi.policyClient.Get(item.Data)

	if err != nil {
		response, responseErr = createGeneralGetFailResponse()
	} else {
		response = createSuccessResponse(data)
	}

	return response, responseErr
}

func (clientApi *RaftServer) PolicyClient() client.Persister {

	return clientApi.policyClient
}

func createPreReplFailureResponse() *api.PutItemResponse {

	response := &api.PutItemResponse{}

	response.Status = api.ClientStatusCodes_PUT_ERROR
	response.IsRetryable = api.RetryCodes_YES

	return response
}

func createReplFailResponse() *api.PutItemResponse {

	response := &api.PutItemResponse{}

	response.ReplicationStatus = api.ReplicationCodes_FAILURE_TO_REACH_QUORUM
	response.IsRetryable = api.RetryCodes_YES
	response.Status = api.ClientStatusCodes_PUT_ERROR

	return response
}

func createReplSuccessResponse() *api.PutItemResponse {

	response := &api.PutItemResponse{}

	response.ReplicationStatus = api.ReplicationCodes_QUORUM_REACHED
	response.Status = api.ClientStatusCodes_PUT_OK
	response.IsRetryable = api.RetryCodes_NO

	return response
}

func createSuccessResponse(data []byte) *api.GetItemResponse {

	response := &api.GetItemResponse{
		Item: &api.Item{
			Data: data,
		},
		IsRetryable: api.RetryCodes_YES,
		Status:      api.ClientStatusCodes_GET_OK,
	}

	return response
}

func createGeneralGetFailResponse() (*api.GetItemResponse, error) {

	response := &api.GetItemResponse{

		Status:       api.ClientStatusCodes_GET_ERROR,
		IsRetryable:  api.RetryCodes_YES,
		ErrorMessage: ErrGetGeneralFailure.Error(),
	}

	return response, ErrGetGeneralFailure
}
