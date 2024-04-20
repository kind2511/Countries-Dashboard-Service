package handler

import (
	"assignment2/utils"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Firebase context and client used by Firestore functions throughout the program.
var ctx context.Context
var client *firestore.Client

// sets the firestore client
func SetFirestoreClient(c context.Context, cli *firestore.Client) {
	ctx = c
	client = cli
}

// function to get a document based on its id field
func GetDocumentByID(ctx context.Context, collection string, dashboardID string) (*firestore.DocumentSnapshot, error) {
	// Query documents where the 'id' field matches the provided dashboardID
	query := client.Collection(collection).Where("id", "==", dashboardID).Limit(1)
	iter := query.Documents(ctx)

	// Retrieve reference to document
	doc, err := iter.Next()
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// name of collection used for dashboards
const collection = "Dashboard"

/*
Handler for all registration-related operations
*/
func RegistrationHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			postRegistration(w, r)
		case http.MethodGet:
			getDashboards(w, r)
		case http.MethodPut:
			updateDashboard(w, r, true)
		case http.MethodPatch:
			updateDashboard(w, r, false)
		case http.MethodDelete:
			deleteDashboard(w, r)
		default:
			log.Println("Unsupported request method" + r.Method)
			http.Error(w, "Unsupported request method"+r.Method, http.StatusMethodNotAllowed)
			return
		}
	}
}

/*
Handler for registering a new dashboard configuration, which get sendt to Firestore as a document
*/
func postRegistration(w http.ResponseWriter, r *http.Request) {

	// Instantiate decoder
	decoder := json.NewDecoder(r.Body)

	// Ensure parser fails on unknown fields (baseline way of detecting different structs than expected ones)
	decoder.DisallowUnknownFields()

	// Empty registration struct to populate
	dashboard := utils.Firestore{}

	// Decode registration instance
	err := decoder.Decode(&dashboard)
	if err != nil {
		log.Println("Error: decoding JSON into Dashboard struct due to ", err)
		http.Error(w, "Error: decoding JSON, Invalid inputs "+err.Error(), http.StatusBadRequest)
		return
	}

	if utils.IsEmptyField(dashboard.Country) && utils.IsEmptyField(dashboard.IsoCode) {
		http.Error(w, "Invalid input: Fields 'Country' and 'Isocode' are empty."+
			"\n Suggestion: Fill both or one of the fields to register a dashboard", http.StatusBadRequest)
		return
	}

	validCountry, validIso, err := utils.CheckCountry(dashboard.Country, dashboard.IsoCode, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dashboard.Country = validCountry
	dashboard.IsoCode = validIso

	validCurrencies := utils.CheckCurrencies(dashboard.Features.TargetCurrencies, w)

	_, checkIfMissingElements, missingElements := utils.UpdatedData(&dashboard, &dashboard, w)
	if checkIfMissingElements {
		http.Error(w, "Missing variables: "+strings.Join(missingElements, ", "), http.StatusBadRequest)
		return
	}

	// Generate a random ID
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
	// Add the decoded date into Firestore
	_, _, err1 := client.Collection(collection).Add(ctx,
		map[string]interface{}{
			"id":      uniqueID,
			"country": dashboard.Country,
			"isoCode": dashboard.IsoCode,
			"features": map[string]interface{}{
				"temperature":      dashboard.Features.Temperature,
				"precipitation":    dashboard.Features.Precipitation,
				"capital":          dashboard.Features.Capital,
				"coordinates":      dashboard.Features.Coordinates,
				"population":       dashboard.Features.Population,
				"area":             dashboard.Features.Area,
				"targetCurrencies": validCurrencies,
			},
			"lastChange": time.Now(),
		})
	if err1 != nil {
		http.Error(w, "Failed to add document", http.StatusInternalServerError)
		return
	} else {
		response := struct {
			ID         string `json:"id"`
			Lastchange string `json:"lastChange"`
		}{
			ID:         uniqueID,
			Lastchange: utils.WhatTimeNow(),
		}
		w.Header().Set("Content-type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "failed to encode result", http.StatusInternalServerError)
			return
		}

		// Trigger event if registered configuration has a registered webhook to invoke
		if !checkWebhook(dashboard.IsoCode) {
			log.Println("No REGISTER event triggered...")
		} else {
			invocationHandler(w, "REGISTER", dashboard.IsoCode)
		}
	}
}

// Function to retrieve document data and write JSON response
func retrieveDocumentData(w http.ResponseWriter, doc *firestore.DocumentSnapshot) {

	// Map document data to Firestore struct
	var originalDoc utils.Dashboard_Get
	if err := doc.DataTo(&originalDoc); err != nil {
		log.Println("Error retrieving document data:", err)
		http.Error(w, "Error retrieving document data ", http.StatusInternalServerError)
		return
	}
	// Create a Registration struct to create desired structure
	response := struct {
		ID       string `json:"id"`
		Country  string `json:"country"`
		IsoCode  string `json:"isoCode"`
		Features struct {
			Temperature      bool     `json:"temperature"`
			Precipitation    bool     `json:"precipitation"`
			Capital          bool     `json:"capital"`
			Coordinates      bool     `json:"coordinates"`
			Population       bool     `json:"population"`
			Area             bool     `json:"area"`
			TargetCurrencies []string `json:"targetCurrencies"`
		} `json:"features"`
		LastChange string `json:"lastChange"`
	}{
		ID:      originalDoc.ID,
		Country: originalDoc.Country,
		IsoCode: originalDoc.IsoCode,
		Features: struct {
			Temperature      bool     `json:"temperature"`
			Precipitation    bool     `json:"precipitation"`
			Capital          bool     `json:"capital"`
			Coordinates      bool     `json:"coordinates"`
			Population       bool     `json:"population"`
			Area             bool     `json:"area"`
			TargetCurrencies []string `json:"targetCurrencies"`
		}{
			Temperature:      originalDoc.Features.Temperature,
			Precipitation:    originalDoc.Features.Precipitation,
			Capital:          originalDoc.Features.Capital,
			Coordinates:      originalDoc.Features.Coordinates,
			Population:       originalDoc.Features.Population,
			Area:             originalDoc.Features.Area,
			TargetCurrencies: originalDoc.Features.TargetCurrencies,
		},
		LastChange: originalDoc.LastChange.Format("20060102 15:04"),
	}

	// Marshal the desired document to JSON
	jsonData, err := json.Marshal(response)
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

// Gets one dashboard based on its Firestore ID. If no ID is provided it gets all dashboards
func getDashboards(w http.ResponseWriter, r *http.Request) {
	// Extract dashboard ID from URL
	elem := strings.Split(r.URL.Path, "/")
	dashboardID := elem[4]

	if len(dashboardID) != 0 {
		doc, err := GetDocumentByID(ctx, collection, dashboardID)
		if err != nil {
			if err == iterator.Done {
				// Document not found
				errorMessage := "Document with ID " + dashboardID + " not found"
				http.Error(w, errorMessage, http.StatusNotFound)
				return
			}
			// If trouble retrieving document
			log.Println("Error retrieving document:", err)
			http.Error(w, "Error retrieving document", http.StatusInternalServerError)
			return
		}

		// Retrieves document and writes JSON response
		retrieveDocumentData(w, doc)
	} else {
		// Collective retrieval of documents
		iter := client.Collection(collection).Documents(ctx)

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
			retrieveDocumentData(w, doc)
		}
	}
}

// Deletes a specific dashboard based on its 'id' field
func deleteDashboard(w http.ResponseWriter, r *http.Request) {
	// Extract dashboard ID from URL
	elem := strings.Split(r.URL.Path, "/")

	if len(elem) < 5 {
		http.Error(w, "Dashboard ID not provided", http.StatusBadRequest)
		return
	}

	dashboardID := elem[4]

	if len(dashboardID) != 0 {
		doc, err := GetDocumentByID(ctx, collection, dashboardID)
		if err != nil {
			//if err == iterator.Done {
			if status.Code(err) == codes.NotFound {
				// Document not found
				errorMessage := "Document with ID " + dashboardID + " not found"
				http.Error(w, errorMessage, http.StatusNotFound)
				return
			}
			// If trouble retrieving document
			log.Println("Error retrieving document:", err)
			http.Error(w, "Error retrieving document", http.StatusInternalServerError)
			return
		}

		// Retrieve isocode value from the document for webhook event
		isocode, ok := doc.Data()["isoCode"].(string)
		if !ok {
			log.Println("Error accessing isocode field")
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

		// Trigger event if deleted configuration has a registered webhook to invoke
		if !checkWebhook(isocode) {
			log.Println("No DELETE event triggered...")
		} else {
			invocationHandler(w, "DELETE", isocode)
		}

	} else {
		// If Dashboard ID is not provided
		http.Error(w, "Dashboard ID not provided", http.StatusBadRequest)
		return
	}
}

// Function that updates a dashboard. Works as both PUT and PATCH, depending on bool given
func updateDashboard(w http.ResponseWriter, r *http.Request, isPut bool) error {

	//Fetching ID from URL
	myId := r.URL.Path[len(utils.REGISTRATION_PATH):]

	//Creates the object variable, with data stored from the user input
	var myObject utils.Firestore
	if err := json.NewDecoder(r.Body).Decode(&myObject); err != nil {
		return err
	}

	//Time is set to now
	myObject.LastChange = time.Now()

	documentExists := true

	doc, err := GetDocumentByID(ctx, collection, myId)
	if err != nil {
		documentExists = false
	}

	var docRef interface{} = nil

	if documentExists {
		//Reference to the document with specified id (will be changed later)
		docRef = client.Collection(collection).Doc(doc.Ref.ID)
	}

	// Retrieve current isocode value from the document for webhook event
	isocode, ok := doc.Data()["isoCode"].(string)
	if !ok {
		log.Println("Error accessing isocode field")
	}

	//If the user puts in PUT request
	if isPut {
		validCountry, validIso, err := utils.CheckCountry(myObject.Country, myObject.IsoCode, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		myObject.Country = validCountry
		myObject.IsoCode = validIso
		var p utils.Firestore
		//Checks for missing elements from user input
		_, checkIfMissingElements, missingElements := utils.UpdatedData(&p, &myObject, w)
		if checkIfMissingElements {
			http.Error(w, "Missing variables: "+strings.Join(missingElements, ", "), http.StatusBadRequest)
			return nil
		}

		myObject.Features.TargetCurrencies = utils.CheckCurrencies(myObject.Features.TargetCurrencies, w)

		//taking the data from the object and applies them to a map
		data := map[string]interface{}{
			"country": myObject.Country,
			"isoCode": myObject.IsoCode,
			"features": map[string]interface{}{
				"temperature":      myObject.Features.Temperature,
				"precipitation":    myObject.Features.Precipitation,
				"capital":          myObject.Features.Capital,
				"coordinates":      myObject.Features.Coordinates,
				"population":       myObject.Features.Population,
				"area":             myObject.Features.Area,
				"targetCurrencies": myObject.Features.TargetCurrencies,
			},
			"lastChange": time.Now(),
		}
		if docRef != nil {
			if firestoreDocRef, ok := docRef.(*firestore.DocumentRef); ok {

				//Creates a new object, that fetches data from firebase
				var newObject utils.Firestore

				//Fetching the document, checks if it exists
				doc, err := firestoreDocRef.Get(ctx)
				if err != nil {
					if status.Code(err) == codes.NotFound {
						http.Error(w, "Document does not exist, cannot PATCH", http.StatusBadRequest)
						return nil
					}
				}
				doc.DataTo(&newObject)

				data["id"] = newObject.ID

				_, err1 := firestoreDocRef.Set(ctx, data)
				if err1 != nil {
					http.Error(w, "failed to update data", http.StatusInternalServerError)
					return nil

				}
			}
		} else {
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
			data["id"] = uniqueID
			_, _, err := client.Collection(collection).Add(ctx, data)
			if err != nil {
				http.Error(w, "Failed to add new document: "+err.Error(), http.StatusInternalServerError)
				return nil
			}

		}

		// Trigger event if registered configuration has a registered webhook to invoke
		if !checkWebhook(isocode) {
			log.Println("No CHANGE event triggered...")
		} else {
			invocationHandler(w, "CHANGE", isocode)
		}

		//If user put in a PATCH request
	} else {
		if docRef != nil {
			if firestoreDocRef, ok := docRef.(*firestore.DocumentRef); ok {

				//Creates a new object, that fetches data from firebase
				var newObject utils.Firestore

				//Fetching the document, checks if it exists
				doc, err := firestoreDocRef.Get(ctx)
				if err != nil {
					if status.Code(err) == codes.NotFound {
						http.Error(w, "Document does not exist, cannot PATCH", http.StatusBadRequest)
						return nil
					}
				}
				doc.DataTo(&newObject)

				validCountry, validIso, err := utils.CheckCountry(myObject.Country, myObject.IsoCode, w)
				//If it turns out that country name or isocode provided in the PATCH request are valid, it will change both variables.
				//Otherwise, it will not
				if err == nil {
					myObject.Country = validCountry
					myObject.IsoCode = validIso
				}

				//Merges the data from firebase with user input (that has been written)
				final, _, _ := utils.UpdatedData(&newObject, &myObject, w)

				final.Features.TargetCurrencies = utils.CheckCurrencies(final.Features.TargetCurrencies, w)

				//Updates the document
				_, err = firestoreDocRef.Set(ctx, map[string]interface{}{
					"id":      newObject.ID,
					"country": final.Country,
					"isoCode": final.IsoCode,
					"features": map[string]interface{}{
						"temperature":      final.Features.Temperature,
						"precipitation":    final.Features.Precipitation,
						"capital":          final.Features.Capital,
						"coordinates":      final.Features.Coordinates,
						"population":       final.Features.Population,
						"area":             final.Features.Area,
						"targetCurrencies": final.Features.TargetCurrencies,
					},
					"lastChange": time.Now(),
				})
				if err != nil {
					http.Error(w, "Failed to patch", http.StatusInternalServerError)
					return err
				}
			}
		} else {
			http.Error(w, "Document does not exist", http.StatusBadRequest)
			return nil
		}

		// Trigger event if registered configuration has a registered webhook to invoke
		if !checkWebhook(isocode) {
			log.Println("No CHANGE event triggered...")
		} else {
			invocationHandler(w, "CHANGE", isocode)
		}

	}

	return nil
}
