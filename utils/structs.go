package utils

// Struct for registration
type Dashboard struct {
	ID         DashboardResponse `json:"id"`
	Country    string            `json:"country"`
	Isocode    string            `json:"isocode,omitempty"`
	Features   Features          `json:"features, omitempty"`
	LastChange string            `json:"lastChange"`
}

type Features struct {
	Temperature      bool     `json:"temperature, omitempty"` // Note: In degrees Celsius
	Precipitation    bool     `json:"precipitation, omitempty"`
	Capital          bool     `json:"capital, omitempty"`
	Coordinates      bool     `json:"coordinates, omitempty"`
	Population       bool     `json:"population, omitempty"`
	Area             bool     `json:"area, omitempty"`
	TargetCurrencies []string `json:"targetcurrencies, omitempty"`
}

// Struct for dashboard registration response
type DashboardResponse struct {
	ID         string `json:"id"`
	LastChange string `json:"lastChange"`
}

// Struct for country API
type CountryInfo struct {
	Name struct {
		Common string `json:"common"`
	} `json:"name"`
	Isocode string `json:"cca2"`
}
