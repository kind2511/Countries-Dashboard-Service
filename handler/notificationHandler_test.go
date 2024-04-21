package handler

import (
	"assignment2/utils"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// Test function for ValidateEvent function
func TestValidEvent(t *testing.T) {
	// Test valid event
	if !utils.ValidateEvent("REGISTER") || !utils.ValidateEvent("INVOKE") || !utils.ValidateEvent("CHANGE") || !utils.ValidateEvent("DELETE") {
		t.Errorf("Expected true, got false")
	}
	// Test invalid event
	if utils.ValidateEvent("INVALID") {
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

// Test function for NotificationHandler that is not supported (HEAD)
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

// Test function for GetWebHooks function
func TestGetWebHooks(t *testing.T) {
	// Create a mock http.ResponseWriter
	w := httptest.NewRecorder()

	// Call getWebHooks with a nil Firestore client
	client = nil
	getWebHooks(w, nil)

	// Check the response status code
	if status := w.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	// Check the response body
	expected := "Error retrieving document"
	if strings.TrimSpace(w.Body.String()) != strings.TrimSpace(expected) {
		t.Errorf("handler returned wrong body: got %v want %v", strings.TrimSpace(w.Body.String()), strings.TrimSpace(expected))
	}

	req, err := http.NewRequest("GET", "/webhook/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() returned error: %v", err)
	}

	rr := httptest.NewRecorder()
	getWebHooks(rr, req)

	// The expected status code is 400 Bad Request
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

// Test function for IsDigit function
func TestIsDigit(t *testing.T) {
	// Create a test struct with input and expected output
	tests := []struct {
		input byte
		want  bool
	}{
		// Test with numbers (valid) and letters (invalid)
		{'0', true},
		{'1', true},
		{'a', false},
		{'A', false},
	}
	// Loop through the test struct and check validity
	for _, test := range tests {
		got := utils.IsDigit(test.input)
		if got != test.want {
			t.Errorf("isDigit(%v) = %v; want %v", test.input, got, test.want)
		}
	}
}

// Test function for PostWebhook function nr 2
func TestPostWebhook(t *testing.T) {
	// Create a test struct with input and expected output
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		// Fill it with test cases both valid and invalid
		{
			name:       "valid localhost url",
			body:       `{"Url": "http://localhost:8000/path", "Country": "NO", "Event": "Temperature"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "localhost url too short",
			body:       `{"Url": "http://localhost:80", "Country": "NO", "Event": "Temperature"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "localhost url missing slash",
			body:       `{"Url": "http://localhost:8000path", "Country": "NO", "Event": "Temperature"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "localhost url invalid port",
			body:       `{"Url": "http://localhost:80a0/path", "Country": "NO", "Event": "Temperature"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid country code",
			body:       `{"Url": "http://example.com", "Country": "INVALID", "Event": "Temperature"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "localhost url invalid port",
			body:       `{"Url": "http://localhost:80a0/path", "Country": "NO", "Event": "Temperature"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing url",
			body:       `{"Country": "Norway", "Event": "Temperature"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing country",
			body:       `{"Url": "http://example.com", "Event": "Temperature"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing event",
			body:       `{"Url": "http://example.com", "Country": "Norway"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid event",
			body:       `{"Url": "http://example.com", "Country": "Norway", "Event": "Invalid"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "localhost url too short",
			body:       `{"Url": "http://localhost:80", "Country": "Norway", "Event": "INVOKE"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "localhost url missing slash",
			body:       `{"Url": "http://localhost:8000path", "Country": "Norway", "Event": "INVOKE"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "country with space",
			body:       `{"Url": "http://localhost:1234/path", "Country": "No rway ", "Event": "Temperature"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	// Loop through the test struct and check validity
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/webhook", strings.NewReader(test.body))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(postWebhook)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != test.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, test.wantStatus)
			}
		})
	}

	// Check that the generated ID is of the correct length
	uniqueID := utils.GenerateUID(5)

	if len(uniqueID) != 5 {
		t.Errorf("Expected length to be 5, got %v", len(uniqueID))
	}
}

// Test function for DeleteWebhook function
func TestDeleteWebhook(t *testing.T) {
	// Test without webhook ID in the URL
	req, err := http.NewRequest("DELETE", "/dashboard/v1/webhook/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() returned error: %v", err)
	}
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	deleteWebhook(rr, req)

	// Test the expected error output
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

// Test function for InvocationHandler function
func TestInvocationHandler(t *testing.T) {
	// Create a resopnse recorder
	rr := httptest.NewRecorder()

	// Check that the handler returns the correct status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test with invalid event type
	_, err := http.NewRequest("POST", "/invoke", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() returned error: %v", err)
	}
	// Update the recorder to test the invalid event type
	rr = httptest.NewRecorder()
	invocationHandler(rr, "INVALID", "NO")

	// Check that the handler returns the correct status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

// Test function for callURL function
func TestCallURL(t *testing.T) {
	// Start an HTTP test server and close it when the test is done
	test := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World")
	}))
	defer test.Close()

	// Test callURL
	hook := utils.WebhookInvokeMessage{
		Url:     test.URL,
		Event:   "TestEvent",
		Country: "TestCountry",
	}
	// Create a ResponseRecorder to record the response and call the function
	rr := httptest.NewRecorder()
	callUrl(rr, hook)

	// Check that it returns the correct status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	// set up buffer to capture log output, and reset when done
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	// Test with invalid URL
	hook = utils.WebhookInvokeMessage{
		Url:     "http://invalid-url",
		Event:   "TestEvent",
		Country: "TestCountry",
	}
	// Update response recorder and call the function
	rr = httptest.NewRecorder()
	callUrl(rr, hook)

	// Check the log output
	if !strings.Contains(buf.String(), "Error in HTTP request") {
		t.Errorf("handler did not log correct message: got %v want %v", buf.String(), "Error in HTTP request")
	}

	// Clear the log output
	buf.Reset()

	// Test with invalid JSON payload
	hook = utils.WebhookInvokeMessage{
		Url:     "http://valid-url",
		Event:   string([]byte{0x80, 0x81, 0x82, 0x83}), // Invalid 0x80-0x83
		Country: "TestCountry",
	}
	// Update response recorder and call the function
	rr = httptest.NewRecorder()
	callUrl(rr, hook)

	// Check the log output is as expected
	if !strings.Contains(buf.String(), "Error in HTTP request") {
		t.Errorf("handler did not log correct message: got %v want %v", buf.String(), "Error in HTTP request")
	}
}