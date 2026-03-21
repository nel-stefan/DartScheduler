import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { distinctUntilChanged } from 'rxjs';
import { CommonModule } from '@angular/common';
import { ScoreService } from '../services/score.service';
import { SeasonService } from '../services/season.service';
import { PlayerStats } from '../models';

@Component({
  selector: 'app-mobile-stand',
  standalone: true,
  imports: [CommonModule],
  styles: [`
    :host { display: block; }

    .header {
      background: #4e342e; color: #fff; padding: 12px 16px;
      position: sticky; top: 0; z-index: 10;
    }
    .header h2 { margin: 0; font-size: 16px; font-weight: 500; }

    .body { padding: 12px; display: flex; flex-direction: column; gap: 16px; }

    .class-card {
      background: #fff; border-radius: 10px; overflow: hidden;
      box-shadow: 0 1px 3px rgba(0,0,0,.1);
    }
    .class-title {
      background: #5d4037; color: #fff;
      font-size: 13px; font-weight: 600; padding: 8px 14px;
    }

    .row {
      display: grid;
      grid-template-columns: 28px 36px 1fr 32px 32px 44px 44px;
      align-items: center; gap: 0;
      padding: 8px 14px; font-size: 13px;
      border-bottom: 1px solid #f5f0ee;
    }
    .row:last-child { border-bottom: none; }
    .row.head {
      font-size: 11px; font-weight: 600; color: #757575;
      text-transform: uppercase; padding: 6px 14px;
      background: #faf7f5; border-bottom: 1px solid #ece7e4;
    }

    .col-rank  { color: #9e9e9e; font-size: 12px; }
    .col-nr    { color: #9e9e9e; font-size: 12px; }
    .col-name  { font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .col-num   { text-align: center; }
    .col-win   { text-align: center; font-weight: 600; color: #2e7d32; }
    .col-lose  { text-align: center; color: #c62828; }
    .col-pts   { text-align: center; font-size: 12px; }

    .loader { padding: 40px; text-align: center; color: #9e9e9e; }
    .empty  { padding: 20px 14px; text-align: center; color: #9e9e9e; font-size: 13px; }
  `],
  template: `
    <div class="header">
      <h2>Stand</h2>
    </div>

    <div class="loader" *ngIf="loading">Laden…</div>

    <div class="body" *ngIf="!loading">
      <div class="class-card" *ngFor="let cls of classes">
        <div class="class-title">{{ cls.label }}</div>

        <div class="row head">
          <span class="col-rank">#</span>
          <span class="col-nr">Nr</span>
          <span class="col-name">Naam</span>
          <span class="col-num col-win">W</span>
          <span class="col-num col-lose">V</span>
          <span class="col-pts">+pnt</span>
          <span class="col-pts">-pnt</span>
        </div>

        <div class="row" *ngFor="let s of cls.stats; let i = index">
          <span class="col-rank">{{ i + 1 }}</span>
          <span class="col-nr">{{ s.player.nr }}</span>
          <span class="col-name">{{ s.player.name }}</span>
          <span class="col-num col-win">{{ s.wins }}</span>
          <span class="col-num col-lose">{{ s.losses }}</span>
          <span class="col-pts">{{ s.pointsFor }}</span>
          <span class="col-pts">{{ s.pointsAgainst }}</span>
        </div>

        <div class="empty" *ngIf="cls.stats.length === 0">
          Nog geen gespeelde wedstrijden.
        </div>
      </div>
    </div>
  `,
})
export class MobileStandComponent implements OnInit {
  private scoreService  = inject(ScoreService);
  private seasonService = inject(SeasonService);
  private destroyRef    = inject(DestroyRef);

  classes: { label: string; stats: PlayerStats[] }[] = [];
  loading = true;

  ngOnInit(): void {
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      distinctUntilChanged(),
    ).subscribe(() => this.loadStats());
  }

  private loadStats(): void {
    this.loading = true;
    const sid = this.seasonService.selectedId$.value || undefined;
    this.scoreService.getStats(sid).subscribe(stats => {
      this.classes = this.buildClasses(stats);
      this.loading = false;
    });
  }

  private buildClasses(allStats: PlayerStats[]): { label: string; stats: PlayerStats[] }[] {
    const classValues = [...new Set(allStats.map(s => s.player.class || ''))].sort();
    if (classValues.every(c => c === '')) {
      return [{ label: 'Alle spelers', stats: this.sortedStats(allStats) }];
    }
    const result = classValues
      .filter(c => c !== '')
      .map(c => ({
        label: `Klasse ${c}`,
        stats: this.sortedStats(allStats.filter(s => (s.player.class || '') === c)),
      }));
    const noClass = allStats.filter(s => !s.player.class);
    if (noClass.length > 0) result.push({ label: 'Overig', stats: this.sortedStats(noClass) });
    return result;
  }

  private sortedStats(stats: PlayerStats[]): PlayerStats[] {
    return [...stats].sort((a, b) => {
      if (b.wins !== a.wins) return b.wins - a.wins;
      return (b.pointsFor - b.pointsAgainst) - (a.pointsFor - a.pointsAgainst);
    });
  }
}
