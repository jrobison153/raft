package acceptance_tests

import (
	"context"
	"fmt"
	"github.com/jrobison153/raft/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"os"
	"os/exec"
	"testing"
)

const serverPort = "8088"

type ProcessContext struct {
	RaftCmd            *exec.Cmd
	StdErr             io.ReadCloser
	CommandErrMessages []byte
	StdOut             io.ReadCloser
	CommandOutMessages []byte
}

func init() {

	buildCommand := exec.Command("go", "build")

	buildStdErr, _ := buildCommand.StderrPipe()

	buildStdOut, _ := buildCommand.StdoutPipe()

	buildCommand.Dir = ".."

	err := buildCommand.Start()

	if err != nil {
		log.Fatalf("Failed to build the raft executable, required for these test: %v", err)
	}

	stdErrAsStr, _ := io.ReadAll(buildStdErr)
	stdOutAsStr, _ := io.ReadAll(buildStdOut)

	log.Printf("go build stdErr: %s", stdErrAsStr)
	log.Printf("go build stdOut: %s", stdOutAsStr)

	buildCommand.Wait()
}

func Setup() *ProcessContext {

	portEnvVar := fmt.Sprintf("PORT=%s", serverPort)

	raftProcessContext := &ProcessContext{}

	raftProcessContext.RaftCmd = exec.Command("../raft")

	raftProcessContext.RaftCmd.Env = append(
		os.Environ(),
		portEnvVar,
	)

	var pipeErr error

	raftProcessContext.StdErr, pipeErr = raftProcessContext.RaftCmd.StderrPipe()

	if pipeErr != nil {
		log.Fatalf("Failed to create stdErr pipe: %v", pipeErr)
	}

	raftProcessContext.StdOut, pipeErr = raftProcessContext.RaftCmd.StdoutPipe()

	if pipeErr != nil {
		log.Fatalf("Failed to create stdOut pipe: %v", pipeErr)
	}

	raftProcessContext.runRaftAndListenForError()

	return raftProcessContext
}

func (c *ProcessContext) runRaftAndListenForError() {

	errCh := make(chan error)
	go func(ch chan error) {

		startErr := c.RaftCmd.Run()
		errCh <- startErr
	}(errCh)

	go func(ch chan error) {
		startErr := <-ch
		log.Printf("Raft server stopping, reason: %v\n", startErr)
	}(errCh)
}

func (c *ProcessContext) Teardown() {

	killErr := c.RaftCmd.Process.Signal(os.Interrupt)

	c.CommandErrMessages, _ = io.ReadAll(c.StdErr)
	c.CommandOutMessages, _ = io.ReadAll(c.StdOut)

	log.Printf("Stderr from raft executable: %s", c.CommandErrMessages)

	log.Printf("Stdout from raft executable: %s", c.CommandOutMessages)

	if killErr != nil {
		log.Printf("Process kill failed with error %v", killErr)
	}

	os.Unsetenv("PORT")
}

func TestWhenServerRunningThenHealthStatusIsOk(t *testing.T) {

	raftProcessContext := Setup()
	defer raftProcessContext.Teardown()

	conn, client := createClient()

	defer conn.Close()

	healthResponse, err := client.Health(context.Background(), &api.Empty{})

	if err != nil {
		t.Errorf("Error retrieving health status %v", err)
	}

	if healthResponse.Status != api.HealthStatusCodes_OK {
		t.Errorf("Expected health status to be ok but got %v", healthResponse.Status)
	}
}

func TestWhenAnItemIsPutItCanBeRetrieved(t *testing.T) {

	raftProcessContext := Setup()
	defer raftProcessContext.Teardown()

	itemToPut := api.Item{
		Key:  "af159-ef7ff",
		Data: []byte("42"),
	}

	conn, client := createClient()

	defer conn.Close()

	_, err := client.PutItem(context.Background(), &itemToPut)

	if err != nil {
		t.Errorf("PutItem request failed with error %s", err.Error())
	} else {

		key := &api.Key{
			Key: itemToPut.Key,
		}

		getResponse, err := client.GetItem(context.Background(), key)

		if err != nil {
			t.Errorf("GetItem request failed with error %s", err.Error())
		}

		retrievedItem := getResponse.Item

		if retrievedItem.Key != itemToPut.Key {
			t.Errorf(
				"Retrieved item keys do not match. Got '%s' expected '%s'",
				retrievedItem.Key,
				itemToPut.Key)
		}

		if string(retrievedItem.Data) != string(itemToPut.Data) {
			t.Errorf(
				"Retrieved item data values do not match. Got '%s' expected '%s'",
				retrievedItem.Data,
				itemToPut.Data)
		}
	}
}

func createClient() (*grpc.ClientConn, api.PersisterClient) {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	retryPolicy := `{
            "methodConfig": [{
                "name": [
					{
						"service": "",
						"method": ""
					}
				],
                "waitForReady": true,

                "retryPolicy": {
                    "MaxAttempts": 4,
                    "InitialBackoff": ".01s",
                    "MaxBackoff": ".01s",
                    "BackoffMultiplier": 1.0,
                    "RetryableStatusCodes": [ "UNAVAILABLE" ]
                }
            }]
        }`

	opts = append(opts, grpc.WithDefaultServiceConfig(retryPolicy))

	serverAddress := fmt.Sprintf("localhost:%s", serverPort)
	conn, err := grpc.Dial(serverAddress, opts...)

	if err != nil {
		panic(err)
	}

	client := api.NewPersisterClient(conn)

	return conn, client
}
