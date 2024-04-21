package handler

import (
	"assignment2/utils"
	"encoding/json"
	"fmt"
	"net/http"
)

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type header
	w.Header().Add("content-type", "application/json")

	// Write header message
	_, err := fmt.Fprintf(w, "This service has the following endpoints with methods:\n\n")
	if err != nil {
		http.Error(w, "Error when returning output", http.StatusInternalServerError)
	}

	// Functions to format output and write it
	Registrationfunction(w)
	DashboardFunction(w)
	NotificationFunction(w)
	StatusFunction(w)

}

// Format and write status endpoint data
func StatusFunction(w http.ResponseWriter) {
	// Define type for output
	type OutputsStatus []utils.DefaultEndpointStruct

	// Define data
	outputStatus := OutputsStatus{
		utils.DefaultEndpointStruct{
			Url:         utils.STATUS_PATH,
			Method:      "GET",
			Description: "Monitoring service availability"},
	}

	// Marshall data into JSON with proper indentation
	jsonDataStatus, err := json.MarshalIndent(outputStatus, "", "\t")
	if err != nil {
		http.Error(w, "Error converting to JSON", http.StatusInternalServerError)
	}

	// Write message to response
	_, err2 := fmt.Fprintf(w, "`\n\nStatus endpoint:\n")
	if err2 != nil {
		http.Error(w, "Error when returning output", http.StatusInternalServerError)
	}
	// Write JSON data to the response
	w.Write(jsonDataStatus)
}

// Format and write notification endpoint data
func NotificationFunction(w http.ResponseWriter) {
	// Define type for output
	type OutputsNotifications []utils.DefaultEndpointStruct

	// Define data
	outputNotification := OutputsNotifications{
		utils.DefaultEndpointStruct{
			Url:         utils.NOTIFICATION_PATH,
			Method:      "POST",
			Description: "Registration of Webhook"},
		utils.DefaultEndpointStruct{
			Url:         utils.NOTIFICATION_PATH + "{id}",
			Method:      "DELETE",
			Description: "Deletion of Webhook"},
		utils.DefaultEndpointStruct{
			Url:         utils.NOTIFICATION_PATH + "{id}",
			Method:      "GET",
			Description: "View specific registered Webhook"},
		utils.DefaultEndpointStruct{
			Url:         utils.NOTIFICATION_PATH,
			Method:      "GET",
			Description: "View all registered Webhooks "},
		utils.DefaultEndpointStruct{
			Url:         "url specified in the cooresponding webhook registration",
			Method:      "POST",
			Description: "Webhook invocation upon trigger"},
	}

	// Marshall data into JSON with proper indentation
	jsonDataNotification, err := json.MarshalIndent(outputNotification, "", "\t")
	if err != nil {
		http.Error(w, "Error converting to JSON", http.StatusInternalServerError)
	}

	// Write message to response
	_, err2 := fmt.Fprintf(w, "\n\nNotification endpoint:\n")
	if err2 != nil {
		http.Error(w, "Error when returning output", http.StatusInternalServerError)
	}
	// Write message to response
	w.Write(jsonDataNotification)
}

// Format and write dashboard enpoint data
func DashboardFunction(w http.ResponseWriter) {
	// Define type for output
	type OutputsDashboard []utils.DefaultEndpointStruct

	// Define data
	outputDashboard := OutputsDashboard{
		utils.DefaultEndpointStruct{
			Url:         utils.DASHBOARD_PATH + "{id}",
			Method:      "GET",
			Description: "Retrieve populated dashboard"},
	}

	// Marshall data into JSON with proper indentation
	jsonDataDashboard, err := json.MarshalIndent(outputDashboard, "", "\t")
	if err != nil {
		http.Error(w, "Error converting to JSON", http.StatusInternalServerError)
	}

	// Write message to response
	_, err2 := fmt.Fprintf(w, "\n\nDashboard endpoint:\n")
	if err2 != nil {
		http.Error(w, "Error when returning output", http.StatusInternalServerError)
	}
	// Write data to response
	w.Write(jsonDataDashboard)
}

// Format and write registration endpoint data
func Registrationfunction(w http.ResponseWriter) {
	// Define type for output
	type OutputsRegistration []utils.DefaultEndpointStruct

	// Define data
	outputRegistration := OutputsRegistration{
		utils.DefaultEndpointStruct{
			Url:         utils.REGISTRATION_PATH,
			Method:      "POST",
			Description: "Registration of new dashboard"},
		utils.DefaultEndpointStruct{
			Url:         utils.REGISTRATION_PATH + "{id}",
			Method:      "GET",
			Description: "View a specific registered dashboard configuration"},
		utils.DefaultEndpointStruct{
			Url:         utils.REGISTRATION_PATH,
			Method:      "GET",
			Description: "View all registered dashboard configurations"},
		utils.DefaultEndpointStruct{
			Url:         utils.REGISTRATION_PATH + "{id}",
			Method:      "PUT",
			Description: "Replace a specific registered dashboard configuration"},
		utils.DefaultEndpointStruct{
			Url:         utils.REGISTRATION_PATH + "{id}",
			Method:      "PATCH",
			Description: "Replace only parts of the registered dashboard configration"},
		utils.DefaultEndpointStruct{
			Url:         utils.REGISTRATION_PATH + "{id}",
			Method:      "DELETE",
			Description: "Delete a specific registered dashboard configuration"},
	}

	// Marshall data into JSON with proper indentation
	jsonDataRegistration, err := json.MarshalIndent(outputRegistration, "", "\t")
	if err != nil {
		http.Error(w, "Error converting to JSON", http.StatusInternalServerError)
	}

	// Write message to response
	_, err2 := fmt.Fprintf(w, "\n\nRegistration endpoint:\n")
	if err2 != nil {
		http.Error(w, "Error when returning output", http.StatusInternalServerError)
	}
	// Write data to response
	w.Write(jsonDataRegistration)
}
