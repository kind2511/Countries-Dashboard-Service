package utils

import "time"

/*
// Struct for registration
type Dashboard struct {
	ID          DashboardResponse `json:"id"`
	Country     string            `json:"country"`
	Isocode     string            `json:"isocode,omitempty"`
	RegFeatures RegFeatures       `json:"features, omitempty"`
	LastChange  string            `json:"lastChange"`
}

type RegFeatures struct {
	Temperature      bool     `json:"temperature,omitempty"` // Note: In degrees Celsius
	Precipitation    bool     `json:"precipitation,omitempty"`
	Capital          bool     `json:"capital,omitempty"`
	Coordinates      bool     `json:"coordinates,omitempty"`
	Population       bool     `json:"population,omitempty"`
	Area             bool     `json:"area,omitempty"`
	TargetCurrencies []string `json:"targetcurrencies,omitempty"`
}

// Struct for dashboard registration response
type DashboardResponse struct {
	ID         string `json:"id"`
	LastChange string `json:"lastChange"`
}*/

// Struct for country API
type CountryInfo struct {
	Name struct {
		Common string `json:"common"`
	} `json:"name"`
	Isocode string `json:"cca2"`
}

/*
The dashboard that is received from dashboards,
checking if certain data should be implemented and fetched
*/
type Dashboard_Get struct {
	ID         string       `json:"id"`
	Country    string       `json:"country"`
	IsoCode    string       `json:"isoCode"`
	Features   Features_Get `json:"features"`
	LastChange time.Time    `json:"lastChange"`
}

type Features_Get struct {
	Temperature      bool     `json:"temperature,omitempty"`
	Precipitation    bool     `json:"precipitation,omitempty"`
	Capital          bool     `json:"capital,omitempty"`
	Coordinates      bool     `json:"coordinates,omitempty"`
	Population       bool     `json:"population,omitempty"`
	Area             bool     `json:"area,omitempty"`
	TargetCurrencies []string `json:"targetCurrencies,omitempty"`
}

// Status Struct for status
type Status struct {
	Countriesapi   int     `json:"countriesapi"`
	Meteoapi       int     `json:"meteoapi"`
	Currencyapi    int     `json:"currencyapi"`
	Notificationdb int     `json:"notificationdb"`
	Webhooks       int     `json:"webhooks"`
	Version        string  `json:"version"`
	Uptime         float64 `json:"uptime"`
}

type WebhookRegistration struct {
	Url     string `json:"url"`
	Country string `json:"country"`
	Event   string `json:"event"`
}

type WebhookRegistrationResponse struct {
	Id string `json:"id"`
}

type WebhookGetResponse struct {
	Id      string `json:"id"`
	Url     string `json:"url"`
	Country string `json:"country"`
	Event   string `json:"event"`
}

type WebhookInvokeMessage struct {
	Id      string `json:"id"`
	Url     string `json:"url"`
	Country string `json:"country"`
	Event   string `json:"event"`
	Time    string `json:"time"`
}

type Firestore struct {
	ID         string    `json:id`
	Country    string    `json:"country"`
	Features   Features  `json:"features"`
	IsoCode    string    `json:"isoCode"`
	LastChange time.Time `json:"lastChange"`
}

type Features struct {
	Temperature      *bool    `json:"temperature"`
	Precipitation    *bool    `json:"precipitation"`
	Capital          *bool    `json:"capital"`
	Coordinates      *bool    `json:"coordinates"`
	Population       *bool    `json:"population"`
	Area             *bool    `json:"area"`
	TargetCurrencies []string `json:"targetCurrencies"`
}

/*
// Desired document structure
type Registration struct {
	ID         string   `json:"id"`
	Country    string   `json:"country"`
	IsoCode    string   `json:"isoCode"`
	Features   Features `json:"features"`
	LastChange string   `json:"lastChange"`
}*/

// Desired output for default handler
type DefaultEndpointStruct struct {
	Url         string `json:"url"`
	Method      string `json:"method"`
	Description string `json:"description"`
}
