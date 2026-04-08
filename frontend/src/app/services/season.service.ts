import { Injectable, inject } from '@angular/core';
import { BehaviorSubject, combineLatest, map } from 'rxjs';
import { ScheduleService } from './schedule.service';
import { PlayerService } from './player.service';
import { SeasonSummary, PlayerList } from '../models';

@Injectable({ providedIn: 'root' })
export class SeasonService {
  private scheduleService = inject(ScheduleService);
  private playerService = inject(PlayerService);

  readonly seasons$ = new BehaviorSubject<SeasonSummary[]>([]);
  readonly selectedId$ = new BehaviorSubject<string>('');
  readonly playerLists$ = new BehaviorSubject<PlayerList[]>([]);

  /**
   * The effective player list ID for the currently selected season.
   * Uses the season's own playerListId if set; otherwise falls back to the
   * list named "Oud" (created by the migration for legacy players).
   * Null when no lists are available.
   */
  readonly effectivePlayerListId$ = combineLatest([
    this.selectedId$,
    this.seasons$,
    this.playerLists$,
  ]).pipe(
    map(([id, seasons, lists]) => {
      const season = seasons.find((s) => s.id === id);
      if (season?.playerListId) return season.playerListId;
      const oud = lists.find((l) => l.name === 'Oud');
      return oud?.id ?? (lists[0]?.id ?? null);
    }),
  );

  /** Reload the season list. Optionally force-select a specific season by id. */
  load(selectId?: string): void {
    this.scheduleService.listSeasons().subscribe({
      next: (list) => {
        this.seasons$.next(list);
        if (selectId) {
          this.selectedId$.next(selectId);
        } else if (!this.selectedId$.value) {
          const active = list.find((s) => s.active);
          this.selectedId$.next(active ? active.id : (list[0]?.id ?? ''));
        }
      },
      error: () => {},
    });
    this.playerService.getPlayerLists().subscribe({
      next: (lists) => this.playerLists$.next(lists),
      error: () => {},
    });
  }

  select(id: string): void {
    this.selectedId$.next(id);
  }
}
