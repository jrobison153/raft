package acceptance_tests

import (
	"context"
	"fmt"
	"github.com/jrobison153/raft/api"
	"github.com/jrobison153/raft/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

var serverPort uint32 = 8088

var raftServer *server.RaftClientApi

func Setup() {

	raftServer = new(server.RaftClientApi)

	go raftServer.Start(serverPort)
}

func Teardown() {

	raftServer.Stop()
}

func TestWhenAnItemIsPutItCanBeRetrieved(t *testing.T) {

	Setup()
	defer Teardown()

	itemToPut := api.Item{
		Key:  "af159-ef7ff",
		Data: []byte("42"),
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	serverAddress := fmt.Sprintf("localhost:%d", serverPort)

	conn, err := grpc.Dial(serverAddress, opts...)

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	client := api.NewPersisterClient(conn)

	_, err = client.PutItem(context.Background(), &itemToPut)

	if err != nil {
		t.Errorf("PutItem request failed with error %s", err.Error())
	}
}
