package handler

import (
	structs "assignment2/utils"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"
)

var webhooks = []structs.WebhookRegistration{}

func NotificationHandler(w http.ResponseWriter, r *http.Request) {
	// Find id
	url := r.URL.Path
	urlParts := strings.Split(url, "/")
	id := urlParts[len(urlParts)-1]

	switch r.Method {
	case http.MethodPost:
		webhook := structs.WebhookRegistration{}
		err := json.NewDecoder(r.Body).Decode(&webhook)
		if err != nil {
			http.Error(w, "Something went wrong"+err.Error(), http.StatusBadRequest)
		}
		if webhook.Event == "REGISTER" {
			// Firebase for å lage persistent storage
			webhooks = append(webhooks, webhook)
			id := uuid.New()
			fmt.Println("ID: " + id.String())

			idStruct := structs.WebhookRegistrationResponse{
				Id: id.String(),
			}

			response, err := json.Marshal(idStruct)
			if err != nil {
				// error
			}

			log.Println("Webhook " + webhook.Url + " has been registered")

			http.Error(w, string(response), http.StatusCreated)

		}

	case http.MethodDelete:
		println("Delete " + id)

		// Finn i DB og slett

		http.Error(w, "", http.StatusNoContent)

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
				// error
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

func createWebhookResponse(id string, url string) []byte {
	getStruct := structs.WebhookGetResponse{
		Id:      id,
		Url:     url,
		Country: "NO",
		Event:   "INVOKE",
	}

	response, err := json.Marshal(getStruct)
	if err != nil {
		// error
	}
	return response
}
