package usecase

import (
	"context"
	"fmt"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

type PlayerUseCase struct {
	repo domain.PlayerRepository
}

func NewPlayerUseCase(repo domain.PlayerRepository) *PlayerUseCase {
	return &PlayerUseCase{repo: repo}
}

func (uc *PlayerUseCase) ImportPlayers(ctx context.Context, inputs []PlayerInput, buddies []BuddyPairInput) error {
	if len(inputs) == 0 {
		return fmt.Errorf("%w: no players provided", domain.ErrInvalidInput)
	}

	players := make([]domain.Player, len(inputs))
	for i, in := range inputs {
		players[i] = domain.Player{
			ID:      uuid.New(),
			Name:    in.Name,
			Email:   in.Email,
			Sponsor: in.Sponsor,
		}
	}

	if err := uc.repo.DeleteAll(ctx); err != nil {
		return err
	}
	if err := uc.repo.SaveBatch(ctx, players); err != nil {
		return err
	}
	for _, bp := range buddies {
		if err := uc.repo.SaveBuddyPreference(ctx, domain.BuddyPreference{
			PlayerID: bp.PlayerID,
			BuddyID:  bp.BuddyID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (uc *PlayerUseCase) ListPlayers(ctx context.Context) ([]domain.Player, error) {
	return uc.repo.FindAll(ctx)
}
