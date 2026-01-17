package notifications

type EmailMessage struct {
	Type    string
	Subject string
	Data    map[string]any
	To      string
}
