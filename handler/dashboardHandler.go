package handler

import (
	"assignment2/utils"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Own float type, either float32 or float64, whatever we see fit
type myFloat float32

/*
The dashboard that is received from dashboards,
checking if certain data should be implemented and fetched
*/
type Recieved_Dashboard struct {
	Id       int    `json:"id"`
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
}

/*
Struct that will display the information in each dasahboard
*/
type OutputDashboardWithData struct {
	Country  string `json:"country"`
	IsoCode  string `json:"isoCode"`
	Features struct {
		Temperature      myFloat            `json:"temperature"`
		Precipitation    myFloat            `json:"precipitation"`
		Capital          string             `json:"capital"`
		Coordinates      Coordinates        `json:"coordinates"`
		Population       int                `json:"population"`
		Area             myFloat            `json:"area"`
		TargetCurrencies map[string]myFloat //`json:"targetCurrencies"`
	} `json:"features"`
	LastRetrieval string `json:"lastRetrieval"`
}

// Coordinates struct that contains latitude and longitude
type Coordinates struct {
	Latitude  myFloat `json:"latitude"`
	Longitude myFloat `json:"longitude"`
}

// Handler function that checks if method is set to GET
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		DashboardFunc(w, r)
	}
}

/*
This function will receive data from json, with the dashboards, check what variables will show values

	and then return the specific values
*/
func DashboardFunc(w http.ResponseWriter, r *http.Request) error {

	//Finding out what ID is written in the URL path
	myId := r.URL.Path[len(utils.DASHBOARD_PATH):]

	var inputData []Recieved_Dashboard

	//Fetching data from json file, which contains dashboards
	err := fetchURLdata("test-data.json", false, w, &inputData)
	if err != nil {
		return err
	}

	//Checking and converting the id to int, which will be used for fetching object
	myIdInt, err := strconv.Atoi(myId)
	if err != nil {
		http.Error(w, "'"+myId+"'"+" is not a valid id", http.StatusBadRequest)
		return err
	}
	//Small function that returns object based on matching object Id
	getObjectByID := func(id int) *Recieved_Dashboard {
		for _, obj := range inputData {
			if obj.Id == id {
				return &obj
			}
		}
		return nil
	}
	//Getting object from an id. If nil, error is returned
	myObject := getObjectByID(myIdInt)
	if myObject == nil {
		http.Error(w, "Object with Id "+"'"+myId+"' not found", http.StatusBadRequest)
		return err
	}

	//Fetching variables from functions

	//Fetching population, capital, their own currency and are
	population, capital, countryCurrency, area, err := retrieveCountryData(myObject.Country, w, r)
	if err != nil {
		return err
	}

	//Fetching coordinates from chosen capital
	longitude, latitude, err := retrieveCoordinates(capital, w, r)
	if err != nil {
		return err
	}

	//Fetching temperature and precipitation using coordinates
	temperature, precipitation, err := retrieveWeather(longitude, latitude, w, r)
	if err != nil {
		return err
	}

	//Creating the result struct
	var Result OutputDashboardWithData

	//Assigning the values to the result struct
	Result.Country = myObject.Country
	Result.IsoCode = myObject.IsoCode

	//Checks if a value is to be displayed, and then assigns the values if true
	//-------------------------------------------------------------------------------------------
	if myObject.Features.Temperature {
		Result.Features.Temperature, err = floatFormat(temperature)
		if err != nil {
			return err
		}
	}
	if myObject.Features.Precipitation {
		Result.Features.Precipitation, err = floatFormat(precipitation)
		if err != nil {
			return err
		}
	}
	if myObject.Features.Capital {
		Result.Features.Capital = capital
	}
	if myObject.Features.Coordinates {
		Result.Features.Coordinates.Longitude, err = floatFormat(longitude)
		if err != nil {
			return err
		}
		Result.Features.Coordinates.Latitude, err = floatFormat(latitude)
		if err != nil {
			return err
		}
	}
	if myObject.Features.Area {
		Result.Features.Area, err = floatFormat(area)
		if err != nil {
			return err
		}
	}
	if myObject.Features.Population {
		Result.Features.Population = population
	}
	//---------------------------------------------------------------------

	//Making own map to set currencies with their exchangerates
	c := make(map[string]myFloat)

	//Makes the map with exchange rates for fetched currency earlier
	currencyRates, err := retrieveCurrencyExchangeRates(countryCurrency, w, r)
	if err != nil {
		return err
	}

	//Runs through all currencies that are fetched from specified object
	for _, currency := range myObject.Features.TargetCurrencies {
		//Assigns currency rate to specified currencies from the fetched object
		c[currency] = currencyRates[currency]
	}
	//Assigns map of exchange rates to result
	Result.Features.TargetCurrencies = c

	//Time for last retrieval being assigned using formatted time
	Result.LastRetrieval = whatTimeNow()

	//Sets header, and encodes the result
	w.Header().Set("Content-type", "application/json")

	if err := json.NewEncoder(w).Encode(Result); err != nil {
		http.Error(w, "Failed to encode result", http.StatusInternalServerError)
		return err
	}
	return nil
}

/*
Retrieves data from URL
*/

func fetchURLdata(myData string, isUrl bool, w http.ResponseWriter, data interface{}) error {

	//If the fetched data is from an API
	if isUrl {
		response, err := http.Get(myData)
		if err != nil {
			http.Error(w, "Failed to fetch url: "+myData, http.StatusInternalServerError)
			return err
		}
		defer response.Body.Close()
		err = json.NewDecoder(response.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Failed to decode url: "+myData, http.StatusInternalServerError)
			return err
		}
		//If the fetched data is from a JSON file
	} else {
		file, err := os.Open(myData)
		if err != nil {
			http.Error(w, "Failed to fetch file data: "+myData, http.StatusInternalServerError)
			return err
		}
		defer file.Close()
		err = json.NewDecoder(file).Decode(&data)
		if err != nil {
			http.Error(w, "Failed to decode file data: "+myData, http.StatusInternalServerError)
			return err
		}
	}
	return nil

}

/*
Function will return population, capital, currency and area on a certain country
*/
func retrieveCountryData(country string, w http.ResponseWriter, r *http.Request) (int, string, string, myFloat, error) {

	myCountry := country

	//Making a struct of elements that will be fetched from Countries API
	type Country struct {
		Population int                    `json:"population"`
		Capital    []string               `json:"capital"`
		Currency   map[string]interface{} `json:"currencies"`
		Area       myFloat                `json:"area"`
	}
	var chosenCountry []Country

	//Fetches data from specified country
	err := fetchURLdata(utils.COUNTRIES_API+"name/"+myCountry, true, w, &chosenCountry)
	if err != nil {
		return 0, "", "", 0, err
	}
	//Initializing variables, with default values
	myCurrency, myCapital := "", ""
	myPopulation := 0
	myArea := myFloat(0.0)

	//Goes through "each country", since it is displayed in an array
	for _, country := range chosenCountry {

		//the variables get their values assigned
		myPopulation = country.Population
		myArea = country.Area
		for _, capital := range country.Capital {
			myCapital = capital
			break
		}
		for currencyName := range country.Currency {
			myCurrency = currencyName
			break
		}

	}

	//Returns the values
	return myPopulation, myCapital, myCurrency, myArea, nil
}

/*
This function will retrieve the capital, and then return coordinates to capital,
Will use Geocoding API to fetch coordinates
*/
func retrieveCoordinates(capital string, w http.ResponseWriter, r *http.Request) (myFloat, myFloat, error) {

	//Creates struct that contains coordinates
	var myCoordinates struct {
		Result []Coordinates `json:"results"`
	}
	//Fetching data from Geocoding API, with count 1, to retrieve first city with this name
	err := fetchURLdata(utils.GEOCODING_API+capital+"&count=1", true, w, &myCoordinates)
	if err != nil {
		return 0, 0, err
	}
	//Initializes longitude and latitude values
	myLongitude := myFloat(0.0)
	myLatitude := myFloat(0.0)

	//Sets values to the coordinates
	for _, r := range myCoordinates.Result {
		myLongitude = r.Longitude
		myLatitude = r.Latitude
	}

	//Returns coordinates
	return myLongitude, myLatitude, nil
}

/*
Function retrieves coordinates (to a certain capital),
then returns temperature and precipitation
*/
func retrieveWeather(longitude myFloat, latitude myFloat, w http.ResponseWriter, r *http.Request) (myFloat, myFloat, error) {

	long := strconv.FormatFloat(float64(longitude), 'f', 2, 32)
	lat := strconv.FormatFloat(float64(latitude), 'f', 2, 32)

	//Struct that contains the temperature (an array of hourly measurements of temperature in one day)
	// and for precipitation in one day
	var myWeather struct {
		Hourly struct {
			Temperature   []myFloat `json:"temperature_2m"`
			Precipitation []myFloat `json:"precipitation"`
		} `json:"hourly"`
	}

	//Fetching data from the forecast API
	err := fetchURLdata(utils.FORECAST_API+"latitude="+lat+"&longitude="+long+"&hourly=temperature_2m,precipitation&forecast_days=1", true, w, &myWeather)
	if err != nil {
		return 0, 0, err
	}
	//Initializing sum of all temperatures, and add them together
	sumTemp := myFloat(0.0)
	for _, value := range myWeather.Hourly.Temperature {
		sumTemp += value
	}
	//finds average temperature using sumTemp and number of measurements
	avgTemp := sumTemp / myFloat(len(myWeather.Hourly.Temperature))

	//Initializing sum of all precipitation, and add them together
	sumPrecipitation := myFloat(0.0)
	for _, value := range myWeather.Hourly.Precipitation {
		sumPrecipitation += value
	}
	//finds average precipitiation using sumPrecipitation and number of measurements
	avgPrecipitation := sumPrecipitation / myFloat(len(myWeather.Hourly.Precipitation))

	//Returns data
	return avgTemp, avgPrecipitation, nil
}

func retrieveCurrencyExchangeRates(currency string, w http.ResponseWriter, r *http.Request) (map[string]myFloat, error) {
	currencyData := make(map[string]myFloat)

	var Currencies struct {
		Currency map[string]myFloat `json:"rates"`
	}

	err := fetchURLdata(utils.CURRENCY_API+currency, true, w, &Currencies)
	if err != nil {
		return nil, err
	}
	for currencyName, currencyValue := range Currencies.Currency {
		currencyData[currencyName] = currencyValue
	}

	return currencyData, nil
}

func whatTimeNow() string {
	currentTime := time.Now()
	timeLayout := "20060102 15:04" //YYYYMMDD HH:mm

	formattedTime := currentTime.Format(timeLayout)
	return formattedTime
}

func floatFormat(number myFloat) (myFloat, error) {
	stringFloat := strconv.FormatFloat(float64(number), 'f', 2, 64)
	newFloat, err := strconv.ParseFloat(stringFloat, 64)
	if err != nil {
		return 0, err
	}
	return myFloat(newFloat), nil
}
