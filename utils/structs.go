package utils

import "time"

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