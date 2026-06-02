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
	ID           string                         `json:"id"`
	HostID       string                         `json:"host_id"`
	Status       SessionStatus                  `json:"status"`
	RadiusMeters int                            `json:"radius_meters"`
	Participants []Participant                  `json:"participants"`
	Pool         []Restaurant                   `json:"pool"`
	MatchedID    string                         `json:"matched_id,omitempty"`
	Votes        map[string]map[string]VoteType `json:"votes"`
	CreatedAt    time.Time                      `json:"created_at"`
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

func (s *Session) RecordVote(userID, restaurantID string, vote VoteType) (bool, error) {
	if s.Status != StatusActive {
		return false, errors.New("cannot vote in an inactive session")
	}

	if s.Votes == nil {
		s.Votes = make(map[string]map[string]VoteType)
	}
	if s.Votes[restaurantID] == nil {
		s.Votes[restaurantID] = make(map[string]VoteType)
	}

	s.Votes[restaurantID][userID] = vote

	return s.CheckConsensus(restaurantID), nil
}

func (s *Session) CheckConsensus(restaurantID string) bool {
	likesCount := 0
	for _, participant := range s.Participants {
		vote, exists := s.Votes[restaurantID][participant.ID]
		if !exists || vote == VoteDislike {
			return false
		}
		if vote == VoteLike || vote == VoteSuperLike {
			likesCount++
		}
	}

	if likesCount == len(s.Participants) && likesCount > 0 {
		s.Status = StatusMatched
		s.MatchedID = restaurantID
		return true
	}

	return false
}
