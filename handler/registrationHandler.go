package handler

import (
	"log"
	"net/http"
)

// name of collection used for dashboards
const collection = "Dashboard"

/*
Handler for all registration-related operations
*/
func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		postRegistration(w, r)
	case http.MethodGet:
		getSpecificDashboardConfiguration(w, r)
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

func getSpecificDashboardConfiguration(w http.ResponseWriter, r *http.Request) {
	
}

