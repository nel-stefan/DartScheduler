import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { distinctUntilChanged } from 'rxjs';
import { CommonModule } from '@angular/common';
import { ScoreService } from '../services/score.service';
import { SeasonService } from '../services/season.service';
import { PlayerStats, DutyStats } from '../models';

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

    .section-tabs {
      display: flex; background: #3e2723;
    }
    .tab {
      flex: 1; padding: 10px 0; text-align: center;
      font-size: 13px; color: rgba(255,255,255,.6);
      cursor: pointer; border-bottom: 2px solid transparent;
    }
    .tab.active { color: #fff; border-bottom-color: #fff; }

    .body { padding: 12px; display: flex; flex-direction: column; gap: 16px; }

    .class-card {
      background: #fff; border-radius: 10px; overflow: hidden;
      box-shadow: 0 1px 3px rgba(0,0,0,.1);
    }
    .class-title {
      background: #5d4037; color: #fff;
      font-size: 13px; font-weight: 600; padding: 8px 14px;
    }

    /* standings: rank nr name W V +pnt -pnt 180 HF */
    .row {
      display: grid;
      grid-template-columns: 20px 28px 1fr 26px 26px 34px 34px 32px 40px;
      align-items: center;
      padding: 7px 10px; font-size: 12px;
      border-bottom: 1px solid #f5f0ee;
    }
    .row:last-child { border-bottom: none; }
    .row.head {
      font-size: 10px; font-weight: 600; color: #757575; text-transform: uppercase;
      background: #faf7f5; border-bottom: 1px solid #ece7e4; padding: 5px 10px;
    }
    .c-rank  { color: #9e9e9e; font-size: 11px; }
    .c-nr    { color: #9e9e9e; font-size: 11px; }
    .c-name  { font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; padding-right: 4px; }
    .c-num   { text-align: center; }
    .c-win   { text-align: center; font-weight: 600; color: #2e7d32; }
    .c-lose  { text-align: center; color: #c62828; }
    .c-pts   { text-align: center; }
    .c-180   { text-align: center; font-weight: 600; color: #7b1fa2; }
    .c-hf    { text-align: center; font-weight: 600; color: #e65100; }

    /* duty: rank nr name keer */
    .duty-row {
      display: grid;
      grid-template-columns: 20px 28px 1fr 40px;
      align-items: center;
      padding: 7px 10px; font-size: 12px;
      border-bottom: 1px solid #f5f0ee;
    }
    .duty-row:last-child { border-bottom: none; }
    .duty-row.head {
      font-size: 10px; font-weight: 600; color: #757575; text-transform: uppercase;
      background: #faf7f5; border-bottom: 1px solid #ece7e4; padding: 5px 10px;
    }

    .loader { padding: 40px; text-align: center; color: #9e9e9e; }
    .empty  { padding: 20px 14px; text-align: center; color: #9e9e9e; font-size: 13px; }
  `],
  template: `
    <div class="header">
      <h2>Stand</h2>
    </div>

    <div class="section-tabs">
      <div class="tab" [class.active]="tab === 'stand'" (click)="tab = 'stand'">Klassement</div>
      <div class="tab" [class.active]="tab === 'duty'"  (click)="tab = 'duty'">Schrijver / Teller</div>
    </div>

    <div class="loader" *ngIf="loading">Laden…</div>

    <!-- Klassement -->
    <div class="body" *ngIf="!loading && tab === 'stand'">
      <div class="class-card" *ngFor="let cls of classes">
        <div class="class-title">{{ cls.label }}</div>

        <div class="row head">
          <span class="c-rank">#</span>
          <span class="c-nr">Nr</span>
          <span class="c-name">Naam</span>
          <span class="c-num c-win">W</span>
          <span class="c-num c-lose">V</span>
          <span class="c-pts">+pnt</span>
          <span class="c-pts">-pnt</span>
          <span class="c-180">180</span>
          <span class="c-hf">HF</span>
        </div>

        <div class="row" *ngFor="let s of cls.stats; let i = index">
          <span class="c-rank">{{ i + 1 }}</span>
          <span class="c-nr">{{ s.player.nr }}</span>
          <span class="c-name">{{ s.player.name }}</span>
          <span class="c-num c-win">{{ s.wins }}</span>
          <span class="c-num c-lose">{{ s.losses }}</span>
          <span class="c-pts">{{ s.pointsFor }}</span>
          <span class="c-pts">{{ s.pointsAgainst }}</span>
          <span class="c-180">{{ s.oneEighties || '—' }}</span>
          <span class="c-hf">{{ s.highestFinish || '—' }}</span>
        </div>

        <div class="empty" *ngIf="cls.stats.length === 0">Nog geen gespeelde wedstrijden.</div>
      </div>
    </div>

    <!-- Schrijver / Teller -->
    <div class="body" *ngIf="!loading && tab === 'duty'">
      <div class="class-card">
        <div class="class-title">Schrijver / Teller</div>

        <div class="duty-row head">
          <span class="c-rank">#</span>
          <span class="c-nr">Nr</span>
          <span class="c-name">Naam</span>
          <span class="c-num">Keer</span>
        </div>

        <div class="duty-row" *ngFor="let s of dutyStats; let i = index">
          <span class="c-rank">{{ i + 1 }}</span>
          <span class="c-nr">{{ s.player.nr }}</span>
          <span class="c-name">{{ s.player.name }}</span>
          <span class="c-num" style="font-weight:600">{{ s.count }}</span>
        </div>

        <div class="empty" *ngIf="dutyStats.length === 0">Nog geen schrijvers of tellers geregistreerd.</div>
      </div>
    </div>
  `,
})
export class MobileStandComponent implements OnInit {
  private scoreService  = inject(ScoreService);
  private seasonService = inject(SeasonService);
  private destroyRef    = inject(DestroyRef);

  classes:   { label: string; stats: PlayerStats[] }[] = [];
  dutyStats: DutyStats[] = [];
  loading = true;
  tab: 'stand' | 'duty' = 'stand';

  ngOnInit(): void {
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      distinctUntilChanged(),
    ).subscribe(() => this.loadStats());
  }

  private loadStats(): void {
    this.loading = true;
    const sid = this.seasonService.selectedId$.value || undefined;
    let done = 0;
    const check = () => { if (++done === 2) this.loading = false; };

    this.scoreService.getStats(sid).subscribe(stats => {
      this.classes = this.buildClasses(stats);
      check();
    });
    this.scoreService.getDutyStats(sid).subscribe(d => {
      this.dutyStats = d.sort((a, b) => b.count - a.count);
      check();
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
