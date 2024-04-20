package handler

import (
	"assignment2/utils"
	"encoding/json"
	"log"
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
// A function that returns a map of status codes
func urlStatuses(urls []string) map[string]int {
	//The map that stores the status codes using url as string index
	statusCodes := make(map[string]int)

	for _, url := range urls {
		//Fetches the url
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error while checking status for %s: %v\n", url, err)
			statusCodes[url] = http.StatusServiceUnavailable // Or any other suitable status code
			continue
		}
		defer resp.Body.Close()

		//Status code being added
		statusCodes[url] = resp.StatusCode
	}

	return statusCodes
}

// Function to handle GET requests to the status endpoint
func statusGetRequest(w http.ResponseWriter, _ *http.Request) {

	// Time the server has been running since start
	upTime := time.Since(startTime).Seconds()

	myApis := []string{
		utils.COUNTRIES_API_NAME + "Norway",
		utils.COUNTRIES_API_ISOCODE + "FR",
		utils.CURRENCY_API + "NOK",
		utils.GEOCODING_API + "Oslo&count=1",
		utils.FORECAST_API + "latitude=59.91&longitude=10.75&hourly=temperature_2m,precipitation&forecast_days=1",
		"https://console.firebase.google.com/project/prog2005-assignment2-ee93a/firestore/databases/-default-/data/~2Fwebhooks",
	}

	statusCodes := urlStatuses(myApis)

	numOfWebhooks, _ := GetWebhookSize(func() ([]*firestore.DocumentSnapshot, error) {
		return client.Collection(collectionWebhooks).Documents(ctx).GetAll()
	})

	//result struct with status codes for Apis. Contains version and uptime as well
	statusStruct := utils.Status{
		Countriesapi:   statusCodes[utils.COUNTRIES_API_NAME+"Norway"],
		Meteoapi:       statusCodes[utils.GEOCODING_API+"Oslo&count=1"],
		Currencyapi:    statusCodes[utils.CURRENCY_API+"NOK"],
		Notificationdb: statusCodes["https://console.firebase.google.com/project/prog2005-assignment2-ee93a/firestore/databases/-default-/data/~2Fwebhooks"],
		Webhooks:       numOfWebhooks,
		Version:        "v1",
		Uptime:         upTime,
	}
	// Set the content-type to be json
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Convert status struct to JSON format
	statusJSON, err := json.Marshal(statusStruct)
	if err != nil {
		http.Error(w, "Unable to marshal status to JSON", http.StatusInternalServerError)
		return
	}
	// Write JSON response to the response body
	w.Write(statusJSON)

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
