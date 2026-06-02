package application

import (
	"context"
	"errors"

	"github.com/CraveCompass/api/internal/domain"
)

type SubmitVoteInput struct {
	SessionID    string          `json:"session_id"`
	UserID       string          `json:"user_id"`
	RestaurantID string          `json:"restaurant_id"`
	Vote         domain.VoteType `json:"vote"`
}

type SubmitVoteUseCase struct {
	sessionRepo SessionRepository
}

func NewSubmitVoteUseCase(sr SessionRepository) *SubmitVoteUseCase {
	return &SubmitVoteUseCase{
		sessionRepo: sr,
	}
}

func (uc *SubmitVoteUseCase) Execute(ctx context.Context, input SubmitVoteInput) (*domain.Session, bool, error) {
	session, err := uc.sessionRepo.GetByID(ctx, input.SessionID)
	if err != nil {
		return nil, false, err
	}
	if session == nil {
		return nil, false, errors.New("session not found")
	}

	isMatch, err := session.RecordVote(input.UserID, input.RestaurantID, input.Vote)
	if err != nil {
		return nil, false, err
	}

	if err := uc.sessionRepo.Save(ctx, session); err != nil {
		return nil, false, err
	}

	return session, isMatch, nil
}
