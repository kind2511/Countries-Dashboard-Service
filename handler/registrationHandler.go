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

/*
Handler for registering a new dashboard configuration, which get sendt to Firestore as a document
*/
func postRegistration(w http.ResponseWriter, r *http.Request) {

	// Instantiate decoder
	decoder := json.NewDecoder(r.Body)

	// Ensure parser fails on unknown fields (baseline way of detecting different structs than expected ones)
	decoder.DisallowUnknownFields()

	// Empty registration struct to populate
	dashboard := utils.Dashboard{}

	// Decode registration instance
	err := decoder.Decode(&dashboard)
	if err != nil {
		log.Println("Error: decoding JSON", err)
		http.Error(w, "Error: decoding JSON, Invalid inputs "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check if both fields are empty
	if dashboard.Country == "" && dashboard.Isocode == "" {
		log.Println("Invalid input: Fields 'Country' and 'Isocode' are empty")
		http.Error(w, "Invalid input: Fields 'Country' and 'Isocode' are empty."+
			"\n Suggestion: Fill both or one of the fields to register a dashboard", http.StatusBadRequest)

	} else { // Validate country and isocode fields
		validIsocode, validCountry, err := handleValidCountryAndCode(w, dashboard)
		if err != nil {
			log.Println("Invalid input: Fields 'Country' and or 'Isocode'", err)
			http.Error(w, "Invalid input: Fields 'Country' and or 'Isocode'", http.StatusBadRequest)
			return
		}

		// If there is returned valid country and isocode
		if validIsocode != "" && validCountry != "" {

			// Check if the target currencies are all valid values
			validCurrencies, err := checkValidCurrencies(w, dashboard)
			if err != nil {
				http.Error(w, "Error: Internal server error. "+err.Error(), http.StatusInternalServerError)
				return
			}
			if validCurrencies != nil {
				log.Println("Valid input: Field values 'country': " + validCountry + " matching 'isocode': " + validIsocode + " will be registereded")

				dashboard.Country = validCountry
				dashboard.Isocode = validIsocode
				dashboard.RegFeatures.TargetCurrencies = validCurrencies

				// Post registered dashboard
				handlePostRegistration(w, dashboard)

				return

			} else {
				http.Error(w, "Detected invalid currency from 'targetCurrencies' field "+
					"\nSuggestion: Please have all currencies be written as valid 3-letter currency code (ISO 4217)", http.StatusBadRequest)
			}

		}
	}
}

/*
Functon to handle country and or isocode accordingly
*/
func handleValidCountryAndCode(w http.ResponseWriter, s utils.Dashboard) (string, string, error) {

	// country and isocode registration input from client
	country := s.Country
	isocode := s.Isocode

	// Empty slice to hold the data from requested country
	var apiInfo []utils.CountryInfo

	// Variables for response body and error
	var res *http.Response
	var err error

	// Fetch countries URL depending on which field has been filled and are valid.
	if country != "" || isocode != "" {

		// HEAD request to the countries API endpoint via country input
		url := utils.COUNTRIES_API_NAME + country
		res, err = http.Head(url)
		if err != nil {
			log.Println("Invalid URL: Invalid country input ", err)
		}

		// Check if the response status code is OK, then GET the URL body
		if res.StatusCode == http.StatusOK {
			log.Println("Valid 'countries' URL with country: Processing")
			res, err = http.Get(url)

		} else { // if the country input is invalid, check the isocode input

			// HEAD request to the countries API endpoint via isocode input
			url := utils.COUNTRIES_API_ISOCODE + isocode
			res, err = http.Head(url)
			if err != nil {
				log.Println("Invalid URL: Invalid isocode input ", err)
			}

			// Check if the response status code is OK, then GET the URL body
			if res.StatusCode == http.StatusOK {
				log.Println("Valid 'countries' URL with isocode: Processing")
				res, err = http.Get(url)
			}
		}
	}
	// Check if country info URL is valid
	if err != nil {
		log.Println("Error: Failed to fetch country info from URL: ", err)
		http.Error(w, "Error: Failed to fetch country info from URL", http.StatusInternalServerError)
		return "", "", err
	}

	// Decode the response
	if err := json.NewDecoder(res.Body).Decode(&apiInfo); err != nil {
		log.Println("Error: decoding JSON", err)
		return "", "", err
	}

	// Handle when either the country or isocode input is a valid value matching the API's value
	if strings.EqualFold(country, apiInfo[0].Name.Common) {
		// Making sure that empty/ invalid isocode field is returned with the matching isocode from the country input
		if isocode != "" || isocode == "" {
			return apiInfo[0].Isocode, apiInfo[0].Name.Common, nil
		}
	} else {
		log.Println("Invalid value: Field 'country': " + country + ", does not match " + apiInfo[0].Name.Common + " found")
		http.Error(w, "Invalid value: Field 'country' or 'isocode'", http.StatusBadRequest)
		return "", "", nil

	}

	// Making sure that empty/ invalid country field is returned with the matching country from the isocode input
	if strings.ToUpper(isocode) == apiInfo[0].Isocode {
		if country != "" || country == "" {
			return apiInfo[0].Isocode, apiInfo[0].Name.Common, nil

		}
	} else {
		log.Println("Invalid value: Field 'isocode': " + isocode + ", does not match" + apiInfo[0].Isocode + "found")
		http.Error(w, "Invalid value: Field 'country' or 'isocode'", http.StatusBadRequest)
		return "", "", nil

	}

	// If no match found
	http.Error(w, "Invalid input: Field 'country' and or 'isocode'", http.StatusBadRequest)
	log.Println("Invalid input: Field 'country' and or 'isocode'")
	return "", "", nil
}

/*
Functon to check valid currencies accordingly
*/
func checkValidCurrencies(w http.ResponseWriter, d utils.Dashboard) ([]string, error) {

	// Currencies from client input
	currencies := d.RegFeatures.TargetCurrencies

	// Initialize an empty hashmap of string-struct pairs
	uniqueCurrencies := make(map[string]struct{})

	// Construct a new slice that will only contain unique currencies
	uniqueCurrenciesSlice := make([]string, 0, len(currencies))

	// iterate through the currnecies array to remove duplicates
	for _, currency := range d.RegFeatures.TargetCurrencies {
		// Skip empty strings
		if currency == "" {
			continue
		}
		// Check if the currency is already present in the map
		if _, found := uniqueCurrencies[currency]; !found {
			// If not found, add the currency to the map and slice
			uniqueCurrencies[currency] = struct{}{}
			uniqueCurrenciesSlice = append(uniqueCurrenciesSlice, currency)
		}
	}
	log.Println("All possible duplicates, if any, has been removed from", currencies, " to ", uniqueCurrenciesSlice)

	// Iterate through each currency and see if they are valid
	for _, currency := range uniqueCurrenciesSlice {

		url := utils.CURRENCY_API + currency

		// Send a Get request to the currency API endpoint
		res, err := http.Get(url)
		if err != nil {
			// Handle error if the request fails
			log.Println("Error: checking currency validity. ", err)
			http.Error(w, "Error: checking currency validity", http.StatusInternalServerError)
			return nil, err
		}

		// Check if the response status code is OK
		if res.StatusCode == http.StatusOK {
			log.Println("Valid currency:", currency)
		} else { // return false when encountering invalid currency
			log.Println("Invalid currency:", currency)
			return nil, err
		}
	}

	return uniqueCurrenciesSlice, nil
}

func handlePostRegistration(w http.ResponseWriter, d utils.Dashboard) {

	nested := d.RegFeatures

	// Current formatted time
	timeNow := whatTimeNow2()

	// Add the decoded date into Firestore
	id, _, err := client.Collection(collection).Add(ctx,
		map[string]interface{}{
			"country": d.Country,
			"isoCode": d.Isocode,
			"features": map[string]interface{}{
				"temperature":      nested.Temperature,
				"precipitation":    nested.Precipitation,
				"capital":          nested.Capital,
				"coordinates":      nested.Coordinates,
				"population":       nested.Population,
				"area":             nested.Area,
				"targetCurrencies": nested.TargetCurrencies,
			},
			"lastChanged": timeNow,
		})
	// Return the associated ID and when the configuration was last changed if the configuration was registered successfully
	if err != nil {
		http.Error(w, "Failed to add document to Firestore", http.StatusInternalServerError)
		log.Println("Failed to add document to Firestore:", err)
		return
	} else {

		// Returns document ID and time last updated in body
		response := utils.DashboardResponse{
			ID:         id.ID,
			LastChange: timeNow,
		}

		// Encode the repsonse in JSON format
		responseJSON, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set content type header
		w.Header().Set("Content-Type", "application/json")

		// Write the response JSON and return appropriate status code
		w.WriteHeader(http.StatusCreated)
		w.Write(responseJSON)

		log.Println("Document added to collection. Identifier of return document: ", id.ID)
		return
	}
}

/*
Function that takes the time now, and shows it in correct format
*/
func whatTimeNow2() string {
	currentTime := time.Now()
	timeLayout := "2006-01-02 15:04" //YYYYMMDD HH:mm

	formattedTime := currentTime.Format(timeLayout)
	return formattedTime

}

// function to get a document based on its id field
func getDocumentByID(ctx context.Context, collection string, dashboardID string) (*firestore.DocumentSnapshot, error) {
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

// Function to retrieve document data and write JSON response
func retrieveDocumentData(w http.ResponseWriter, doc *firestore.DocumentSnapshot) {
	// Map document data to Firestore struct
	var originalDoc utils.Firestore
	if err := doc.DataTo(&originalDoc); err != nil {
		log.Println("Error retrieving document data:", err)
		http.Error(w, "Error retrieving document data", http.StatusInternalServerError)
		return
	}

	// Create a Registration struct to create desired structure
	desiredDoc := utils.Registration{
		ID:         originalDoc.ID,
		Country:    originalDoc.Country,
		IsoCode:    originalDoc.IsoCode,
		Features:   originalDoc.Features,
		LastChange: originalDoc.LastChange.Format("20060102 15:04"), // Format timestamp
	}

	// Marshal the desired document to JSON
	jsonData, err := json.Marshal(desiredDoc)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Set the status code to 200
	w.WriteHeader(http.StatusOK)

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
		doc, err := getDocumentByID(ctx, collection, dashboardID)
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
	dashboardID := elem[4]

	if len(dashboardID) != 0 {
		doc, err := getDocumentByID(ctx, collection, dashboardID)
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
		http.Error(w, "Dashboard ID not provided", http.StatusBadRequest)
		return
	}
}