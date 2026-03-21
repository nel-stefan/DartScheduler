import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { filter, distinctUntilChanged } from 'rxjs';
import { Evening, Match, Player } from '../models';
import { ScheduleService } from '../services/schedule.service';
import { PlayerService } from '../services/player.service';
import { SeasonService } from '../services/season.service';

function displayName(name: string): string {
  const idx = name.indexOf(', ');
  return idx >= 0 ? `${name.slice(idx + 2)} ${name.slice(0, idx)}` : name;
}

@Component({
  selector: 'app-mobile-avond',
  standalone: true,
  imports: [CommonModule],
  styles: [`
    :host { display: block; }

    .header {
      background: #4e342e; color: #fff; padding: 12px 16px 0;
      position: sticky; top: 0; z-index: 10;
    }
    .header-title { font-size: 15px; font-weight: 500; margin: 0 0 8px; }

    .chips {
      display: flex; gap: 6px; overflow-x: auto; padding-bottom: 10px;
      scrollbar-width: none;
    }
    .chips::-webkit-scrollbar { display: none; }

    .chip {
      flex-shrink: 0; padding: 4px 12px; border-radius: 16px; font-size: 12px;
      cursor: pointer; border: 1.5px solid rgba(255,255,255,.35);
      color: rgba(255,255,255,.65); background: transparent; white-space: nowrap;
    }
    .chip.active {
      background: rgba(255,255,255,.18); border-color: #fff; color: #fff; font-weight: 600;
    }
    .chip.inhaal { border-color: rgba(206,147,216,.5); color: rgba(206,147,216,.85); }
    .chip.inhaal.active {
      background: rgba(206,147,216,.15); border-color: #ce93d8; color: #ce93d8;
    }

    .body { padding: 12px; display: flex; flex-direction: column; gap: 8px; }

    .card {
      background: #fff; border-radius: 10px; padding: 12px 14px;
      box-shadow: 0 1px 3px rgba(0,0,0,.1);
    }

    .players {
      display: grid; grid-template-columns: 1fr auto 1fr;
      align-items: center; gap: 8px;
    }
    .player-a { font-size: 13px; font-weight: 600; }
    .player-b { font-size: 13px; font-weight: 600; text-align: right; }
    .score-badge {
      font-size: 17px; font-weight: 700; color: #4e342e;
      white-space: nowrap; text-align: center;
    }

    .card-footer {
      display: flex; justify-content: space-between; align-items: center;
      margin-top: 8px;
    }
    .status {
      font-size: 11px; padding: 2px 8px; border-radius: 10px; font-weight: 500;
    }
    .status.played { background: #e8f5e9; color: #2e7d32; }
    .status.open   { background: #fff8e1; color: #e65100; }

    .btn-score {
      font-size: 12px; padding: 5px 14px; border-radius: 16px;
      border: none; cursor: pointer; font-weight: 500;
      background: #4e342e; color: #fff;
    }
    .btn-score.edit { background: #6d4c41; }

    .empty { padding: 40px 16px; text-align: center; color: #9e9e9e; font-size: 14px; }
    .loader { padding: 40px 16px; text-align: center; color: #9e9e9e; }
  `],
  template: `
    <div class="header">
      <p class="header-title">Avond</p>
      <div class="chips">
        <button
          *ngFor="let ev of evenings"
          class="chip"
          [class.active]="ev.id === selectedEveningId"
          [class.inhaal]="ev.isInhaalAvond"
          (click)="selectEvening(ev.id)">
          {{ ev.isInhaalAvond ? 'Inhaal' : 'Avond ' + ev.number }}
        </button>
      </div>
    </div>

    <div class="loader" *ngIf="loading">Laden…</div>

    <div class="body" *ngIf="!loading && selectedEvening">
      <div class="empty" *ngIf="selectedEvening.matches.length === 0">
        Geen wedstrijden gevonden.
      </div>
      <div class="card" *ngFor="let m of selectedEvening.matches">
        <div class="players">
          <div class="player-a">{{ playerName(m.playerA) }}</div>
          <div class="score-badge">{{ matchScore(m) }}</div>
          <div class="player-b">{{ playerName(m.playerB) }}</div>
        </div>
        <div class="card-footer">
          <span class="status" [class.played]="m.played" [class.open]="!m.played">
            {{ m.played ? 'Gespeeld' : 'Open' }}
          </span>
          <button class="btn-score" [class.edit]="m.played" (click)="enterScore(m)">
            {{ m.played ? 'Wijzigen' : 'Score invoeren' }}
          </button>
        </div>
      </div>
    </div>
  `,
})
export class MobileAvondComponent implements OnInit {
  private scheduleService = inject(ScheduleService);
  private playerService   = inject(PlayerService);
  private seasonService   = inject(SeasonService);
  private router          = inject(Router);
  private destroyRef      = inject(DestroyRef);

  evenings:          Evening[] = [];
  selectedEveningId = '';
  players:           Player[]  = [];
  loading           = true;

  get selectedEvening(): Evening | undefined {
    return this.evenings.find(e => e.id === this.selectedEveningId);
  }

  ngOnInit(): void {
    this.playerService.list().subscribe(p => (this.players = p));
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      filter(id => !!id),
      distinctUntilChanged(),
    ).subscribe(id => this.loadSchedule(id));
  }

  private loadSchedule(id: string): void {
    this.loading = true;
    this.scheduleService.getById(id).subscribe(schedule => {
      this.evenings = schedule.evenings;
      this.autoSelect();
      this.loading = false;
    });
  }

  private autoSelect(): void {
    if (this.evenings.length === 0) return;
    const today    = new Date().toISOString().split('T')[0];
    const upcoming = this.evenings.find(e => e.date >= today);
    this.selectedEveningId = upcoming?.id ?? this.evenings[this.evenings.length - 1].id;
  }

  selectEvening(id: string): void {
    this.selectedEveningId = id;
  }

  playerName(id: string): string {
    const p = this.players.find(pl => pl.id === id);
    return p ? displayName(p.name) : '…';
  }

  matchScore(m: Match): string {
    return m.played ? `${m.scoreA} – ${m.scoreB}` : 'vs';
  }

  enterScore(m: Match): void {
    this.router.navigate(['/m/score', m.id], {
      state: { match: m, eveningId: this.selectedEveningId, players: this.players },
    });
  }
}
