package domain

import (
	"errors"
	"time"
)

type SessionStatus string

const (
	StatusActive   SessionStatus = "ACTIVE"
	StatusMatched  SessionStatus = "MATCHED"
	StatusFinished SessionStatus = "FINISHED"
)

type VoteType string

const (
	VoteDislike   VoteType = "DISLIKE"
	VoteLike      VoteType = "LIKE"
	VoteSuperLike VoteType = "SUPERLIKE"
)

type Participant struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Session struct {
	ID           string        `json:"id"`
	HostID       string        `json:"host_id"`
	Status       SessionStatus `json:"status"`
	RadiusMeters int           `json:"radius_meters"`
	Participants []Participant `json:"participants"`
	Pool         []Restaurant  `json:"pool"`
	MatchedID    string        `json:"matched_id,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

func (s *Session) AddParticipant(p Participant) error {
	if s.Status != StatusActive {
		return errors.New("cannot join an inactive session")
	}
	for _, existing := range s.Participants {
		if existing.ID == p.ID {
			return nil
		}
	}
	s.Participants = append(s.Participants, p)
	return nil
}
