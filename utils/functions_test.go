package utils

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"
)

// Test for CheckCurrencies function
func TestCheckCurrencies(t *testing.T) {
	// Create a ResponseRecorder to record the response.
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

	// Check that the result is correct
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
	// FIll the object with data and bool pointer values to satisfy the struct
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
	// Check that the fields are updated correctly with more missing elements
	if !missing || len(missingElements) != 9 {
		t.Errorf("Did not identify missing objects correctly %v", missingElements)
	}
}

// Test for ValidateEvent function
func TestValidateEvent(t *testing.T) {
	// Create struc to contain name, argument and expected result
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		// Test cases with valid and invalid events
		{"Valid event REGISTER", "REGISTER", true},
		{"Valid event INVOKE", "INVOKE", true},
		{"Valid event CHANGE", "CHANGE", true},
		{"Valid event DELETE", "DELETE", true},
		{"Invalid event", "INVALID", false},
	}

	// Loop through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateEvent(tt.arg); got != tt.want {
				t.Errorf("ValidateEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test for IsDigit function
func TestIsDigit(t *testing.T) {
	// Create struct with name, argument and expected result
	tests := []struct {
		name string
		arg  byte
		want bool
	}{
		// Test cases with valid and invalid "digits"
		{"Digit 0", '0', true},
		{"Digit 1", '1', true},
		{"Letter a", 'a', false},
		{"Letter A", 'A', false},
	}

	// Loop through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDigit(tt.arg); got != tt.want {
				t.Errorf("IsDigit() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test for checkCountry function
func TestCheckCountry(t *testing.T) {
	// Create struct with name, country name, ISO code, expected name, expected ISO code and expected error
	tests := []struct {
		name        string
		countryName string
		isoCode     string
		wantName    string
		wantIso     string
		wantErr     error
	}{
		// Test cases with valid and invalid country names and ISO codes
		{
			"Valid country name",
			"USA",
			"",
			"United States",
			"US",
			nil,
		},
		{
			"Valid ISO code",
			"",
			"US",
			"United States",
			"US",
			nil,
		},
		{
			"No valid countries",
			"",
			"",
			"",
			"",
			errors.New("no valid countries"),
		},
	}

	// Loop through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()
			// Call the function
			gotName, gotIso, gotErr := CheckCountry(tt.countryName, tt.isoCode, rr)
			if gotName != tt.wantName || gotIso != tt.wantIso || (gotErr != nil && gotErr.Error() != tt.wantErr.Error()) {
				t.Errorf("CheckCountry() = %v, %v, %v, want %v, %v, %v", gotName, gotIso, gotErr, tt.wantName, tt.wantIso, tt.wantErr)
			}
		})
	}
}
