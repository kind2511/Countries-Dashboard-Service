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
	"time"
)

func TestStatusHandler(t *testing.T) {

	// Initialize handler instance
	handlerStatus := StatusHandler

	// Set up structure to be used for testing
	server := httptest.NewServer(http.HandlerFunc(handlerStatus))
	defer server.Close()

	// Create client instance
	client := http.Client{}

	// URL where instance is running
	fmt.Println("URL: ", server.URL)

	// Retrieve content from server
	res, err := client.Get(server.URL + utils.STATUS_PATH)
	if err != nil {
		t.Fatal("Get request to URL failed:", err.Error())
	}

	// Decode JSON response body into a status struct
	var s utils.Status
	err2 := json.NewDecoder(res.Body).Decode(&s)
	if err2 != nil {
		t.Fatal("Error during decoding:", err2.Error())
	}

	// Mocking the start time for testing purposes
	startTime = time.Now()

	// Mock the external API calls using httptest.NewServer
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == utils.COUNTRIES_API+"/name/norway" {
			w.WriteHeader(http.StatusOK)
		} else if r.URL.String() == utils.GEOCODING_API+"Norway&count=1&language=en&format=json" {
			w.WriteHeader(http.StatusOK)
		} else if r.URL.String() == utils.CURRENCY_API+"/nok" {
			w.WriteHeader(http.StatusOK)
		} else if r.URL.String() == "https://console.firebase.google.com/project/prog2005-assignment2-ee93a/firestore/databases/-default-/data/~2FDashboard~2FXhjRXlZcd7uLTMd8kdxb" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	// Close the mock server when the test ends
	defer mockServer.Close()

	// Create a request to the mock server
	req := httptest.NewRequest("GET", "/", nil)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the StatusHandler function with the ResponseRecorder and Request
	StatusHandler(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	var result utils.Status
	err3 := json.Unmarshal(rr.Body.Bytes(), &result)
	if err3 != nil {
		t.Errorf("error unmarshalling JSON response: %v", err3)
	}

	// Check if there is no error returned from StatusHandler
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("unexpected status code returned by StatusHandler: %v", status)
	}

}

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
