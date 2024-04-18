package handler

import (
	"assignment2/utils"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test function for ValidateEvent function
func TestValidEvent(t *testing.T) {
	// Test valid event
	if !ValidateEvent("REGISTER") || !ValidateEvent("INVOKE") || !ValidateEvent("CHANGE") || !ValidateEvent("DELETE") {
		t.Errorf("Expected true, got false")
	}
	// Test invalid event
	if ValidateEvent("INVALID") {
		t.Errorf("Expected false, got true")

	}
}

// Test function for NotificationHandler function
func TestNotificationHandler(t *testing.T) {

	// Initialize handler instance
	handler := NotificationHandler()

	// set up structure to be used for testing and close when finished testing
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// local server URL
	fmt.Println("URL: ", server.URL)

	// HTTP method checking for POST
	req, err := http.NewRequest("POST", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response.
	rr := httptest.NewRecorder()
	handler2 := http.HandlerFunc(StatusHandler())

	handler2.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotImplemented {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotImplemented)
	}

	// Check the response body is what we expect for unsupported method (trimming whitespace to get equal strings)
	expected := "Method not supported. Currently only GET is supported."
	if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			strings.TrimSpace(rr.Body.String()), strings.TrimSpace(expected))
	}

}

// Test function for NotificationHandler_POST function
func TestNotificationHandler_POST(t *testing.T) {
	// Create a mock HTTP request with data in the body
	body := bytes.NewBuffer([]byte(`{"id": "12345", "url":"http://test.com", "country":"NO", "event": "INVOKE"}`)) // Replace with your actual data
	req, err := http.NewRequest("POST", "/notification", body)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NotificationHandler())

	// Call ServeHTTP directly and pass in our Request and ResponseRecorder
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotImplemented)
	}

	// Check the response body is what we expect for unsupported method (trimming whitespace to get equal strings)
	expected := "failed decoding JSON"
	if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			strings.TrimSpace(rr.Body.String()), strings.TrimSpace(expected))
	}

}

func TestNotificationHandler_UnsupportedMethod(t *testing.T) {
	// Create a mock HTTP request with an unsupported method
	req, err := http.NewRequest("HEAD", "/notification", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NotificationHandler())

	// Call ServeHTTP directly and pass in our Request and ResponseRecorder
	handler.ServeHTTP(rr, req)

	// Check the status code is StatusMethodNotAllowed
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}

	// Check the response body contains the expected error message
	expected := "Method HEAD not supported for " + utils.NOTIFICATION_PATH
	if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expected) {
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	}
}
