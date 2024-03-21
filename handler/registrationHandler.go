package handler

import (
	"log"
	"net/http"
)

// Struct for registration
type Registration struct {
	Country  string   `json:"country"`
	Isocode  string   `json:"isocode"`
	Features Features `json:"features"`
}

type Features struct {
	Temperature      bool     `json:"temperature"` // Note: In degrees Celsius
	Precipitation    bool     `json:"precipitation"`
	Capital          bool     `json:"capital"`
	Coordinates      bool     `json:"coordinates"`
	Population       bool     `json:"population"`
	Area             bool     `json:"area"`
	TargetCurrencies []string `json:"targetcurrencies"`
}

/*
Handler for all registration-related operations
*/
func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		postRegistration(w, r)
	default:
		log.Println("Unsupported request method" + r.Method)
		http.Error(w, "Unsupported request method"+r.Method, http.StatusMethodNotAllowed)
		return
	}
}

//TODO

// do stub testing

// Handler for registering a new dashboard configuration, which get sendt to Firestore as a document
func postRegistration(w http.ResponseWriter, r *http.Request) {
	// check if the country or ISO code has already been registered and is valid
	// if it has, return a message to suggest using put/patch to update the information

	// Country name can be empty if ISO code field is filled and vice versa

	// Check if the input felts and format is registered correctly
	// ERROR: tell which ones, StatusBadRequest

	// check if the target currencies are valid

	// Return the associated ID and when the configuration was last changed if the configuration was registered successfully
	//StatusCreated

}
