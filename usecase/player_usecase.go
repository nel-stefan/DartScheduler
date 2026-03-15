// Package usecase bevat de bedrijfslogica van DartScheduler georganiseerd als use cases.
// Use cases orkestreren domeinoperaties en zijn onafhankelijk van de infrastructuurlaag;
// ze communiceren via de repository-interfaces uit het domain-pakket.
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

// ImportPlayers vervangt alle bestaande spelers door de opgegeven lijst en slaat
// buddy-voorkeuren op. Een lege lijst geeft ErrInvalidInput terug.
func (uc *PlayerUseCase) ImportPlayers(ctx context.Context, inputs []PlayerInput, buddies []BuddyPairInput) error {
	if len(inputs) == 0 {
		return fmt.Errorf("%w: no players provided", domain.ErrInvalidInput)
	}

	players := make([]domain.Player, len(inputs))
	for i, in := range inputs {
		players[i] = domain.Player{
			ID:          uuid.New(),
			Nr:          in.Nr,
			Name:        in.Name,
			Email:       in.Email,
			Sponsor:     in.Sponsor,
			Address:     in.Address,
			PostalCode:  in.PostalCode,
			City:        in.City,
			Phone:       in.Phone,
			Mobile:      in.Mobile,
			MemberSince: in.MemberSince,
			Class:       in.Class,
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

// UpdatePlayer updates the mutable fields of an existing player.
func (uc *PlayerUseCase) UpdatePlayer(ctx context.Context, p domain.Player) error {
	return uc.repo.Save(ctx, p)
}

// GetBuddies returns the buddy player IDs for the given player.
func (uc *PlayerUseCase) GetBuddies(ctx context.Context, playerID domain.PlayerID) ([]domain.PlayerID, error) {
	return uc.repo.FindBuddiesForPlayer(ctx, playerID)
}

// SetBuddies replaces all buddy preferences for the given player.
func (uc *PlayerUseCase) SetBuddies(ctx context.Context, playerID domain.PlayerID, buddyIDs []domain.PlayerID) error {
	if err := uc.repo.DeleteBuddiesForPlayer(ctx, playerID); err != nil {
		return err
	}
	for _, bid := range buddyIDs {
		if err := uc.repo.SaveBuddyPreference(ctx, domain.BuddyPreference{
			PlayerID: playerID,
			BuddyID:  bid,
		}); err != nil {
			return err
		}
	}
	return nil
}
