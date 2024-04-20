package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud.google.com/go/firestore"
)

// Test function for RegistrationHandler
func TestRegistrationHandler(t *testing.T) {

	// Initialize handler instance
	handler := RegistrationHandler()

	// set up structure to be used for testing and close when finished testing
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// local server URL
	fmt.Println("URL: ", server.URL)

	req, err := http.NewRequest("HEAD", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response.
	rr := httptest.NewRecorder()

	server.Config.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}

	// Check the response body is what we expect for unsupported method (trimming whitespace to get equal strings)
	expected := "Unsupported request methodHEAD"
	if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			strings.TrimSpace(rr.Body.String()), strings.TrimSpace(expected))
	}

}
func TestSetFirestoreClient(t *testing.T) {
	// Create a mock context
	ctx := context.Background()

	// Create a mock Firestore client
	client := &firestore.Client{} // Replace with your actual Firestore client

	// Call the function with the mock context and client
	SetFirestoreClient(ctx, client)

	// Check the result
	if client == nil {
		t.Errorf("SetFirestoreClient() did not correctly set the client")
	}
	if ctx == nil {
		t.Errorf("SetFirestoreClient() did not correctly set the context")
	}
}
