// Package usecase contains the business logic of DartScheduler, organised as use cases.
// Use cases orchestrate domain operations and are independent of the infrastructure layer;
// they communicate through the repository interfaces from the domain package.
package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

type PlayerUseCase struct {
	repo     domain.PlayerRepository
	matches  domain.MatchRepository
	listRepo domain.PlayerListRepository
}

func NewPlayerUseCase(repo domain.PlayerRepository, matches domain.MatchRepository, listRepo domain.PlayerListRepository) *PlayerUseCase {
	return &PlayerUseCase{repo: repo, matches: matches, listRepo: listRepo}
}

// ImportPlayers upserts players by member number, preserving existing UUIDs so that
// match references remain intact. Buddy preferences from the "samen" column replace
// all existing buddy settings.
func (uc *PlayerUseCase) ImportPlayers(ctx context.Context, inputs []PlayerInput, buddies []BuddyPairInput, listName string) error {
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

	var listID *uuid.UUID
	if listName != "" {
		list, found, err := uc.listRepo.FindByName(ctx, listName)
		if err != nil {
			return fmt.Errorf("find player list: %w", err)
		}
		if !found {
			list = domain.PlayerList{ID: uuid.New(), Name: listName, CreatedAt: time.Now()}
			if err := uc.listRepo.Save(ctx, list); err != nil {
				return fmt.Errorf("create player list: %w", err)
			}
			log.Printf("[ImportPlayers] ledenlijst aangemaakt naam=%q id=%s", listName, list.ID)
		} else {
			log.Printf("[ImportPlayers] ledenlijst gevonden naam=%q id=%s", listName, list.ID)
		}
		id := list.ID
		listID = &id
	}
	for i := range players {
		players[i].ListID = listID
	}

	if err := uc.repo.SaveBatch(ctx, players); err != nil {
		return err
	}
	log.Printf("[ImportPlayers] upserted %d players", len(players))

	// Resolve BuddyNr → player IDs using the just-saved (or pre-existing) players.
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

	// Collect buddy pairs from the "samen" (buddy) column.
	var fromExcel []BuddyPairInput
	for _, in := range inputs {
		if in.BuddyNr == "" || in.Nr == "" {
			continue
		}
		playerID, ok := nrToID[in.Nr]
		if !ok {
			continue
		}
		buddyID, ok := nrToID[in.BuddyNr]
		if !ok {
			log.Printf("[ImportPlayers] buddy nr %q not found for player %q", in.BuddyNr, in.Nr)
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
// Names are converted from the stored "Achternaam, Voornaam" format to "Voornaam Achternaam".
func (uc *PlayerUseCase) ListPlayers(ctx context.Context) ([]domain.Player, error) {
	log.Printf("[ListPlayers]")
	players, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	for i := range players {
		players[i].Name = domain.FormatDisplayName(players[i].Name)
	}
	log.Printf("[ListPlayers] found %d players", len(players))
	return players, nil
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

// DeletePlayer removes a player along with all their matches and buddy preferences.
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

// ListPlayerLists returns all named player lists.
func (uc *PlayerUseCase) ListPlayerLists(ctx context.Context) ([]PlayerListSummary, error) {
	lists, err := uc.listRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]PlayerListSummary, len(lists))
	for i, l := range lists {
		result[i] = PlayerListSummary{ID: l.ID.String(), Name: l.Name, CreatedAt: l.CreatedAt}
	}
	return result, nil
}
