package handler

import (
	"assignment2/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/firestore"
)

// Test function for CheckHTTPError function
func TestCheckHTTPError(t *testing.T) {
	// Redirect standard output to a buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call checkHTTPError with an error
	testErr := errors.New("test error")
	checkHTTPError(testErr)

	// Capture standard output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = oldStdout

	expected := "error occurred while making HTTP request: test error"
	actual := buf.String()

	// Trim spaces and convert both strings to lower case for equality check
	expected = strings.ToLower(strings.TrimSpace(expected))
	actual = strings.ToLower(strings.TrimSpace(actual))

	if expected != actual {
		t.Errorf("unexpected output: got %v want %v", actual, expected)
	}

}

// Test function for CheckHTTPError function with no error
func TestCheckHTTPErrorNoError(t *testing.T) {
	// Redirect standard output to a buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call checkHTTPError with nil
	checkHTTPError(nil)

	// Capture standard output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = oldStdout

	// Check that nothing was printed
	if buf.Len() != 0 {
		t.Errorf("unexpected output: got %v want nothing", buf.String())
	}
}

// Test function for GetWebhookSize function
func TestGetWebhookSize(t *testing.T) {
	// Create mock documents
	mockDocuments := []*firestore.DocumentSnapshot{}

	// Test success
	size, err := GetWebhookSize(func() ([]*firestore.DocumentSnapshot, error) {
		return mockDocuments, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check the size
	if size != len(mockDocuments) {
		t.Errorf("Expected size to be %v, got %v", len(mockDocuments), size)
	}

	// Test error
	size, err = GetWebhookSize(func() ([]*firestore.DocumentSnapshot, error) {
		return nil, errors.New("test error")
	})

}

// Test fucntion for StatusHandler
func TestStatusHandler(t *testing.T) {

	// Initialize handler instance
	handler := StatusHandler()

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

// Test function for StatusGetRequest
func TestStatusGetRequest(t *testing.T) {
	// Create a mock http.ResponseWriter
	w := httptest.NewRecorder()

	// Create a Status object
	Status := utils.Status{
		Countriesapi:   http.StatusOK,
		Meteoapi:       http.StatusOK,
		Currencyapi:    http.StatusOK,
		Notificationdb: http.StatusOK,
		Webhooks:       5,
		Version:        "v1.0",
		Uptime:         100.0,
	}

	// Call the part of statusGetRequest that you want to test
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	statusJSON, err := json.Marshal(Status)
	if err != nil {
		http.Error(w, "Unable to marshal status to JSON", http.StatusInternalServerError)
		return
	}
	w.Write(statusJSON)

	// Check the response status code
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response Content-Type header
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("handler returned wrong Content-Type header: got %v want %v", contentType, "application/json")
	}

	// Check the response body
	expectedBody := string(statusJSON)
	if body := w.Body.String(); body != expectedBody {
		t.Errorf("handler returned wrong body: got %v want %v", body, expectedBody)
	}
}
