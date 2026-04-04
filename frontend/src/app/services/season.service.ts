import { Injectable, inject } from '@angular/core';
import { BehaviorSubject } from 'rxjs';
import { ScheduleService } from './schedule.service';
import { SeasonSummary } from '../models';

@Injectable({ providedIn: 'root' })
export class SeasonService {
  private scheduleService = inject(ScheduleService);

  readonly seasons$    = new BehaviorSubject<SeasonSummary[]>([]);
  readonly selectedId$ = new BehaviorSubject<string>('');

  /** Reload the season list. Optionally force-select a specific season by id. */
  load(selectId?: string): void {
    this.scheduleService.listSeasons().subscribe({
      next: (list) => {
        this.seasons$.next(list);
        if (selectId) {
          this.selectedId$.next(selectId);
        } else if (!this.selectedId$.value) {
          const active = list.find(s => s.active);
          this.selectedId$.next(active ? active.id : list[0]?.id ?? '');
        }
      },
      error: () => {},
    });
  }

  select(id: string): void {
    this.selectedId$.next(id);
  }
}
