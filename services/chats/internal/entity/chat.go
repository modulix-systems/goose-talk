package entity

import "time"

type ParticipantRole string

const (
	Owner     ParticipantRole = "owner"
	Moderator                 = "Moderator"
)

type (
	Chat struct {
		ID           int `json:"id"`
		OwnerId      int
		Participants []User
		CreatedAt    time.Time
	}
	ChatParticipant struct {
		ID       int
		Role     ParticipantRole
		JoinedAt time.Time
	}
)
