package geoip

type GetLocationResponse struct {
	City    string `json:"city"`
	Country string `json:"country"`
}