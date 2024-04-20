package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Checks if a value is empty, returns true if it is
func IsEmptyField(value interface{}) bool {
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

// Function to return the time now, in string format
func WhatTimeNow() string {
	currentTime := time.Now()
	timeLayout := "20060102 15:04" //YYYYMMDD HH:mm

	formattedTime := currentTime.Format(timeLayout)
	return formattedTime
}

// Checks for all the elements in the struct if the input by user includes these values.
// will then replace with only the written in values, avoids multiple null values if they are not written in
// returns the object with values, a bool to check if values are missing, and a string array containing all
// names of the missing elements, to inform the user
func UpdatedData(newObject *Firestore, myObject *Firestore, w http.ResponseWriter) (*Firestore, bool, []string) {
	checkIfMissingElements := false
	missingElements := make([]string, 0)
	if !IsEmptyField(myObject.Country) {
		newObject.Country = myObject.Country
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Country")
	}
	if !IsEmptyField(myObject.IsoCode) {
		newObject.IsoCode = myObject.IsoCode
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "IsoCode")
	}
	if !IsEmptyField(myObject.Features.Area) {
		newObject.Features.Area = myObject.Features.Area
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Area")
	}
	if !IsEmptyField(myObject.Features.Capital) {
		newObject.Features.Capital = myObject.Features.Capital
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Capital")
	}
	if !IsEmptyField(myObject.Features.Coordinates) {
		newObject.Features.Coordinates = myObject.Features.Coordinates
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Coordinates")
	}
	if !IsEmptyField(myObject.Features.Precipitation) {
		newObject.Features.Precipitation = myObject.Features.Precipitation
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Precipitation")
	}
	if !IsEmptyField(myObject.Features.Temperature) {
		newObject.Features.Temperature = myObject.Features.Temperature
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Temperature")
	}
	if !IsEmptyField(myObject.Features.Population) {
		newObject.Features.Population = myObject.Features.Population
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Population")
	}
	if !IsEmptyField(myObject.Features.TargetCurrencies) {
		newObject.Features.TargetCurrencies = CheckCurrencies(myObject.Features.TargetCurrencies, w)
	} else {
		checkIfMissingElements = true
		missingElements = append(missingElements, "Target Currencies")
	}
	return newObject, checkIfMissingElements, missingElements

}

// Function to check if currencies are valid. Will make them capitalized, '
//
//	and exclude the currencies that do not have a valid value
func CheckCurrencies(arr []string, w http.ResponseWriter) []string {

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
			url := CURRENCY_API + myCurrency
			type c struct {
				Result string `json:"result"`
			}
			var a c

			//Fetching data from currency api, and putting the data into the struct
			err := FetchURLdata(url, w, &a)
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

func CheckCountry(countryName string, isoCode string, w http.ResponseWriter) (string, string, error) {

	//Creates variables for country name (if country is found), and url with country name for api
	countryNameFound := true
	var CountryWithName []CountryInfo
	countryUrl := url.QueryEscape(countryName)
	urlName := fmt.Sprintf(COUNTRIES_API_NAME+"%s", countryUrl)

	//Creates variables for iso code (if country exists), and url with iso code for api
	isoCountryFound := true
	var CountryWithIso []CountryInfo
	isoUrl := url.QueryEscape(isoCode)
	urlIso := fmt.Sprintf(COUNTRIES_API_ISOCODE+"%s", isoUrl)

	//Fetching data from country api and putting it in a struct array
	err := FetchURLdata(urlName, w, &CountryWithName)

	//If there is no such country, bool is set to false
	if err != nil {
		countryNameFound = false
	}

	//Fetching data from country api and putting it in a struct array
	err1 := FetchURLdata(urlIso, w, &CountryWithIso)

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

/*
Retrieves data from URL
*/

func FetchURLdata(myData string, w http.ResponseWriter, data interface{}) error {

	//If the fetched data is from an API
	response, err := http.Get(myData)
	if err != nil {
		return errors.New("failed to fetch url: " + myData)
	}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return errors.New("failed to decode url: " + myData)
	}
	//If the fetched data is from a JSON file
	return nil

}
