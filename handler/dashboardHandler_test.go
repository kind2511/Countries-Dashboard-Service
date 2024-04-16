package handler

import (
	"assignment2/utils"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDashboardsHandler(t *testing.T) {

	// Initialize handler instance
	handler := DashboardHandler()

	// set up structure to be used for testing and close when finished testing
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// local server URL
	fmt.Println("URL: ", server.URL)

	// Create client instance
	client := http.Client{}

	// test Get request
	res, err := client.Get(server.URL + utils.DASHBOARD_PATH)
	if err != nil {
		t.Fatal("Get request to URL failed:", err.Error())
	}

	// test http status
	if res.Status != "200 OK" {
		t.Fatal("Get request has wrong status code", err.Error())
	}

	// test of unsupported method
	res2, err2 := client.Head(server.URL + utils.DASHBOARD_PATH)
	if err2 != nil {
		t.Fatal("Head request to URL failed:", err.Error())
	}

	// test http status for unsupported method
	if res2.Status != "405 Method Not Allowed" {
		t.Fatal("un supported method has wrong status code", err.Error())
	}

	// HTTP method checking for POST
	req, err := http.NewRequest("POST", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response.
	rr := httptest.NewRecorder()
	handler2 := http.HandlerFunc(DashboardHandler())

	handler2.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}

	// Check the response body is what we expect for unsupported method (trimming whitespace to get equal strings)
	expected := "Method POST not supported."
	if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			strings.TrimSpace(rr.Body.String()), strings.TrimSpace(expected))
	}

}

// Test function for fetchURLdata
func TestFetchURLData(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		rw.Write([]byte(`{"key":"value"}`))
	}))
	// Close the server when test finishes
	defer server.Close()

	// Use server.URL as the argument to fetchURLdata
	var data map[string]string
	err := fetchURLdata(server.URL, nil, &data)
	if err != nil {
		t.Fatal("Expected error, got nil")
	}

	// Check the data
	expected := "Failed to fetch url: " + server.URL
	if data["key"] != "value" {
		t.Errorf("Expected value for key to be '%s', got %s", expected, err.Error())
	}

}

// Test function whatTimeNow
func TestWhatTimeNow(t *testing.T) {
	// Call the function
	timeFunc := whatTimeNow()

	// Check that the string is in the correct format
	format := "20060102 15:04"
	// Parse the time string
	_, err := time.Parse(format, timeFunc)
	// Check if there is an error
	if err != nil {
		t.Errorf("Returned time is not in the correct format: %v", err)
	}
}

// Test function for floatFormat
func TestFloatFormat(t *testing.T) {
	// Call the function
	floatFunc, err := floatFormat(1.23456789)

	// Check if there is an error
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// Check the value of the float is as expected two decimals
	expected := myFloat(1.23)
	if floatFunc != expected {
		t.Errorf("Expected %v, got %v", expected, floatFunc)
	}
}
func TestRetrieveCountryData(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		rw.Write([]byte(`[{"population": 123456, "capital": ["Capital"], "currencies": {"Currency": {}}, "area": 123.45}]`))
	}))
	// Close the server when test finishes
	defer server.Close()

	// Call the function with the mock server URL
	population, capital, currency, area, err := retrieveCountryData(server.URL+"/name/", "TestCountry", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check the data
	if population != 123456 || capital != "Capital" || currency != "Currency" || area != 123.45 {
		t.Errorf("Expected population to be 123456, capital to be Capital, currency to be Currency, and area to be 123.45, got %v, %v, %v and %v", population, capital, currency, area)
	}
}

func TestRetrieveCoordinates(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		rw.Write([]byte(`{"results": [{"geometry": {"location": {"lng": 0, "lat": 0}}}]}`))
	}))
	// Close the server when test finishes
	defer server.Close()

	// Call the function with the mock server URL
	longitude, latitude, err := retrieveCoordinates(server.URL+"/json?address=", "TestLocation", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check the data
	if longitude != 0 || latitude != 0 {
		t.Errorf("Expected longitude to be 0 and latitude to be 0, got %v and %v", longitude, latitude)
	}
}

func TestRetrieveCurrencyExchangeRates(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		rw.Write([]byte(`{"rates": {"USD": 1.23, "EUR": 0.89}}`))
	}))
	// Close the server when test finishes
	defer server.Close()

	// Call the function with the mock server URL
	currencyData, err := retrieveCurrencyExchangeRates(server.URL+"/latest?base=", "USD", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check the data
	if currencyData["USD"] != 1.23 || currencyData["EUR"] != 0.89 {
		t.Errorf("Expected currencyData to be {\"USD\": 1.23, \"EUR\": 0.89}, got %v", currencyData)
	}
}
