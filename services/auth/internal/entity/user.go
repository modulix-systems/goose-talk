package entity

import "time"

type (
	User struct {
		ID         int
		Username   string
		Email      string
		FirstName  string
		LastName   string
		PhotoUrl   string
		Friends    []User
		CreatedAt  time.Time `db:"created_at" json:"created_at"`
		UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
		LastSeenAt time.Time
		IsActive   bool
		BirthDate  time.Time
		AboutMe    string
	}
)
