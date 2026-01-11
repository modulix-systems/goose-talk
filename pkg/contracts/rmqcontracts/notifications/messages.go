package notifications

type EmailMessage struct {
	Name        string
	Data        map[string]any
	To string
}
