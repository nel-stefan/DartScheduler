import { Component, inject, OnInit } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { forkJoin } from 'rxjs';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { ScheduleService } from '../services/schedule.service';
import { PlayerService } from '../services/player.service';
import { Evening, Match, Player } from '../models';

@Component({
  selector: 'app-mobile-matches',
  standalone: true,
  imports: [CommonModule, DatePipe, MatButtonModule, MatIconModule, MatProgressSpinnerModule],
  styles: [`
    .page-header {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 8px 8px 4px;
      background: #fff;
      border-bottom: 1px solid rgba(0,0,0,.1);
      position: sticky;
      top: 0;
      z-index: 10;
    }

    .header-text { flex: 1; }

    .header-title {
      font-size: 16px;
      font-weight: 500;
      color: #212121;
    }

    .header-date {
      font-size: 12px;
      color: #757575;
    }

    .match-card {
      margin: 8px 12px;
      background: #fff;
      border-radius: 8px;
      padding: 12px 16px;
      box-shadow: 0 1px 3px rgba(0,0,0,.12);
    }

    .matchup {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 10px;
    }

    .player-name {
      flex: 1;
      font-size: 14px;
      font-weight: 500;
      color: #212121;
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .vs {
      font-size: 12px;
      color: #9e9e9e;
      flex-shrink: 0;
    }

    .match-footer {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    .score {
      font-size: 18px;
      font-weight: 700;
      color: #2e7d32;
    }

    .absent-label {
      font-size: 13px;
      color: #9e9e9e;
      font-style: italic;
    }

    .spinner-wrap { display: flex; justify-content: center; padding: 40px; }

    .empty {
      text-align: center;
      color: #9e9e9e;
      padding: 40px 16px;
      font-size: 14px;
    }
  `],
  template: `
    <div *ngIf="loading" class="spinner-wrap">
      <mat-spinner diameter="40" />
    </div>

    <ng-container *ngIf="!loading">
      <div class="page-header">
        <button mat-icon-button (click)="back()">
          <mat-icon>arrow_back</mat-icon>
        </button>
        <div class="header-text">
          <div class="header-title">
            {{ evening?.isInhaalAvond ? 'Inhaalavond' : 'Avond ' + evening?.number }}
          </div>
          <div class="header-date">{{ evening?.date | date:'EEEE d MMMM yyyy' }}</div>
        </div>
      </div>

      <div *ngIf="(evening?.matches?.length ?? 0) === 0" class="empty">
        Geen wedstrijden op deze avond.
      </div>

      <div *ngFor="let m of evening?.matches" class="match-card">
        <div class="matchup">
          <span class="player-name">{{ playerName(m.playerA) }}</span>
          <span class="vs">vs</span>
          <span class="player-name">{{ playerName(m.playerB) }}</span>
        </div>
        <div class="match-footer">
          <span class="score" *ngIf="m.played">{{ m.scoreA }} – {{ m.scoreB }}</span>
          <span class="absent-label" *ngIf="!m.played && m.reportedBy">Afgemeld</span>
          <span *ngIf="!m.played && !m.reportedBy"></span>
          <button mat-raised-button color="primary" *ngIf="!m.played && !m.reportedBy"
            (click)="enterScore(m)">
            Score invoeren
          </button>
          <button mat-stroked-button *ngIf="m.played"
            (click)="enterScore(m)">
            Wijzigen
          </button>
        </div>
      </div>
    </ng-container>
  `,
})
export class MobileMatchesComponent implements OnInit {
  private route           = inject(ActivatedRoute);
  private router          = inject(Router);
  private scheduleService = inject(ScheduleService);
  private playerService   = inject(PlayerService);

  evening: Evening | null = null;
  private playerMap = new Map<string, Player>();
  loading = false;

  ngOnInit(): void {
    const eveningId = this.route.snapshot.paramMap.get('id')!;
    this.loading = true;
    forkJoin({
      evening: this.scheduleService.getEvening(eveningId),
      players: this.playerService.list(),
    }).subscribe({
      next: ({ evening, players }) => {
        this.evening = evening;
        for (const p of players) this.playerMap.set(p.id, p);
        this.loading = false;
      },
      error: () => { this.loading = false; },
    });
  }

  playerName(id: string): string {
    const p = this.playerMap.get(id);
    if (!p) return '—';
    const parts = p.name.split(', ');
    const first = parts.length === 2 ? parts[1].split(' ')[0] : p.name;
    return `${first} - ${p.nr}`;
  }

  enterScore(m: Match): void {
    const players = [...this.playerMap.values()];
    this.router.navigate(['/m/match', m.id], {
      state: { match: m, eveningId: this.evening!.id, players },
    });
  }

  back(): void {
    this.router.navigate(['/m/schema']);
  }
}
