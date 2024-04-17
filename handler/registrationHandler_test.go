package handler

import (
	"assignment2/utils"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestIsEmptyField(t *testing.T) {
	// Test with empty string
	if !isEmptyField("") {
		t.Error("expected isEMptyField to return true, got false")
	}

	// Test with a filled string
	if isEmptyField("test not empty") {
		t.Error("expected isEmptyField to return false, got true")
	}

	// Test with a pointer to a bool
	var testBool *bool
	if !isEmptyField(testBool) {
		t.Error("expected isEmptyField to return true, got false")
	}
	// Test with a pointer to a bool that should fail
	testBool2 := new(bool)
	if isEmptyField(testBool2) {
		t.Error("expected isEmptyField to return false, got true")
	}

	// Test with strings
	if !isEmptyField([]string{}) {
		t.Error("expected isEmptyField to return true, got false")
	}
	// Test with strings that is filled
	if isEmptyField([]string{"test", "with", "several", "strings"}) {
		t.Error("expected isEmptyField to return false, got true")
	}

	// Test default option
	if isEmptyField(12345) {
		t.Error("expected isEmptyField to return false, got true")
	}

}

// Test function whatTimeNow
func TestWhatTimeNow2(t *testing.T) {
	// Call the function
	timeFunc := whatTimeNow2()

	// Check that the string is in the correct format
	format := "2006-01-02 15:04"
	// Parse the time string
	_, err := time.Parse(format, timeFunc)
	// Check if there is an error
	if err != nil {
		t.Errorf("Returned time is not in the correct format: %v", err)
	}
}

// Test function checkValidCurrencies
func TestCheckValidCurrencies(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a mock http.ResponseWriter
	w := httptest.NewRecorder()

	// Create a mock Dashboard
	dashboard := utils.Dashboard{
		RegFeatures: utils.RegFeatures{
			TargetCurrencies: []string{"NOK", "nok", "EUR", "eur", ""},
		},
	}

	// Call checkValidCurrencies with the mock http.ResponseWriter and Dashboard
	currencies, err := checkValidCurrencies(server.URL+"/", w, dashboard)
	if err != nil {
		t.Fatal(err)
	}

	// Check that checkValidCurrencies returns the correct currencies
	expectedCurrencies := []string{"NOK", "EUR"}
	if !reflect.DeepEqual(currencies, expectedCurrencies) {
		t.Errorf("expected currencies to be %v, got %v", expectedCurrencies, currencies)
	}

	// Call checkValidCurrencies with an invalid URL
	_, err = checkValidCurrencies("http://invalid-url", w, dashboard)
	if err == nil {
		t.Error("expected an error, got nil")
	}
}
