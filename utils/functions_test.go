package utils

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckCurrencies(t *testing.T) {
	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	// Test data
	currencies := []string{"NOK", "EUR", "INVALID"}

	// Call the function with the test data
	result := CheckCurrencies(currencies, rr)

	// Check the result
	expected := []string{"NOK", "EUR"}
	if len(result) != len(expected) {
		t.Errorf("checkCurrencies() returned %v, want %v", result, expected)
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("checkCurrencies() returned %v, want %v", result, expected)
			break
		}
	}
}

// Test for IsEmptyField function
func TestIsEmptyField(t *testing.T) {
	// Create test cases
	tests := []struct {
		name string
		arg  interface{}
		want bool
	}{
		{"Empty string", "", true},
		{"Non-empty string", "test", false},
		{"Nil boolean pointer", (*bool)(nil), true},
		{"Non-nil boolean pointer", new(bool), false},
		{"Empty string slice", []string{}, true},
		{"Non-empty string slice", []string{"test"}, false},
		{"Other type (int)", 0, false},
	}
	// Loop through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmptyField(tt.arg); got != tt.want {
				t.Errorf("IsEmptyField() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test for WhatTimeNow function
func TestWhatTimeNow(t *testing.T) {
	// Call the function
	result := WhatTimeNow()

	// Parse the result
	parsedTime, err := time.Parse("20060102 15:04", result)
	if err != nil {
		t.Errorf("WhatTimeNow() returned an invalid time: %v", err)
	}

	// Check that time is not more than a minute in the past (account for time execution)
	if time.Since(parsedTime) > time.Minute {
		t.Errorf("WhatTimeNow() returned a time more than a minute in the past")
	}
}

// BoolPtr is a helper function for creating a pointer to a bool.
func BoolPtr(b bool) *bool {
	return &b
}

// Test function UpdatedData
func TestUpdatedData(t *testing.T) {
	// Make an empty object
	emptyObject := &Firestore{}
	// FIll the object with data
	filledObject := &Firestore{
		Country: "Norway",
		IsoCode: "NO",
		Features: Features{
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
	updatedObject, missing, missingElements := UpdatedData(emptyObject, filledObject, w)

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
	updatedObject, missing, missingElements = UpdatedData(emptyObject, filledObject, w)

	// Check that the fields are updated correctly
	if !missing || len(missingElements) != 2 || missingElements[0] != "Country" || missingElements[1] != "IsoCode" {
		t.Errorf("Did not identify missing objects correctly")
	}

	// Make the Features field empty
	emptyFilledObject := &Firestore{}

	// Call the function
	updatedObject, missing, missingElements = UpdatedData(emptyObject, emptyFilledObject, w)
	// Check that the fields are updated correctly (bools will be false, but not empty)
	if !missing || len(missingElements) != 9 {
		t.Errorf("Did not identify missing objects correctly %v", missingElements)
	}
}

/*
// Test function CheckCountry
func TestCheckCountry(t *testing.T) {
	// Create a mock HTTP server and close when finished testing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write a mock response
		if r.URL.Path == "/name/Norway" || r.URL.Path == "/iso/NO" {
			w.Write([]byte(`[{"name": {"common": "Norway"}, "cca2": "NO"}]`))
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Set the COUNTRIES_API_NAME and COUNTRIES_API_ISOCODE to the mock server's URL
	COUNTRIES_API_NAME = server.URL + "/name/"
	COUNTRIES_API_ISOCODE = server.URL + "/iso/"

	// Create a mock http.ResponseWriter
	w := httptest.NewRecorder()

	// Test with valid country name and ISO code
	name, iso, err := CheckCountry("Norway", "NO", w)
	if name != "Norway" || iso != "NO" || err != nil {
		t.Errorf("CheckCountry() returned %v, %v, %v; want Norway, NO, nil", name, iso, err)
	}

	// Test with invalid country name and ISO code
	name, iso, err = CheckCountry("InvalidCountry", "InvalidISO", w)
	if name != "" || iso != "" || !errors.Is(err, errors.New("no valid countries")) {
		t.Errorf("CheckCountry() returned %v, %v, %v; want '', '', 'no valid countries'", name, iso, err)
	}
}
*/
