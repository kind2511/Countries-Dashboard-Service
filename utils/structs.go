package utils

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
