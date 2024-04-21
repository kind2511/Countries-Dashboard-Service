package handler

import (
	"assignment2/utils"
	structs "assignment2/utils"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// name of collection used for webhooks
const webhookCollection = "webhooks"

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
		doc, err := GetDocumentByID(ctx, webhookCollection, webhookID)
		if err != nil {
			// Document not found
			errorMessage := "Document with ID " + webhookID + " not found"
			http.Error(w, errorMessage, http.StatusNotFound)
			return
		}

		// Delete the document
		_, err = doc.Ref.Delete(ctx)
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

// POST requests being handled with this function
func postWebhook(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()

	//Creating a struct variable, and assigning values to this struct with data provided by user
	var hook utils.WebhookRegistration
	err := decoder.Decode(&hook)
	if err != nil {
		http.Error(w, "failed decoding JSON", http.StatusInternalServerError)
		return
	}

	//If url or event is empty, error is returned
	if utils.IsEmptyField(hook.Url) || utils.IsEmptyField(hook.Event) {
		http.Error(w, "Url or Event is not included", http.StatusBadRequest)
		return
	}

	//if event is not REGISTER, CHANGE, INVOKE or DELETE, returns error
	if !utils.ValidateEvent(hook.Event) {
		http.Error(w, "Event is not added in correctly", http.StatusBadRequest)
		return
	}

	/*
		Here is where it checks the url.
		Checks for either localhost urls or regular urls
		For localhost urls, it checks if it contains the string http://localhost:XXXX/
		at the start, where XXXX needs to be digits, and not other characters
	*/
	if strings.HasPrefix(hook.Url, "http://localhost:") {
		//Checks if the substring contains at least 5 characters
		substring := hook.Url[len("http://localhost:"):]
		if len(substring) < 5 {
			http.Error(w, "Localhost url is not valid", http.StatusBadRequest)
			return

			//If the fifth character in the substring is not '/', it returns error
		} else if substring[4] != '/' {
			http.Error(w, "Localhost url is not valid", http.StatusBadRequest)
			return
		}

		//Bool to see if url is valid
		valid := true
		//Runs through the four characters, to see if the port contains only digits
		//Returns false if 1 of them does not contain a digit
		for i := 0; i < 4; i++ {
			if i >= len(substring) || !utils.IsDigit(substring[i]) {
				valid = false
				break
			}
		}

		//If not valid
		if !valid {
			http.Error(w, "Localhost url is not valid", http.StatusBadRequest)
			return
		}
	}

	//If it is a regular url, it does a HEAD request,
	//and if it returns status code other than 200, error is returned
	check, _ := http.Head(hook.Url)
	if check.StatusCode != http.StatusOK {
		http.Error(w, "Url provided is not valid", http.StatusBadRequest)
		return

	}
	//Checks if country is valid
	a, _ := http.Get(structs.COUNTRIES_API_ISOCODE + hook.Country)
	// If country is not valid, country is set to blank (works also if country is not written in)
	if a.StatusCode != http.StatusOK {
		hook.Country = ""
	}

	a.Body.Close()

	//Unique id for webhook
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

	//Capitalizes isocode
	isocode := strings.ToUpper(hook.Country)

	//Adds document to webhook collection in firestore with data
	_, _, err1 := client.Collection(webhookCollection).Add(ctx,
		map[string]interface{}{
			"id":      uniqueID,
			"url":     hook.Url,
			"country": isocode,
			"event":   hook.Event,
		})
	if err1 != nil {
		return
	} else {
		//Response to user with id that is given to webhook
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
func invocationHandler(w http.ResponseWriter, event string, isocode string) bool {
	if !checkWebhook(isocode) {
		return false
	}

	if event == "REGISTER" || event == "CHANGE" || event == "DELETE" || event == "INVOKE" {

		// retrieve the webhooks which will be triggered by the conditions
		query := client.Collection(webhookCollection).
			Where("event", "==", event).
			Where("country", "in", []string{isocode, ""})

		iter := query.Documents(ctx)
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break // finished iterating
			}
			if err != nil {
				log.Println("Error iterating over webhook documents: ", err)
				http.Error(w, "Error iterating over webhook documents ", http.StatusInternalServerError)
				return false
			}

			log.Println(event + " event triggered...")

			//  Message body when invoking url
			var hook utils.WebhookInvokeMessage
			if err := doc.DataTo(&hook); err != nil {
				log.Println("Error retrieving document data: ", err)
				http.Error(w, "Error retrieving document data ", http.StatusInternalServerError)
				return false
			}

			// Call the url
			go callUrl(w, hook)
		}
	}
	return false
}

/*
Calls given URL with given content and awaits response (status and body)
*/
func callUrl(w http.ResponseWriter, hook utils.WebhookInvokeMessage) {
	url := hook.Url
	event := hook.Event
	country := hook.Country

	log.Println("Attempting invocation of URL " + url + " by Country:'" +
		country + "' and Event:'" + event)

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
		return
	}
	defer res.Body.Close()

	// Read the response
	response, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading invocation response: ", err)
		return
	}

	log.Println("Webhook " + url + " invoked. Received status code " +
		strconv.Itoa(res.StatusCode) + " and body: " + string(response))
}

/*
Checks if the event meets the conditions to trigger @ the individual handles and their methods
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
