package handler

import (
	structs "assignment2/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var webhooks = []structs.WebhookRegistration{}

func NotificationHandler(w http.ResponseWriter, r *http.Request) {
	// Find id
	url := r.URL.Path
	urlParts := strings.Split(url, "/")
	id := urlParts[len(urlParts)-1]

	switch r.Method {
	case http.MethodPost: //registration of webhooks
		RegisterWebhook(r, w)

	case http.MethodDelete:
		deleteWebhook(w, r)

	case http.MethodGet:
		if id == "" {
			var getAllNotifications = []structs.WebhookGetResponse{}

			getStruct := structs.WebhookGetResponse{
				Id:      id,
				Url:     url,
				Country: "NO",
				Event:   "INVOKE",
			}

			getAllNotifications = append(getAllNotifications, getStruct)

			response, err := json.Marshal(getAllNotifications)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			http.Error(w, string(response), http.StatusOK)

		} else {
			response := createWebhookResponse(id, url)
			http.Error(w, string(response), http.StatusOK)
		}

	default:
		http.Error(w, "Method "+r.Method+" not supported for "+structs.NOTIFICATION_PATH, http.StatusMethodNotAllowed)

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

// Adds new webhook registration to firebase
func RegisterWebhook(r *http.Request, w http.ResponseWriter) {
	webhook := structs.WebhookRegistration{}
	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		http.Error(w, "Something went wrong"+err.Error(), http.StatusBadRequest)
	}
	if webhook.Event == "REGISTER" {

		webhooks = append(webhooks, webhook)
		id := uuid.New()
		fmt.Println("ID: " + id.String())

		idStruct := structs.WebhookRegistrationResponse{
			Id: id.String(),
		}

		response, err := json.Marshal(idStruct)
		if err != nil {
			http.Error(w, "Something went wrong"+err.Error(), http.StatusBadRequest)

		}

		log.Println("Webhook " + webhook.Url + " has been registered")

		http.Error(w, string(response), http.StatusCreated)

	}
}

func createWebhookResponse(id string, url string) []byte {
	getStruct := structs.WebhookGetResponse{
		Id:      id,
		Url:     url,
		Country: "NO",
		Event:   "INVOKE",
	}

	response, err := json.Marshal(getStruct)
	if err != nil {
		var errorResponse []byte = nil
		return errorResponse
	}
	return response
}
