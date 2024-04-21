package handler

import (
	"assignment2/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

// Test for postRegistration function
func TestPostRegistration(t *testing.T) {
	// Create a Firestore object with all fields filled
	dashboard := utils.Firestore{
		ID:      "testID",
		Country: "testCountry",
		Features: utils.Features{
			Temperature:      new(bool),
			Precipitation:    new(bool),
			Capital:          new(bool),
			Coordinates:      new(bool),
			Population:       new(bool),
			Area:             new(bool),
			TargetCurrencies: []string{"USD", "EUR"},
		},
		IsoCode:    "testIsoCode",
		LastChange: time.Now(),
	}

	// Convert the Firestore object to JSON
	dashboardJSON, err := json.Marshal(dashboard)
	if err != nil {
		t.Fatalf("json.Marshal() returned error: %v", err)
	}

	// Create a new request with the JSON body
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(dashboardJSON))
	if err != nil {
		t.Fatalf("http.NewRequest() returned error: %v", err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call postRegistration with the request and the ResponseRecorder
	postRegistration(rr, req)

	// Check the result
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("postRegistration() returned status code %v; want %v", rr.Code, http.StatusInternalServerError)
	}

	// Test with empty Country and IsoCode fields
	dashboard2 := utils.Firestore{
		ID:         "testID",
		Features:   utils.Features{},
		LastChange: time.Now(),
	}

	// Convert the Firestore object to JSON
	dashboardJSON, err2 := json.Marshal(dashboard2)
	if err2 != nil {
		t.Fatalf("json.Marshal() returned error: %v", err)
	}
	// Create a new request with the JSON body
	req, err = http.NewRequest("POST", "/register", bytes.NewBuffer(dashboardJSON))
	if err != nil {
		t.Fatalf("http.NewRequest() returned error: %v", err)
	}
	// Create a ResponseRecorder to record the response
	rr = httptest.NewRecorder()
	postRegistration(rr, req)
	// Check the result
	if rr.Code != http.StatusBadRequest {
		t.Errorf("postRegistration() returned status code %v; want %v", rr.Code, http.StatusBadRequest)
	}

	// Create a new body with missing variables
	body := `{
			"country": "Norway",
			"isoCode": "NO"
		}`
	// Create a new request with the body that is missing variables
	req2, err2 := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	if err2 != nil {
		t.Fatalf("http.NewRequest() returned error: %v", err)
	}
	// call the function with the new request
	postRegistration(rr, req2)

	// Check the response
	response := rr.Result()
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("postRegistration() returned wrong status code: got %v want %v", response.StatusCode, http.StatusBadRequest)
	}

	// Check the response body
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Returned error: %v", err)
	}
	// Check if the body contains the correct message
	if !strings.Contains(string(bodyBytes), "Missing variables:") {
		t.Errorf("postRegistration returned wrong body: got %v want %v", string(bodyBytes), "Missing :")
	}
}

// Test for DeleteDashboard function
func TestDeleteDashboard(t *testing.T) {
	// Test without dashboard ID
	req, err := http.NewRequest("DELETE", "/dashboard/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() returned error: %v", err)
	}
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	deleteDashboard(rr, req)
	// Check the result is as expected
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
