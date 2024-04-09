package handler

import (
	"assignment2/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusHandler(t *testing.T) {
	// Initialize handler instance
	handlerStatus := StatusHandler

	// Set up infrastructure to be used for invocation - important: wrap handler function in http.HandlerFunc()
	server := httptest.NewServer(http.HandlerFunc(handlerStatus))
	// Ensure it is torn down properly at the end
	defer server.Close()

	// Create client instance
	client := http.Client{}

	// URL under which server is instantiated
	fmt.Println("URL: ", server.URL)

	// Retrieve content from server
	res, err := client.Get(server.URL + utils.STATUS_PATH)
	if err != nil {
		t.Fatal("Get request to URL failed:", err.Error())
	}

	var s utils.Status
	err2 := json.NewDecoder(res.Body).Decode(&s)
	if err2 != nil {
		t.Fatal("Error during decoding:", err2.Error())
	}
}
