package handler

import (
	"assignment2/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

	if isEmptyField(dashboard.Country) && isEmptyField(dashboard.IsoCode) {
		http.Error(w, "Invalid input: Fields 'Country' and 'Isocode' are empty."+
			"\n Suggestion: Fill both or one of the fields to register a dashboard", http.StatusBadRequest)
		return
	}

	validCountry, validIso, err := checkCountry(dashboard.Country, dashboard.IsoCode, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	validCurrencies := checkCurrencies(dashboard.Features.TargetCurrencies, w)

	_, checkIfMissingElements, missingElements := updatedData(&dashboard, &dashboard, w)
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
			"country": validCountry,
			"isoCode": validIso,
			"features": map[string]interface{}{
				"temperature":      dashboard.Features.Temperature,
				"precipitation":    dashboard.Features.Precipitation,
				"capital":          dashboard.Features.Capital,
				"coordinates":      dashboard.Features.Coordinates,
				"population":       dashboard.Features.Population,
				"area":             dashboard.Features.Area,
				"targetCurrencies": validCurrencies,
			},
			"lastChanged": time.Now(),
		})
	if err1 != nil {
		http.Error(w, "Failed to add document", http.StatusInternalServerError)
		return
	} else {
		response := struct {
			ID         string `json:"id"`
			Lastchange string `json:"lastChanged"`
		}{
			ID:         uniqueID,
			Lastchange: whatTimeNow(),
		}
		w.Header().Set("Content-type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "failed to encode result", http.StatusInternalServerError)
			return
		}
	}
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
		doc, err := utils.getDocumentByID(ctx, collection, dashboardID)
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

// Checks if a value is empty, returns true if it is
func isEmptyField(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return v == ""
	case *bool:
		return v == nil
	case []string:
		return len(v) == 0
	default:
		return false
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

	doc, err := getDocumentByID(ctx, collection, myId)
	if err != nil {
		documentExists = false
	}

	var docRef interface{} = nil

	if documentExists {
		//Reference to the document with specified id (will be changed later)
		docRef = client.Collection(collection).Doc(doc.Ref.ID)
	}

	//If the user puts in PUT request
	if isPut {
		var p utils.Firestore
		//Checks for missing elements from user input
		_, checkIfMissingElements, missingElements := updatedData(&p, &myObject, w)
		if checkIfMissingElements {
			http.Error(w, "Missing variables: "+strings.Join(missingElements, ", "), http.StatusBadRequest)
			return nil
		}

		var c []utils.CountryInfo
		err1 := fetchURLdata(utils.COUNTRIES_API_NAME+myObject.Country, w, &c)
		if err1 != nil {
			http.Error(w, "Failed to retrieve country: "+myObject.Country, http.StatusBadRequest)
			return nil
		}
		if myObject.IsoCode != c[0].Isocode {
			myObject.IsoCode = c[0].Isocode
		}

		myObject.Features.TargetCurrencies = checkCurrencies(myObject.Features.TargetCurrencies, w)

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
			http.Error(w, "This is a test", http.StatusInternalServerError)

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

				//Merges the data from firebase with user input (that has been written)
				final, _, _ := updatedData(&newObject, &myObject, w)

				var c []utils.CountryInfo
				err1 := fetchURLdata(utils.COUNTRIES_API_NAME+final.Country, w, &c)
				if err1 != nil {
					http.Error(w, "Failed to retrieve country: "+final.Country, http.StatusBadRequest)
					return nil
				}
				if final.IsoCode != c[0].Isocode {
					final.IsoCode = c[0].Isocode
				}

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
	}

	return nil
}

// Checks for all the elements in the struct if the input by user includes these values.
// will then replace with only the written in values, avoids multiple null values if they are not written in
// returns the object with values, a bool to check if values are missing, and a string array containing all
// names of the missing elements, to inform the user
func updatedData(newObject *utils.Firestore, myObject *utils.Firestore, w http.ResponseWriter) (*utils.Firestore, bool, []string) {
	checkIfMissingElements := false
	missingElements := make([]string, 0)
	if !isEmptyField(myObject.Country) {
		newObject.Country = myObject.Country
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Country")
	}
	if !isEmptyField(myObject.IsoCode) {
		newObject.IsoCode = myObject.IsoCode
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "IsoCode")
	}
	if !isEmptyField(myObject.Features.Area) {
		newObject.Features.Area = myObject.Features.Area
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Area")
	}
	if !isEmptyField(myObject.Features.Capital) {
		newObject.Features.Capital = myObject.Features.Capital
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Capital")
	}
	if !isEmptyField(myObject.Features.Coordinates) {
		newObject.Features.Coordinates = myObject.Features.Coordinates
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Coordinates")
	}
	if !isEmptyField(myObject.Features.Precipitation) {
		newObject.Features.Precipitation = myObject.Features.Precipitation
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Precipitation")
	}
	if !isEmptyField(myObject.Features.Temperature) {
		newObject.Features.Temperature = myObject.Features.Temperature
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Temperature")
	}
	if !isEmptyField(myObject.Features.Population) {
		newObject.Features.Population = myObject.Features.Population
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Population")
	}
	if !isEmptyField(myObject.Features.TargetCurrencies) {
		newObject.Features.TargetCurrencies = checkCurrencies(myObject.Features.TargetCurrencies, w)
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Target Currencies")
	}
	return newObject, checkIfMissingElements, missingElements

}

// Function to check if currencies are valid. Will make them capitalized, '
//
//	and exclude the currencies that do not have a valid value
func checkCurrencies(arr []string, w http.ResponseWriter) []string {

	//Making a map that contains a bool, if the element has already
	//been included or not
	uniqueCurrenciesMap := make(map[string]bool)

	//Making a new array which will contain the valid currencies
	uniqueCurrenciesArr := make([]string, 0)

	//Going through all currencies in the array
	for _, currency := range arr {
		//If the length of the currency is not 3, it is not valid,
		//and will continue to next element in array
		if len(currency) != 3 {
			continue
		}
		//Capitalizing the letters in the currency
		myCurrency := strings.ToUpper(currency)

		//If it hasn't been discovered already
		if !uniqueCurrenciesMap[myCurrency] {
			uniqueCurrenciesMap[myCurrency] = true

			//Url for currency api with said currency
			url := utils.CURRENCY_API + myCurrency
			type c struct {
				Result string `json:"result"`
			}
			var a c

			//Fetching data from currency api, and putting the data into the struct
			err := fetchURLdata(url, w, &a)
			if err != nil {
				http.Error(w, "Failed to retrieve currency", http.StatusBadRequest)
				return nil
			}
			//If result is not success (such as error), it will print message to user, and continue to next element
			if a.Result != "success" {
				log.Println("Currency: " + myCurrency + " is not valid, is being excluded")
				continue
			}
			//Appends currency to the array
			uniqueCurrenciesArr = append(uniqueCurrenciesArr, myCurrency)
		}
	}
	//Returns array with unique and valid currencies, getting rid of duplicates
	return uniqueCurrenciesArr
}

//Function that makes sure both country name and isocode matches

func checkCountry(countryName string, isoCode string, w http.ResponseWriter) (string, string, error) {

	//Creates variables for country name (if country is found), and url with country name for api
	countryNameFound := true
	var CountryWithName []utils.CountryInfo
	countryUrl := url.QueryEscape(countryName)
	urlName := fmt.Sprintf(utils.COUNTRIES_API_NAME+"%s", countryUrl)

	//Creates variables for iso code (if country exists), and url with iso code for api
	isoCountryFound := true
	var CountryWithIso []utils.CountryInfo
	isoUrl := url.QueryEscape(isoCode)
	urlIso := fmt.Sprintf(utils.COUNTRIES_API_ISOCODE+"%s", isoUrl)

	//Fetching data from country api and putting it in a struct array
	err := fetchURLdata(urlName, w, &CountryWithName)

	//If there is no such country, bool is set to false
	if err != nil {
		countryNameFound = false
	}

	//Fetching data from country api and putting it in a struct array
	err1 := fetchURLdata(urlIso, w, &CountryWithIso)

	//If there is no such country, bool is set to false
	if err1 != nil {
		isoCountryFound = false
	}

	//If countryname is a valid country, returns said data from Api with country name
	if countryNameFound {
		return CountryWithName[0].Name.Common, CountryWithName[0].Isocode, nil
		//If countryname is not valid, but isocode is, it will return data from Api with Iso code
	} else if !countryNameFound && isoCountryFound {
		return CountryWithIso[0].Name.Common, CountryWithIso[0].Isocode, nil
		//Else, it will return blank strings, and an error message
	} else {
		return "", "", errors.New("no valid countries")
	}

}
