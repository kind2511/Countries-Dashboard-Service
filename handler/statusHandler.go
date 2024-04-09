package handler

import (
	"assignment2/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var startTime = time.Now()

func StatusHandler(w http.ResponseWriter, r *http.Request) {

	// Time the server has been running since start
	upTime := time.Since(startTime).Seconds()

	// Gets status-code for Countries API
	countriesApiStatusCode, err := http.Get(utils.COUNTRIES_API + "/name/norway")
	if err != nil {
		err = fmt.Errorf("error occured while making HTTP request: %v", err)
		fmt.Println(err)
		return
	}

	// Gets status-code for Meteo API
	meteoApiStatusCode, err := http.Get(utils.GEOCODING_API + "Norway&count=1&language=en&format=json")
	if err != nil {
		err = fmt.Errorf("error occured while making HTTP request: %v", err)
		fmt.Println(err)
		return
	}

	// Gets status-code for Currency API
	currencyApiStatusCode, err := http.Get(utils.CURRENCY_API + "/nok")
	if err != nil {
		err = fmt.Errorf("error occured while making HTTP request: %v", err)
		fmt.Println(err)
		return
	}

	notificationStatus, err := http.Get("https://console.firebase.google.com/project/prog2005-assignment2-ee93a/firestore/databases/-default-/data/~2FDashboard~2FXhjRXlZcd7uLTMd8kdxb")
	if err != nil {
		err = fmt.Errorf("error occured while making HTTP request: %v", err)
		fmt.Println(err)
		return
	}

	// Creates status struct (hardcoded now)
	Status := utils.Status{
		Countriesapi:   countriesApiStatusCode.StatusCode,
		Meteoapi:       meteoApiStatusCode.StatusCode,
		Currencyapi:    currencyApiStatusCode.StatusCode,
		Notificationdb: notificationStatus.StatusCode,
		Webhooks:       1,
		Version:        "v1",
		Uptime:         upTime,
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
