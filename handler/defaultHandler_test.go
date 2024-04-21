package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test for default handler
func TestDefaultHandler(t *testing.T) {
	// Create a new request
	w := httptest.NewRecorder()

	// Create a new request
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Call the DefaultHandler function
	DefaultHandler(w, r)
	// Check the status code
	response := w.Result()
	if response.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", response.Status)
	}
	// Check the response body
	body, _ := io.ReadAll(response.Body)
	if !strings.Contains(string(body), "This service has the following endpoints with methods:") {
		t.Errorf("expected response body to contain 'This service has the following endpoints with methods:', got %v", string(body))
	}
}
