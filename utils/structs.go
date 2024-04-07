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
