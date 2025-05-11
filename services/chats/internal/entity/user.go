package entity

import "time"

type (
	User struct {
		ID          int
		Username    string
		FirstName   string
		LastName    string
		PhoneNumber string
		PhotoUrl    string
		CreatedAt   time.Time `db:"created_at" json:"created_at"`
		UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
		LastSeenAt  time.Time
	}
)
