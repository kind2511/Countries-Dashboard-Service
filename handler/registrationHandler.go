package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// Firebase context and client used by Firestore functions throughout the program.
var ctx context.Context
var client *firestore.Client

// sets the firestore client
func SetFirestoreClient(c context.Context, cli *firestore.Client) {
    ctx = c
    client = cli
}

// name of collection used for dashboards
const collection = "Dashboard"

/*
Handler for all registration-related operations
*/
func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		postRegistration(w, r)
	case http.MethodGet:
		getDashboards(w, r)
    case http.MethodDelete:
        deleteDashboard(w, r)
	default:
		log.Println("Unsupported request method" + r.Method)
		http.Error(w, "Unsupported request method"+r.Method, http.StatusMethodNotAllowed)
		return
	}
}

//TODO

// do stub testing

// Handler for registering a new dashboard configuration, which get sendt to Firestore as a document
func postRegistration(w http.ResponseWriter, r *http.Request) {
	// check if the country or ISO code has already been registered and is valid
	// if it has, return a message to suggest using put/patch to update the information

	// Country name can be empty if ISO code field is filled and vice versa

	// Check if the input felts and format is registered correctly
	// ERROR: tell which ones, StatusBadRequest

	// check if the target currencies are valid

	// Return the associated ID and when the configuration was last changed if the configuration was registered successfully
	//StatusCreated

}

// Gets one dashboard based on its Firestore ID. If no ID is provided it gets all dashboards
func getDashboards(w http.ResponseWriter, r *http.Request) {
    // Test for embedded dashboard ID from URL
    elem := strings.Split(r.URL.Path, "/")
    dashboardId := elem[4]

    if len(dashboardId) != 0 {
        // Retrieve specific document based on ID
        res := client.Collection(collection).Doc(dashboardId)

        // Retrieve reference to document
        doc, err := res.Get(ctx)
        if err != nil {
            log.Println("Error retrieving document:", err)
            http.Error(w, "Error retrieving document", http.StatusInternalServerError)
            return
        }

        // Get the entire document data
        docData := doc.Data()

        // Add the document ID to the data
        docData["id"] = doc.Ref.ID

        // Marshal the entire document data to JSON
        jsonData, err := json.Marshal(docData)
        if err != nil {
            log.Println("Error marshaling JSON:", err)
            http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
            return
        }

        // Set the Content-Type header to application/json
        w.Header().Set("Content-Type", "application/json")

        // Write the JSON data to the response
        _, err = w.Write(jsonData)
        if err != nil {
            log.Println("Error writing JSON response:", err)
            http.Error(w, "Error writing JSON response", http.StatusInternalServerError)
            return
        }
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

            // Get the entire document data
            docData := doc.Data()

            // Add the document ID to the data
            docData["id"] = doc.Ref.ID

            // Marshal the entire document data to JSON
            jsonData, err := json.Marshal(docData)
            if err != nil {
                log.Println("Error marshaling JSON:", err)
                http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
                return
            }

            // Set the Content-Type header to application/json
            w.Header().Set("Content-Type", "application/json")

            // Write the JSON data to the response
            _, err = w.Write(jsonData)
            if err != nil {
                log.Println("Error writing JSON response:", err)
                http.Error(w, "Error writing JSON response", http.StatusInternalServerError)
                return
            }
        }
    }
}

// Deletes a specific dashboard based on its Firestore ID
func deleteDashboard(w http.ResponseWriter, r *http.Request) {
    // Extract dashboard ID from URL
    elem := strings.Split(r.URL.Path, "/")
    dashboardID := elem[4]

    if len(dashboardID) != 0 {
        // Get reference to the document
        docRef := client.Collection(collection).Doc(dashboardID)

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
        http.Error(w, "Dashboard ID not provided", http.StatusBadRequest)
        return
    }
}
