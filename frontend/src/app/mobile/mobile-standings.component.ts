import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { distinctUntilChanged } from 'rxjs';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { ScoreService } from '../services/score.service';
import { SeasonService } from '../services/season.service';
import { PlayerStats } from '../models';

@Component({
  selector: 'app-mobile-standings',
  standalone: true,
  imports: [CommonModule, MatProgressSpinnerModule],
  styles: [`
    .header {
      padding: 16px 16px 8px;
      font-size: 18px;
      font-weight: 500;
      color: #212121;
    }

    .class-section { margin-bottom: 4px; }

    .class-label {
      font-size: 11px;
      font-weight: 600;
      letter-spacing: .6px;
      text-transform: uppercase;
      color: #757575;
      padding: 12px 16px 4px;
    }

    .standings-header, .standing-row {
      display: grid;
      grid-template-columns: 28px 36px 1fr 36px 36px 56px;
      align-items: center;
      padding: 0 16px;
      gap: 4px;
    }

    .standings-header {
      font-size: 11px;
      color: #9e9e9e;
      font-weight: 600;
      padding-top: 4px;
      padding-bottom: 4px;
      border-bottom: 1px solid rgba(0,0,0,.1);
    }

    .standing-row {
      min-height: 44px;
      border-bottom: 1px solid rgba(0,0,0,.05);
      background: #fff;

      &:nth-child(even) { background: #f9f9f9; }
    }

    .rank   { font-size: 12px; color: #9e9e9e; font-weight: 600; }
    .nr     { font-size: 12px; color: #757575; }
    .name   { font-size: 14px; font-weight: 500; color: #212121; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .wins   { font-size: 13px; font-weight: 700; color: #2e7d32; text-align: center; }
    .losses { font-size: 13px; color: #c62828; text-align: center; }
    .legs   { font-size: 12px; color: #616161; text-align: right; }

    .spinner-wrap { display: flex; justify-content: center; padding: 40px; }

    .empty {
      text-align: center;
      color: #9e9e9e;
      padding: 24px 16px;
      font-size: 13px;
    }
  `],
  template: `
    <div *ngIf="loading" class="spinner-wrap">
      <mat-spinner diameter="40" />
    </div>

    <ng-container *ngIf="!loading">
      <div class="header">Klassement</div>

      <div *ngFor="let cls of classes" class="class-section">
        <div class="class-label">{{ cls.label }}</div>

        <div class="standings-header">
          <span>#</span>
          <span>Nr</span>
          <span>Naam</span>
          <span style="text-align:center">W</span>
          <span style="text-align:center">V</span>
          <span style="text-align:right">Legs</span>
        </div>

        <div *ngFor="let s of cls.stats; let i = index" class="standing-row">
          <span class="rank">{{ i + 1 }}</span>
          <span class="nr">{{ s.player.nr }}</span>
          <span class="name">{{ s.player.name }}</span>
          <span class="wins">{{ s.wins }}</span>
          <span class="losses">{{ s.losses }}</span>
          <span class="legs">{{ s.pointsFor }}/{{ s.pointsAgainst }}</span>
        </div>

        <div *ngIf="cls.stats.length === 0" class="empty">
          Geen gespeelde wedstrijden.
        </div>
      </div>

      <div *ngIf="classes.length === 0" class="empty">
        Nog geen gespeelde wedstrijden.
      </div>
    </ng-container>
  `,
})
export class MobileStandingsComponent implements OnInit {
  private scoreService  = inject(ScoreService);
  private seasonService = inject(SeasonService);
  private destroyRef    = inject(DestroyRef);

  classes: { label: string; stats: PlayerStats[] }[] = [];
  loading = false;

  ngOnInit(): void {
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      distinctUntilChanged(),
    ).subscribe(() => this.load());
  }

  private load(): void {
    const sid = this.seasonService.selectedId$.value || undefined;
    this.loading = true;
    this.scoreService.getStats(sid).subscribe({
      next: (stats) => {
        this.classes = this.buildClasses(stats);
        this.loading = false;
      },
      error: () => { this.loading = false; },
    });
  }

  private buildClasses(allStats: PlayerStats[]): { label: string; stats: PlayerStats[] }[] {
    const classValues = [...new Set(allStats.map(s => s.player.class || ''))].sort();
    if (classValues.every(c => c === '')) {
      return [{ label: 'Alle spelers', stats: this.sortedStats(allStats) }];
    }
    const result = classValues
      .filter(c => c !== '')
      .map(c => {
        const filtered = allStats.filter(s => (s.player.class || '') === c);
        return { label: `Klasse ${c}`, stats: this.sortedStats(filtered) };
      });
    const noClass = allStats.filter(s => !s.player.class);
    if (noClass.length > 0) {
      result.push({ label: 'Overig', stats: this.sortedStats(noClass) });
    }
    return result;
  }

  private sortedStats(stats: PlayerStats[]): PlayerStats[] {
    return [...stats].sort((a, b) => {
      if (b.wins !== a.wins) return b.wins - a.wins;
      return (b.pointsFor - b.pointsAgainst) - (a.pointsFor - a.pointsAgainst);
    });
  }
}
