package handler

import (
	"assignment2/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore" //Firestore-specific support
	"google.golang.org/api/iterator"
)

// Collection name in firestore
const collection = "registrations"

// Firebase context and client used by Firestore functions throughout the program
var ctx context.Context
var client *firestore.Client

// Sets the frestore client
func SetFirestoreClient(c context.Context, cli *firestore.Client) {
	ctx = c
	client = cli
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handlePostRegistration(w, r)
	default:
		log.Println("Unsupported request method " + r.Method)
		http.Error(w, "Unsupported request method "+r.Method, http.StatusMethodNotAllowed)
		return
	}
}

// Handler for registering a new dashboard configuration, which get sendt to Firestore as a document
func handlePostRegistration(w http.ResponseWriter, r *http.Request) {

	// Instantiate decoder
	decoder := json.NewDecoder(r.Body)

	// Ensure parser fails on unknown fields (baseline way of detecting different structs than expected ones)
	decoder.DisallowUnknownFields()

	// Empty registration struct to populate
	registration := utils.Registration{}

	// Decode registration instance
	err := decoder.Decode(&registration)
	if err != nil {
		log.Println("Error: decoding JSON", err)
		http.Error(w, "Error: decoding JSON, Invalid inputs "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check if both fields are empty
	if registration.Country == "" && registration.Isocode == "" {
		log.Println("Invalid input: Fields 'Country' and 'Isocode' are empty")
		http.Error(w, "Invalid input: Fields 'Country' and 'Isocode' are empty."+
			"\n Suggestion: Fill both or one of the fields to register a dashboard", http.StatusBadRequest)

	} else { // Validate country and isocode fields
		validIsocode, validCountry, err := handleValidCountryAndCode(w, registration)
		if err != nil {
			log.Println("Invalid input: Fields 'Country' and or 'Isocode'", err)
			http.Error(w, "Invalid input: Fields 'Country' and or 'Isocode'", http.StatusBadRequest)
			return
		}
		if validIsocode != "" && validCountry != "" {
			exists, err := checkCountryExists(ctx, client, w, validCountry)
			if err != nil {
				http.Error(w, "Error: Internal server error. "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Check if the valid country country/isocode already exists in the Firestore database
			if exists {
				log.Println("Invalid input: Country is already registered")
				return

			} else {
				// Check if the target currencies are all valid values
				validCurrencies, err := checkValidCurrencies(w, registration)
				if err != nil {
					http.Error(w, "Error: Internal server error. "+err.Error(), http.StatusInternalServerError)
					return
				}
				if validCurrencies != nil {
					log.Println("Valid input: Field values " + validIsocode + " with " + validCountry + " will be registereded")

					registration.Country = validCountry
					registration.Isocode = validIsocode
					registration.Features.TargetCurrencies = validCurrencies

					postRegistration(w, r, registration)

					return

				} else {
					http.Error(w, "Detected invalid currency from 'targetCurrencies' field "+
						"\nSuggestion: Please have all currencies be written as valid 3-letter currency code (ISO 4217)", http.StatusBadRequest)
				}

			}

		}
	}
}

// Functon to handle country and or isocode accordingly
func handleValidCountryAndCode(w http.ResponseWriter, s utils.Registration) (string, string, error) {

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
	if strings.ToUpper(country) == strings.ToUpper(apiInfo[0].Name.Common) {
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

// Check if a specific country is already registered into the Firestore database
func checkCountryExists(ctx context.Context, client *firestore.Client, w http.ResponseWriter, country string) (bool, error) {

	field := "country"
	desiredValue := country

	// Query the collection for documents where the specified field has the desired value
	query := client.Collection(collection).Where(field, "==", desiredValue).Limit(1)
	iter := query.Documents(ctx)

	// Iterate through the documents
	doc, err := iter.Next()
	if err == iterator.Done {
		// No more documents
		log.Println("Country: " + country + " does not exist in Firestore database")
		return false, nil
	} else {
		// Document found with the specified field value, returning the document ID
		log.Println("Country: " + country + " already exists in Firestore database, with ID: " + doc.Ref.ID)
		http.Error(w, "Invalid input: "+country+" already exists in document, with ID: "+doc.Ref.ID+
			".\n Suggestion: Use 'UPDATE' or 'PUT' to change pre-existing dashboards", http.StatusConflict)
		return true, nil
	}
}

// Functon to check valid currencies accordingly
func checkValidCurrencies(w http.ResponseWriter, s utils.Registration) ([]string, error) {

	// Currencies from client input
	currencies := s.Features.TargetCurrencies

	// Initialize an empty hashmap of string-struct pairs
	uniqueCurrencies := make(map[string]struct{})

	// Construct a new slice that will only contain unique currencies
	uniqueCurrenciesSlice := make([]string, 0, len(currencies))

	// iterate through the currnecies array to remove duplicates
	for _, currency := range s.Features.TargetCurrencies {
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

func postRegistration(w http.ResponseWriter, r *http.Request, reg utils.Registration) {

	nested := reg.Features

	// Add the decoded date into Firestore
	id, _, err := client.Collection(collection).Add(ctx,
		map[string]interface{}{
			"country": reg.Country,
			"isoCode": reg.Isocode,
			"features": map[string]interface{}{
				"temperature":      nested.Temperature,
				"precipitation":    nested.Precipitation,
				"capital":          nested.Capital,
				"coordinates":      nested.Coordinates,
				"population":       nested.Population,
				"area":             nested.Area,
				"targetCurrencies": nested.TargetCurrencies,
			},
		})
	// Return the associated ID and when the configuration was last changed if the configuration was registered successfully
	if err != nil {
		http.Error(w, "Failed to add document to Firestore", http.StatusInternalServerError)
		log.Println("Failed to add document to Firestore:", err)
		return
	} else {

		// Returns document ID and time last updated in body
		regResponse := utils.RegResponse{
			ID:         id.ID,
			LastChange: time.Now(),
		}

		// Encode the repsonse in JSON format
		responseJSON, err := json.Marshal(regResponse)
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
