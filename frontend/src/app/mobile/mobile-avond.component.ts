import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { filter, distinctUntilChanged } from 'rxjs';
import { Evening, Match, Player } from '../models';
import { ScheduleService } from '../services/schedule.service';
import { PlayerService } from '../services/player.service';
import { SeasonService } from '../services/season.service';
import { MobileStateService } from './mobile-state.service';
import { displayName } from '../utils/display-name';

@Component({
  selector: 'app-mobile-avond',
  imports: [CommonModule, FormsModule],
  styles: [
    `
      :host {
        display: block;
      }

      .header {
        background: #4e342e;
        color: #fff;
        padding: 12px 16px;
        position: sticky;
        top: 0;
        z-index: 10;
      }
      .header-title {
        font-size: 15px;
        font-weight: 500;
        margin: 0 0 8px;
      }

      .evening-select {
        width: 100%;
        padding: 6px 10px;
        border-radius: 8px;
        font-size: 14px;
        background: rgba(255, 255, 255, 0.15);
        color: #fff;
        border: 1.5px solid rgba(255, 255, 255, 0.35);
        appearance: none;
        cursor: pointer;
      }
      .evening-select option {
        background: #4e342e;
        color: #fff;
      }

      .body {
        padding: 12px;
        display: flex;
        flex-direction: column;
        gap: 8px;
      }

      .card {
        background: #fff;
        border-radius: 10px;
        padding: 12px 14px;
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
      }

      .players {
        display: grid;
        grid-template-columns: 1fr auto 1fr;
        align-items: center;
        gap: 8px;
      }
      .player-a {
        font-size: 13px;
        font-weight: 600;
      }
      .player-b {
        font-size: 13px;
        font-weight: 600;
        text-align: right;
      }
      .score-badge {
        font-size: 17px;
        font-weight: 700;
        color: #4e342e;
        white-space: nowrap;
        text-align: center;
      }

      .card-footer {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-top: 8px;
      }
      .status {
        font-size: 11px;
        padding: 2px 8px;
        border-radius: 10px;
        font-weight: 500;
      }
      .status.played {
        background: #e8f5e9;
        color: #2e7d32;
      }
      .status.open {
        background: #fff8e1;
        color: #e65100;
      }

      .btn-score {
        font-size: 12px;
        padding: 5px 14px;
        border-radius: 16px;
        border: none;
        cursor: pointer;
        font-weight: 500;
        background: #4e342e;
        color: #fff;
      }
      .btn-score.edit {
        background: #6d4c41;
      }

      .empty {
        padding: 40px 16px;
        text-align: center;
        color: #9e9e9e;
        font-size: 14px;
      }
      .loader {
        padding: 40px 16px;
        text-align: center;
        color: #9e9e9e;
      }
    `,
  ],
  template: `
    <div class="header">
      <p class="header-title">Avond</p>
      <select class="evening-select" [(ngModel)]="selectedEveningId" (ngModelChange)="onSelectChange($event)">
        @for (ev of evenings; track ev) {
          <option [value]="ev.id">
            {{ ev.isInhaalAvond ? 'Inhaalavond' : 'Avond ' + ev.number }} — {{ ev.date | date: 'd MMM' }}
          </option>
        }
      </select>
    </div>

    @if (loading) {
      <div class="loader">Laden…</div>
    }

    @if (!loading && selectedEvening) {
      <div class="body">
        @if (selectedEvening.matches.length === 0) {
          <div class="empty">Geen wedstrijden gevonden.</div>
        }
        @for (m of selectedEvening.matches; track m) {
          <div class="card">
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
        }
      </div>
    }
  `,
})
export class MobileAvondComponent implements OnInit {
  private scheduleService = inject(ScheduleService);
  private playerService = inject(PlayerService);
  private seasonService = inject(SeasonService);
  private mobileState = inject(MobileStateService);
  private router = inject(Router);
  private destroyRef = inject(DestroyRef);

  evenings: Evening[] = [];
  selectedEveningId = '';
  players: Player[] = [];
  loading = true;

  get selectedEvening(): Evening | undefined {
    return this.evenings.find((e) => e.id === this.selectedEveningId);
  }

  ngOnInit(): void {
    this.playerService.list().subscribe((p) => (this.players = p));
    this.seasonService.selectedId$
      .pipe(
        takeUntilDestroyed(this.destroyRef),
        filter((id) => !!id),
        distinctUntilChanged()
      )
      .subscribe((id) => this.loadSchedule(id));
  }

  private loadSchedule(id: string): void {
    this.loading = true;
    this.scheduleService.getById(id).subscribe((schedule) => {
      this.evenings = schedule.evenings;
      if (this.mobileState.selectedEveningId) {
        this.selectedEveningId = this.mobileState.selectedEveningId;
      } else {
        this.autoSelect();
      }
      this.loading = false;
    });
  }

  private autoSelect(): void {
    if (this.evenings.length === 0) return;
    const today = new Date().toISOString().split('T')[0];
    const upcoming = this.evenings.find((e) => e.date >= today);
    this.selectedEveningId = upcoming?.id ?? this.evenings[this.evenings.length - 1].id;
    this.mobileState.selectedEveningId = this.selectedEveningId;
  }

  onSelectChange(id: string): void {
    this.mobileState.selectedEveningId = id;
  }

  playerName(id: string): string {
    const p = this.players.find((pl) => pl.id === id);
    return p ? displayName(p.name) : '…';
  }

  matchScore(m: Match): string {
    return m.played ? `${m.scoreA} – ${m.scoreB}` : 'vs';
  }

  enterScore(m: Match): void {
    const isInhaalAvond = this.selectedEvening?.isInhaalAvond ?? false;
    this.router.navigate(['/m/score', m.id], {
      state: {
        match: m,
        eveningId: this.selectedEveningId,
        players: this.players,
        isInhaalAvond,
        evenings: this.evenings,
        lastCatchUpPlayedDate: this.mobileState.lastCatchUpPlayedDate,
      },
    });
  }
}
