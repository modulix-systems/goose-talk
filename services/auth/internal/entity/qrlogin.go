package entity

// Entity for QR code login flow. Inspired from telegram (https://core.telegram.org/api/qr-login)
type QRCodeLoginToken struct {
	Value string `json:"value"`
	// ClientId is unique identifier for client which requested token
	ClientId string

	// Login metadata
	IpAddr     string `json:"-"`
	DeviceInfo string `json:"-"`
}
