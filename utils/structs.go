package utils

import "time"

// Status Struct for status
type Status struct {
	Countriesapi   int     `json:"countriesapi"`
	Meteoapi       int     `json:"meteoapi"`
	Currencyapi    int     `json:"currencyapi"`
	Notificationdb int     `json:"notificationdb"`
	Webhooks       float64 `json:"webhooks"`
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

type Firestore struct {
	Country    string   `json:"country"`
	Features   Features `json:"features"`
	IsoCode    string   `json:"isoCode"`
	LastChange time.Time `json:"lastChange"`
}

type Features struct {
    Temperature      bool     `json:"temperature"`
    Precipitation    bool     `json:"precipitation"`
    Capital          bool     `json:"capital"`
    Coordinates      bool     `json:"coordinates"`
    Population       bool     `json:"population"`
    Area             bool     `json:"area"`
    TargetCurrencies []string `json:"targetCurrencies"`
}

// Desired document structure
type Registration struct {
	ID         string   `json:"id"`
	Country    string   `json:"country"`
	IsoCode    string   `json:"isoCode"`
	Features   Features `json:"features"`
	LastChange string   `json:"lastChange"`
}