package handler

import (
	"assignment2/utils"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Test function for ISEmptyField function
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
	currencies, err := checkValidCurrencies(w, dashboard)
	if err != nil {
		t.Fatal(err)
	}

	// Check that checkValidCurrencies returns the correct currencies
	expectedCurrencies := []string{"NOK", "EUR"}
	if !reflect.DeepEqual(currencies, expectedCurrencies) {
		t.Errorf("expected currencies to be %v, got %v", expectedCurrencies, currencies)
	}
}

// Helper function to create a bool pointer
func BoolPtr(b bool) *bool {
	return &b
}

// Test function UpdatedData
func TestUpdatedData(t *testing.T) {
	// Make an empty object
	emptyObject := &utils.Firestore{}
	// FIll the object with data
	filledObject := &utils.Firestore{
		Country: "Norway",
		IsoCode: "NO",
		Features: utils.Features{
			Temperature:      BoolPtr(true),
			Precipitation:    BoolPtr(true),
			Capital:          BoolPtr(true),
			Coordinates:      BoolPtr(true),
			Population:       BoolPtr(true),
			Area:             BoolPtr(true),
			TargetCurrencies: []string{"NOK", "EUR"},
		},
	}

	// Call the function
	w := httptest.NewRecorder()
	updatedObject, missing, missingElements := updatedData(emptyObject, filledObject, w)

	// Check that the fields are updated correctly
	if updatedObject.Country != "Norway" || updatedObject.IsoCode != "NO" || *updatedObject.Features.Area != true {
		t.Errorf("The fields were not successfully updated")
	}
	// Check that the missing elements are correct
	if missing || len(missingElements) != 0 {
		t.Errorf("Did not identify missing objects correctly")
	}

	// Make the Country field empty
	filledObject.Country = ""
	filledObject.IsoCode = ""
	// Call the function
	updatedObject, missing, missingElements = updatedData(emptyObject, filledObject, w)

	// Check that the fields are updated correctly
	if !missing || len(missingElements) != 2 || missingElements[0] != "Country" || missingElements[1] != "IsoCode" {
		t.Errorf("Did not identify missing objects correctly")
	}

	// Make the Features field empty
	emptyFilledObject := &utils.Firestore{}

	// Call the function
	updatedObject, missing, missingElements = updatedData(emptyObject, emptyFilledObject, w)
	// Check that the fields are updated correctly
	if !missing || len(missingElements) != 9 {
		t.Errorf("Did not identify missing objects correctly %v", missingElements)
	}
}

// Test function handleValidCountryAndCode
func TestHandleValidCountryAndCode(t *testing.T) {
	// Create a mock HTTP server and close when finished testing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write a mock response
		if strings.Contains(r.URL.Path, "notFound") {
			http.Error(w, "Not Found", http.StatusNotFound)
		} else {
			w.Write([]byte(`[{"name": {"common": "Norway"}, "cca2": "NO"}]`))
		}
	}))
	defer server.Close()

	// Create a mock http.ResponseWriter
	w := httptest.NewRecorder()
	// fill the dashboard with data
	dashboard := utils.Dashboard{
		Country: "Norway",
		Isocode: "NO",
	}
	// Call the function
	isocode, country, err := handleValidCountryAndCode(w, dashboard)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// check that the isocode and country are correct
	// Call the function again with the updated dashboard
	isocode2, country2, err := handleValidCountryAndCode(w, dashboard)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// check that the isocode and country are correct
	if isocode2 != "NO" || country2 != "Norway" {
		t.Errorf("expected isocode to be NO and country to be Norway, got %v and %v", isocode, country)
	}
}
