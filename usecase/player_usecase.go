// Package usecase bevat de bedrijfslogica van DartScheduler georganiseerd als use cases.
// Use cases orkestreren domeinoperaties en zijn onafhankelijk van de infrastructuurlaag;
// ze communiceren via de repository-interfaces uit het domain-pakket.
package usecase

import (
	"context"
	"fmt"
	"log"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

type PlayerUseCase struct {
	repo    domain.PlayerRepository
	matches domain.MatchRepository
}

func NewPlayerUseCase(repo domain.PlayerRepository, matches domain.MatchRepository) *PlayerUseCase {
	return &PlayerUseCase{repo: repo, matches: matches}
}

// ImportPlayers voegt spelers toe of werkt ze bij (upsert op lidnummer).
// Bestaande spelers behouden hun UUID zodat wedstrijdreferenties intact blijven.
// Buddy-voorkeuren uit de "samen"-kolom vervangen alle bestaande buddy-instellingen.
func (uc *PlayerUseCase) ImportPlayers(ctx context.Context, inputs []PlayerInput, buddies []BuddyPairInput) error {
	log.Printf("[ImportPlayers] count=%d", len(inputs))
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

	if err := uc.repo.SaveBatch(ctx, players); err != nil {
		return err
	}
	log.Printf("[ImportPlayers] upserted %d players", len(players))

	// Resolve SamenNr → player IDs using the just-saved players.
	// Re-fetch so we have the actual (possibly pre-existing) UUIDs.
	allPlayers, err := uc.repo.FindAll(ctx)
	if err != nil {
		return err
	}
	nrToID := make(map[string]domain.PlayerID, len(allPlayers))
	for _, p := range allPlayers {
		if p.Nr != "" {
			nrToID[p.Nr] = p.ID
		}
	}

	// Collect buddy pairs from the "samen" column.
	var fromExcel []BuddyPairInput
	for _, in := range inputs {
		if in.SamenNr == "" || in.Nr == "" {
			continue
		}
		playerID, ok := nrToID[in.Nr]
		if !ok {
			continue
		}
		buddyID, ok := nrToID[in.SamenNr]
		if !ok {
			log.Printf("[ImportPlayers] samen nr %q not found for player %q", in.SamenNr, in.Nr)
			continue
		}
		fromExcel = append(fromExcel, BuddyPairInput{PlayerID: playerID, BuddyID: buddyID})
	}

	// Merge: Excel buddies take precedence; manual buddies (passed in) are also kept.
	allBuddies := append(fromExcel, buddies...)

	if len(allBuddies) > 0 {
		if err := uc.repo.DeleteAllBuddyPairs(ctx); err != nil {
			return err
		}
		for _, bp := range allBuddies {
			if err := uc.repo.SaveBuddyPreference(ctx, domain.BuddyPreference{
				PlayerID: bp.PlayerID,
				BuddyID:  bp.BuddyID,
			}); err != nil {
				return err
			}
		}
		log.Printf("[ImportPlayers] saved %d buddy pairs (%d from excel)", len(allBuddies), len(fromExcel))
	}
	return nil
}

// ListPlayers returns all players regardless of season.
func (uc *PlayerUseCase) ListPlayers(ctx context.Context) ([]domain.Player, error) {
	log.Printf("[ListPlayers]")
	players, err := uc.repo.FindAll(ctx)
	log.Printf("[ListPlayers] found %d players", len(players))
	return players, err
}

// UpdatePlayer updates the mutable fields of an existing player.
func (uc *PlayerUseCase) UpdatePlayer(ctx context.Context, p domain.Player) error {
	log.Printf("[UpdatePlayer] id=%s name=%q class=%q", p.ID, p.Name, p.Class)
	return uc.repo.Save(ctx, p)
}

// GetBuddies returns the buddy player IDs for the given player.
func (uc *PlayerUseCase) GetBuddies(ctx context.Context, playerID domain.PlayerID) ([]domain.PlayerID, error) {
	return uc.repo.FindBuddiesForPlayer(ctx, playerID)
}

// DeletePlayer verwijdert een speler inclusief alle wedstrijden en buddy-voorkeuren.
func (uc *PlayerUseCase) DeletePlayer(ctx context.Context, id domain.PlayerID) error {
	log.Printf("[DeletePlayer] id=%s", id)
	if err := uc.matches.DeleteByPlayer(ctx, id); err != nil {
		return err
	}
	if err := uc.repo.DeleteBuddiesForPlayer(ctx, id); err != nil {
		return err
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	log.Printf("[DeletePlayer] done id=%s", id)
	return nil
}

// SetBuddies replaces all buddy preferences for the given player.
func (uc *PlayerUseCase) SetBuddies(ctx context.Context, playerID domain.PlayerID, buddyIDs []domain.PlayerID) error {
	log.Printf("[SetBuddies] playerID=%s count=%d", playerID, len(buddyIDs))
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
