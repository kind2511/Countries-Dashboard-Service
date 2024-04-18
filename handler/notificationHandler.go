package handler

import (
	"assignment2/utils"
	structs "assignment2/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"google.golang.org/api/iterator"
)

var webhooks = []structs.WebhookRegistration{}

func NotificationHandler(w http.ResponseWriter, r *http.Request) {
	// Find id
	url := r.URL.Path
	urlParts := strings.Split(url, "/")
	id := urlParts[len(urlParts)-1]

	switch r.Method {
	case http.MethodPost:
		postWebhook(w, r)
		/*webhook := structs.WebhookRegistration{}
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

		}*/

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

	if isEmptyField(hook.Url) || isEmptyField(hook.Country) || isEmptyField(hook.Event) {
		http.Error(w, "Not all elements are included", http.StatusBadRequest)
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
