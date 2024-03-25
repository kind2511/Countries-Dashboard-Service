package utils

// Struct for registration
type Registration struct {
	Country  string   `json:"country"`
	Isocode  string   `json:"isocode"`
	Features Features `json:"features"`
}

// Struct for features
type Features struct {
	Temperature      bool     `json:"temperature"` // Note: In degrees Celsius
	Precipitation    bool     `json:"precipitation"`
	Capital          bool     `json:"capital"`
	Coordinates      bool     `json:"coordinates"`
	Population       bool     `json:"population"`
	Area             bool     `json:"area"`
	TargetCurrencies []string `json:"targetcurrencies"`
}

