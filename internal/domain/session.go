package domain

import (
	"errors"
	"time"
)

type SessionStatus string

const (
	StatusActive         SessionStatus = "ACTIVE"
	StatusMatched        SessionStatus = "MATCHED"
	StatusFinished       SessionStatus = "FINISHED"
	StatusHostTieBreaker SessionStatus = "HOST_TIE_BREAKER"
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

type SessionFilters struct {
	PriceTiers []int    `json:"price_tiers"`
	MinRating  float64  `json:"min_rating"`
	Cuisines   []string `json:"cuisines"`
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
	TiedIDs      []string                       `json:"tied_ids,omitempty"`
	Filters      SessionFilters                 `json:"filters"`
	CreatedAt    time.Time                      `json:"created_at"`
}

type SessionUpdatedEvent struct {
	Event   string  `json:"event"`
	Session Session `json:"session"`
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

	if s.CheckConsensus(restaurantID) {
		return true, nil
	}

	if s.CheckCompletionFallback() {
		return true, nil
	}

	return false, nil
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

func (s *Session) CheckCompletionFallback() bool {
	totalExpectedVotes := len(s.Participants) * len(s.Pool)
	totalCastVotes := 0

	for _, userVotes := range s.Votes {
		totalCastVotes += len(userVotes)
	}

	if totalCastVotes < totalExpectedVotes || totalExpectedVotes == 0 {
		return false
	}

	bestScore := -9999
	var tiedIDs []string

	for restID, userVotes := range s.Votes {
		score := 0
		for _, vote := range userVotes {
			if vote == VoteLike {
				score += 1
			} else if vote == VoteSuperLike {
				score += 2
			} else if vote == VoteDislike {
				score -= 1
			}
		}

		if score > bestScore {
			bestScore = score
			tiedIDs = []string{restID}
		} else if score == bestScore {
			tiedIDs = append(tiedIDs, restID)
		}
	}

	if len(tiedIDs) == 1 {
		s.Status = StatusMatched
		s.MatchedID = tiedIDs[0]
		return true
	} else if len(tiedIDs) > 1 {
		s.Status = StatusHostTieBreaker
		s.TiedIDs = tiedIDs
		return true
	}

	return false
}
