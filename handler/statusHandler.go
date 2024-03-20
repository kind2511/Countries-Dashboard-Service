package handler

import (
	"assignment2/utils"
	"encoding/json"
	"net/http"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {

	// Creates status struct (hardcoded now)
	Status := utils.Status{
		Countriesapi:   1,
		Meteoapi:       1,
		Currencyapi:    1,
		Notificationdb: 1,
		Webhooks:       1,
		Version:        "v1",
		Uptime:         1,
	}

	// Set the content-type to be json
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Convert status struct to JSON format
	statusJSON, err := json.Marshal(Status)
	if err != nil {
		http.Error(w, "Unable to marshal status to JSON", http.StatusInternalServerError)
		return
	}

	// Write JSON response to the response body
	w.Write(statusJSON)

}
