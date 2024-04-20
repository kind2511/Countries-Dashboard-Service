package handler

import (
	"assignment2/utils"
	structs "assignment2/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func NotificationHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			postWebhook(w, r)
		case http.MethodDelete:
			deleteWebhook(w, r)
		case http.MethodGet:
			getWebHooks(w, r)
		default:
			http.Error(w, "Method "+r.Method+" not supported for "+structs.NOTIFICATION_PATH, http.StatusMethodNotAllowed)
		}
	}
}

// Function to delete a webhook by its ID
func deleteWebhook(w http.ResponseWriter, r *http.Request) {
	// Extract dashboard ID from URL
	elem := strings.Split(r.URL.Path, "/")
	webhookID := elem[4]

	if len(webhookID) != 0 {
		// Get reference to the document
		docRef := client.Collection(collectionWebhooks).Doc(webhookID)

		// Delete the document
		_, err := docRef.Delete(ctx)
		if err != nil {
			log.Println("Error deleting document:", err)
			http.Error(w, "Error deleting document", http.StatusInternalServerError)
			return
		}

		// Return success message
		w.WriteHeader(http.StatusNoContent)
	} else {
		// If Dashboard ID is not provided
		http.Error(w, "Webhook ID not provided", http.StatusBadRequest)
		return
	}
}

func ValidateEvent(e string) bool {
	return e == "REGISTER" || e == "INVOKE" || e == "CHANGE" || e == "DELETE"
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func postWebhook(w http.ResponseWriter, r *http.Request) {

	webCollection := "webhooks"
	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()

	var hook utils.WebhookRegistration
	err := decoder.Decode(&hook)
	if err != nil {
		http.Error(w, "failed decoding JSON", http.StatusInternalServerError)
		return
	}

	if utils.IsEmptyField(hook.Url) || utils.IsEmptyField(hook.Event) {
		http.Error(w, "Not all needed elements are included", http.StatusBadRequest)
		return
	}

	if !ValidateEvent(hook.Event) {
		http.Error(w, "Event is not added in correctly", http.StatusBadRequest)
		return
	}

	if strings.HasPrefix(hook.Url, "http://localhost:") {
		fmt.Println("This is a test")
		substring := hook.Url[len("http://localhost:"):]
		if len(substring) < 5 {
			http.Error(w, "Localhost url is not valid", http.StatusBadRequest)
			return

		} else if substring[4] != '/' {
			http.Error(w, "Localhost url is not valid", http.StatusBadRequest)
			return
		}

		valid := true
		for i := 0; i < 4; i++ {
			fmt.Println(substring[i])
			if i >= len(substring) || !isDigit(substring[i]) {
				valid = false
				break
			}
		}

		if !valid {
			http.Error(w, "Localhost url is not valid", http.StatusBadRequest)
			return
		}

	} else {
		fmt.Println("Another test")
		check, _ := http.Get(hook.Url)
		if check.StatusCode != http.StatusOK {
			http.Error(w, "Url provided is not valid", http.StatusBadRequest)
			return
		}
	}
	a, err := http.Get(structs.COUNTRIES_API_ISOCODE + hook.Country)
	if err != nil {
		http.Error(w, "Failed to check for country data", http.StatusBadRequest)
		return
	}
	// Have it so that you can register a webhook with empty country field
	if a.StatusCode == http.StatusBadRequest && hook.Country != "" {
		http.Error(w, "isocode: "+hook.Country+" is not valid", http.StatusBadRequest)
		return
	}

	a.Body.Close()

	var uniqueID string

	for {
		uniqueID = utils.GenerateUID(5)

		// Check if the generated ID already exists in a document
		iter := client.Collection(collection).Where("id", "==", uniqueID).Limit(1).Documents(ctx)

		doc, err := iter.Next()
		if err == iterator.Done {
			// No document found with current ID, continue with further processing
			break
		}
		if err != nil {
			log.Println("Error retrieving document:", err)
			break
		}
		if doc != nil {
			// ID already exists, generating a new one
			log.Println("ID already exists...generating new one")
			continue
		}
	}

	isocode := strings.ToUpper(hook.Country)

	_, _, err1 := client.Collection(webCollection).Add(ctx,
		map[string]interface{}{
			"id":      uniqueID,
			"url":     hook.Url,
			"country": isocode,
			"event":   hook.Event,
		})
	if err1 != nil {
		return
	} else {
		response := struct {
			ID string `json:"id"`
		}{
			ID: uniqueID,
		}
		w.Header().Set("Content-type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "failed to encode result", http.StatusInternalServerError)
			return
		}
	}

}

// Function to retrieve document data and write JSON response
func retrieveWebHookData(w http.ResponseWriter, doc *firestore.DocumentSnapshot) {
	// Map document data to WebhookGetResponse struct
	var document utils.WebhookGetResponse
	if err := doc.DataTo(&document); err != nil {
		log.Println("Error retrieving document data:", err)
		http.Error(w, "Error retrieving document data", http.StatusInternalServerError)
		return
	}

	// Marshal the document to JSON
	jsonData, err := json.Marshal(document)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON data to the response
	if _, err := w.Write(jsonData); err != nil {
		log.Println("Error writing JSON response:", err)
		http.Error(w, "Error writing JSON response", http.StatusInternalServerError)
		return
	}
}

// Gets one webhook based on its Firestore ID. If no ID is provided it gets all webhooks
func getWebHooks(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		http.Error(w, "Error retrieving document", http.StatusInternalServerError)
		return
	}

	const webhookCollection = "webhooks"
	// Extract webhook ID from URL
	elem := strings.Split(r.URL.Path, "/")
	webhookID := elem[4]

	if len(webhookID) != 0 {
		// Query documents where the 'id' field matches the provided webhookID
		query := client.Collection(webhookCollection).Where("id", "==", webhookID).Limit(1)
		iter := query.Documents(ctx)

		// Retrieve reference to document
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				// Document not found
				errorMessage := "Document with ID " + webhookID + " not found"
				http.Error(w, errorMessage, http.StatusNotFound)
				return
			}
			// If trouble retrieving document
			log.Println("Error retrieving document:", err)
			http.Error(w, "Error retrieving document", http.StatusInternalServerError)
			return
		}

		// Retrieves document and writes JSON response
		retrieveWebHookData(w, doc)
	} else {
		// Collective retrieval of documents
		iter := client.Collection(webhookCollection).Documents(ctx)

		for {
			doc, err := iter.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				log.Printf("Failed to iterate: %v", err)
				return
			}

			// Retrieves document and writes JSON response
			retrieveWebHookData(w, doc)
		}
	}
}

/*
Handles the invocation of events
*/
func invocationHandler(w http.ResponseWriter, event string, country string) {

	switch event {
	case "REGISTER":
		triggerEvent(w, event, country)

	case "CHANGE":
		triggerEvent(w, event, country)

	case "DELETE":
		triggerEvent(w, event, country)

	case "INVOKE":
		triggerEvent(w, event, country)

	default:
		http.Error(w, "No webhook invocation", http.StatusNotFound)
	}
}

/*
Checks if the event meets the conditions to trigger
*/
func checkWebhook(isocode string) bool {

	webhookCollection := "webhooks"

	query := client.Collection(webhookCollection).Where("country", "in", []string{isocode, ""})
	iter := query.Documents(ctx)

	// Retrieve reference to document
	_, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			// Document not found
			return false
		}
		// If trouble retrieving document
		log.Println("Error retrieving document:", err)
		return false
	} else {
		// iterate through webhooks and trigger based on country and event conditions
		return true
	}
}

/*
Handle the triggered event
*/
func triggerEvent(w http.ResponseWriter, event string, country string) {

	webhookCollection := "webhooks"

	// retrieve the webhooks which will be triggered by the conditions
	query := client.Collection(webhookCollection).
		Where("event", "==", event).
		Where("country", "in", []string{country, ""})

	iter := query.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break // finished iterating
		}
		if err != nil {
			log.Println("Error iterating over webhook documents: ", err)
			http.Error(w, "Error iterating over webhook documents ", http.StatusInternalServerError)

		}

		log.Println(event + " event triggered...")

		//  Message body when invoking url
		var hook utils.WebhookInvokeMessage
		if err := doc.DataTo(&hook); err != nil {
			log.Println("Error retrieving document data: ", err)
			http.Error(w, "Error retrieving document data ", http.StatusInternalServerError)
		}

		// Call the url
		go callUrl(w, hook)
	}

}

/*
Calls given URL with given content and awaits response (status and body)
*/
func callUrl(w http.ResponseWriter, hook utils.WebhookInvokeMessage) {
	url := hook.Url
	event := hook.Event
	country := hook.Country

	log.Println("Attempting invocation of URL " + url + " of content Country:'" +
		country + "' and Event:'" + event + "'.")

	// Check if URL service is OK
	ok, err := http.Head(url)
	if err != nil {
		if ok.StatusCode != http.StatusOK {
			log.Println("Invalid URL")
		}
	}

	// Set current time
	timeNow := utils.WhatTimeNow()
	hook.Time = timeNow

	// Convert webhook message to JSON bytes
	payload, err := json.Marshal(hook)
	if err != nil {
		log.Println("Error marshaling JSON: ", err)
		http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Error creating HTTP request: ", err)
		http.Error(w, "Error creating HTTP request", http.StatusInternalServerError)
		return
	}

	// Set content-Type header
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error in HTTP request. Error: ", err)
		http.Error(w, "Error in HTTP request", http.StatusInternalServerError)
		return
	}

	// Read the response
	response, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading invocation response: ", err)
		http.Error(w, "Error reading invocation response", http.StatusInternalServerError)
	}

	log.Println("Webhook " + url + " invoked. Received status code " +
		strconv.Itoa(res.StatusCode) + " and body: " + string(response))
}
