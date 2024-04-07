package handler

import (
	"assignment2/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

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
		http.Error(w, "Failed to decode request body due to: "+err.Error(), http.StatusBadRequest)
		log.Println("Error decoding JSON ", err)
		return
	}

	// Validate country and isocode fields
	if registration.Country == "" && registration.Isocode == "" {
		log.Println("Both the country and isocode are empty")
		http.Error(w, "Country and isocode fields are empty, please input both or one of them to register a dashboard", http.StatusBadRequest)

	} else {
		validIsocode, validCountry, err := handleValidCountryAndCode(w, registration)
		if err != nil {
			log.Println("Something went wrong when validating country and isocode", err)
			http.Error(w, "Something went wrong when validating country and isocode", http.StatusBadRequest)
			return
		}

		if validIsocode != "" && validCountry != "" {
			log.Println("Is a valid input:", validIsocode, validCountry)
			//postRegistration(w, registration)
		}

		// Check if country or isocode already exists in the Firestore database
		exists, err := checkCountryExists(ctx, client, w, validCountry)
		if err != nil {
			http.Error(w, "Error checking country: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if exists {
			http.Error(w, "Use PUT or UPDATE if change is wanted", http.StatusConflict)
			return
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

	var res *http.Response
	var err error

	// Fetch URL depending on which field has been filled and are valid, if both are, use the country value
	if country != "" || isocode != "" {
		// HEAD request to check if the country field is valid without doing a GET request
		res, err = http.Head(utils.COUNTRIES_API_NAME + country)
		if err != nil {
			log.Println("The country field is an invalid path for the URL")
			http.Error(w, "Invalid country URL, checking for ISOCODE", http.StatusContinue)
		}
		// if the HEAD request is OK, GET the country info URL by country name
		if res.StatusCode == http.StatusOK {
			http.Error(w, "valid country using country for registration", http.StatusContinue)
			log.Println(utils.COUNTRIES_API_NAME + country)
			res, err = http.Get(utils.COUNTRIES_API_NAME + country)

		} else { // if the country value is invalid, check the isocode value
			log.Println("URL for country is not reachable")

			// Make a HEAD request to check if the isocode field is valid without doing a GET request
			res, err = http.Head(utils.COUNTRIES_API_ISOCODE + isocode)
			if err != nil {
				log.Println("The isocode field is an invalid path for the URL")
				http.Error(w, "Invalid isocode for URL", http.StatusBadRequest)
			}
			// if the HEAD request is OK, GET the country info URL by isocode
			if res.StatusCode == http.StatusOK {
				log.Println(utils.COUNTRIES_API_ISOCODE + isocode)
				http.Error(w, "valid isocode using ISOCODE for registration", http.StatusContinue)
				res, err = http.Get(utils.COUNTRIES_API_ISOCODE + isocode)
			} else {
				log.Println("URL for isocode is not reachable")
				http.Error(w, "Invalid isocode and or country code", http.StatusBadRequest)
			}
		}
	}

	// Check if country info URL is valid
	if err != nil {
		http.Error(w, "Failed to fetch country info", http.StatusInternalServerError)
		log.Println("Failed to fetch country info:", err)
		return "", "", err
	}

	// Decode the response
	if err := json.NewDecoder(res.Body).Decode(&apiInfo); err != nil {
		http.Error(w, "Failed to decode country info", http.StatusInternalServerError)
		log.Println("Failed to decode country info:", err)
		return "", "", err
	}

	// Check if the country name or ISO code is valid or empty, handle accordingly
	if strings.ToUpper(country) == strings.ToUpper(apiInfo[0].Name.Common) {
		// Makes sure that even if the ISO code field is empty/ invalid, it get corrected to fit with the valid country value
		if isocode != "" || isocode == "" {
			http.Error(w, "Valid country name input: "+apiInfo[0].Isocode+" and "+apiInfo[0].Name.Common+" has be registered", http.StatusBadRequest)
			return apiInfo[0].Isocode, apiInfo[0].Name.Common, nil
		}
	} else {
		log.Println("invalid country input did not pass")
	}

	// Makes sure that even if the country field is empty/ invalid, it get corrected to fit with the valid ISO code value
	if strings.ToUpper(isocode) == apiInfo[0].Isocode {
		if country != "" || country == "" {
			http.Error(w, "Valid isocode input: "+apiInfo[0].Isocode+" and "+apiInfo[0].Name.Common+" will be registered", http.StatusBadRequest)
			return apiInfo[0].Isocode, apiInfo[0].Name.Common, nil

		}
	} else {
		log.Println("invalid isocode did not pass")
	}

	// If no match found
	http.Error(w, "Invalid country or ISO code", http.StatusBadRequest)
	log.Println("Invalid country or ISO code")
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
		log.Println("Country: " + country + " does not exist")
		http.Error(w, "Country: "+country+" does not exist", http.StatusContinue)
		return false, nil
	}

	// Document found with the specified field value
	log.Println("Country: " + country + " already exists")
	http.Error(w, "Country: "+country+" exists in document with ID: "+doc.Ref.ID, http.StatusBadRequest)
	return true, nil
}

/*
func postRegistration(w http.ResponseWriter, r utils.Registration) {

	// Add the decoded date into Firestore
	id, _, err := client.Collection(collection).Add(ctx,
		map[string]interface{}{
			"country": r.Country,
			"isoCode": r.Isocode,
			"features": map[string]interface{}{
				"temperature":      r.Features.Temperature,
				"precipitation":    r.Features.Precipitation,
				"capital":          r.Features.Capital,
				"coordinates":      r.Features.Coordinates,
				"population":       r.Features.Population,
				"area":             r.Features.Area,
				"targetCurrencies": r.Features.TargetCurrencies,
			},
		})
	// Return the associated ID and when the configuration was last changed if the configuration was registered successfully
	if err != nil {
		http.Error(w, "Failed to add document to Firestore", http.StatusInternalServerError)
		log.Println("Failed to add document to Firestore:", err)
		return
	} else {
		// Returns document ID and time last updated in body
		log.Println("Document added to collection. Identifier of return document: ", id.ID)
		http.Error(w, id.ID, http.StatusCreated)
		return
	}

}
*/
