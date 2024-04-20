package handler

import (
	"assignment2/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
)

// StatusHandler handles the status endpoint
func StatusHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			statusGetRequest(w, r)
		default:
			http.Error(w, "Method not supported. Currently only GET is supported.", http.StatusNotImplemented)
			return
		}
	}
}

var startTime = time.Now()

// Function to check if an error occured while making an HTTP request
func checkHTTPError(err error) {
	if err != nil {
		err = fmt.Errorf("error occurred while making HTTP request: %v", err)
		fmt.Println(err)
		return
	}
}

// Function to handle GET requests to the status endpoint
func statusGetRequest(w http.ResponseWriter, _ *http.Request) {

	// Time the server has been running since start
	upTime := time.Since(startTime).Seconds()

	// Gets status-code for Countries API, Meteo API, Currency API and Notification DB
	countriesApiStatusCode, meteoApiStatusCode, currencyApiStatusCode, notificationStatus := GetStatusCode()

	// get the number of webhooks in the database
	// Creates status struct
	Status := createFormat(countriesApiStatusCode, meteoApiStatusCode, currencyApiStatusCode, notificationStatus, upTime)

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

func GetStatusCode() (*http.Response, *http.Response, *http.Response, *http.Response) {
	countriesApiStatusCode, err := http.Get(utils.COUNTRIES_API + "name/Norway")
	checkHTTPError(err)

	meteoApiStatusCode, err := http.Get(utils.GEOCODING_API + "Norway&count=1&language=en&format=json")
	checkHTTPError(err)

	currencyApiStatusCode, err := http.Get(utils.CURRENCY_API + "/nok")
	checkHTTPError(err)

	notificationStatus, err := http.Get("https://console.firebase.google.com/project/prog2005-assignment2-ee93a/firestore/databases/-default-/data/~2Fwebhooks")
	checkHTTPError(err)
	return countriesApiStatusCode, meteoApiStatusCode, currencyApiStatusCode, notificationStatus
}

func createFormat(countriesApiStatusCode *http.Response, meteoApiStatusCode *http.Response, currencyApiStatusCode *http.Response, notificationStatus *http.Response, upTime float64) utils.Status {

	numOfWebhooks, _ := GetWebhookSize(func() ([]*firestore.DocumentSnapshot, error) {
		return client.Collection(collectionWebhooks).Documents(ctx).GetAll()
	})

	Status := utils.Status{
		Countriesapi:   countriesApiStatusCode.StatusCode,
		Meteoapi:       meteoApiStatusCode.StatusCode,
		Currencyapi:    currencyApiStatusCode.StatusCode,
		Notificationdb: notificationStatus.StatusCode,
		Webhooks:       numOfWebhooks,
		Version:        "v1",
		Uptime:         upTime,
	}
	return Status
}

// name of collection used for webhooks
const collectionWebhooks = "webhooks"

type FirestoreClient interface {
	Collection(path string) *firestore.CollectionRef
}

type DocumentRetriever func() ([]*firestore.DocumentSnapshot, error)

// Function for getting the number of webhooks in the database
func GetWebhookSize(getDocuments DocumentRetriever) (int, error) {
	webhooks, err := getDocuments()
	if err != nil {
		return -1, err
	}
	return len(webhooks), nil
}
