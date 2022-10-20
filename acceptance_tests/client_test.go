package acceptance_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jrobison153/raft/server"
	"net/http"
	"testing"
)

func Setup() {

	fmt.Println("First")
}

func Teardown() {

	fmt.Println("Last")
}

func TestWhenAnItemIsPutItCanBeRetrieved(t *testing.T) {

	Setup()
	defer Teardown()

	type AnItem struct {
		Key  string
		Data uint32
	}

	port := 8080
	url := fmt.Sprintf("http://localhost:%d/item", port)
	itemToPut := AnItem{"af159-ef7ff", 42}
	marshalledItemToPut, err := json.Marshal(itemToPut)

	if err != nil {
		panic(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(marshalledItemToPut))

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Errorf("Expected HTTP status code 201 but got %d", resp.StatusCode)
	}

	server.Start()
}
