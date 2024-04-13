package handler

import (
	"assignment2/utils"
	structs "assignment2/utils"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"errors"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)


func NotificationHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		postWebhook(w, r)
	case http.MethodDelete:
		DeleteWebhook(w, r)
	case http.MethodGet:
		getWebHooks(w, r)
	default:
		http.Error(w, "Method "+r.Method+" not supported for "+structs.NOTIFICATION_PATH, http.StatusMethodNotAllowed)
	}
}

// Function to delete a webhook
func DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	// split the url to get the id
	elem := strings.Split(r.URL.Path, "/")

	// Check if the id is provided
	if elem[4] == "" || len(elem) != 5 {
		http.Error(w, "Invalid lenght og the url path, include the id for the webhook you want to delete", http.StatusBadRequest)
		return
	}

	// Get the id of the webhook
	webhooksID := elem[4]

	// Delete a webhook with the spesified id from the firestore database
	webhook := client.Collection(collectionWebhooks).Doc(webhooksID)
	_, err := webhook.Get(ctx)
	if err != nil {
		http.Error(w, "No webhooks of that ID can be found", http.StatusNotFound)
		return
	}

	// Delete webhook from storage
	_, status := webhook.Delete(ctx)
	if status != nil {
		http.Error(w, "Error while deleting webhook", http.StatusInternalServerError)
		return
	}

	// Return success message
	http.Error(w, "The webhook have been successfully deleted", http.StatusOK)
}



func ValidateEvent(e string) bool {
	return e == "REGISTER" || e == "INVOKE" || e == "CHANGE" || e == "DELETE"
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

	if isEmptyField(hook.Url) || isEmptyField(hook.Country) || isEmptyField(hook.Event) {
		http.Error(w, "Not all elements are included", http.StatusBadRequest)
		return
	}

	if !ValidateEvent(hook.Event) {
		http.Error(w, "Event is not added in correctly", http.StatusBadRequest)
		return
	}

	a, err := http.Get(structs.COUNTRIES_API_ISOCODE + hook.Country)
	if err != nil {
		http.Error(w, "Failed to check for country data", http.StatusBadRequest)
		return
	}
	if a.StatusCode == http.StatusBadRequest {
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

	_, _, err1 := client.Collection(webCollection).Add(ctx,
		map[string]interface{}{
			"id":      uniqueID,
			"url":     hook.Url,
			"country": hook.Country,
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