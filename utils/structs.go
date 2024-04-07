package utils

import "time"

// Struct for registration
type Registration struct {
	Country  string   `json:"country"`
	Isocode  string   `json:"isocode,omitempty"`
	Features Features `json:"features"`
}

type Features struct {
	Temperature      bool     `json:"temperature, omitempty"` // Note: In degrees Celsius
	Precipitation    bool     `json:"precipitation, omitempty"`
	Capital          bool     `json:"capital, omitempty"`
	Coordinates      bool     `json:"coordinates, omitempty"`
	Population       bool     `json:"population, omitempty"`
	Area             bool     `json:"area, omitempty"`
	TargetCurrencies []string `json:"targetcurrencies, omitempty"`
	LastChange       []Change `json:"lastchange, omitempty"`
}

type Change struct {
	ID         int       `json:"id"`
	LastChange time.Time `json:"lastchange"`
}

// Struct for country API
type CountryInfo struct {
	Name struct {
		Common string `json:"common"`
	} `json:"name"`
	Isocode string `json:"cca2"`
}
