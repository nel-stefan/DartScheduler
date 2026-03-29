import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

import { FormsModule } from '@angular/forms';
import { filter, distinctUntilChanged } from 'rxjs';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { Evening, Player } from '../models';
import { ScheduleService } from '../services/schedule.service';
import { PlayerService } from '../services/player.service';
import { SeasonService } from '../services/season.service';
import { EveningStatService } from '../services/evening-stat.service';
import { displayName } from '../utils/display-name';

interface StatRow {
  playerId:      string;
  name:          string;
  oneEighties:   number;
  highestFinish: number;
  dirty:         boolean;
}

@Component({
  selector: 'app-mobile-stats',
  standalone: true,
  imports: [FormsModule, MatSnackBarModule],
  styles: [`
    :host { display: block; }

    .header {
      background: #4e342e; color: #fff; padding: 12px 16px 0;
      position: sticky; top: 0; z-index: 10;
    }
    .header h2 { margin: 0 0 8px; font-size: 16px; font-weight: 500; }

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
    .chip.active { background: rgba(255,255,255,.18); border-color: #fff; color: #fff; font-weight: 600; }
    .chip.inhaal { border-color: rgba(206,147,216,.5); color: rgba(206,147,216,.85); }
    .chip.inhaal.active { background: rgba(206,147,216,.15); border-color: #ce93d8; color: #ce93d8; }

    .body { padding: 12px; display: flex; flex-direction: column; gap: 8px; }

    .info {
      font-size: 12px; color: #757575; padding: 4px 2px 8px; text-align: center;
    }

    .card {
      background: #fff; border-radius: 10px; padding: 12px 14px;
      box-shadow: 0 1px 3px rgba(0,0,0,.1);
      display: grid; grid-template-columns: 1fr auto auto;
      align-items: center; gap: 8px;
    }
    .card.dirty { border-left: 3px solid #4e342e; }

    .player-name { font-size: 13px; font-weight: 500; }

    .stat-group {
      display: flex; flex-direction: column; align-items: center; gap: 2px;
    }
    .stat-label {
      font-size: 10px; text-transform: uppercase; letter-spacing: .3px; color: #9e9e9e;
    }
    .stepper {
      display: flex; align-items: center; gap: 0;
    }
    .step-btn {
      width: 28px; height: 28px; border: 1.5px solid #d7ccc8; background: #faf7f5;
      cursor: pointer; font-size: 16px; display: flex; align-items: center; justify-content: center;
      color: #4e342e; font-weight: 700;
    }
    .step-btn:first-child { border-radius: 6px 0 0 6px; }
    .step-btn:last-child  { border-radius: 0 6px 6px 0; }
    .step-val {
      width: 28px; height: 28px; border-top: 1.5px solid #d7ccc8; border-bottom: 1.5px solid #d7ccc8;
      background: #fff; display: flex; align-items: center; justify-content: center;
      font-size: 14px; font-weight: 600; color: #7b1fa2;
    }

    .finish-input {
      width: 52px; padding: 4px 6px; border: 1.5px solid #d7ccc8; border-radius: 6px;
      font-size: 14px; font-weight: 600; color: #e65100; text-align: center;
      background: #faf7f5; outline: none;
    }
    .finish-input:focus { border-color: #4e342e; }

    .save-btn {
      width: 100%; padding: 13px; border: none; border-radius: 10px;
      background: #4e342e; color: #fff; font-size: 15px; font-weight: 600;
      cursor: pointer; margin-top: 4px;
    }
    .save-btn:disabled { background: #bcaaa4; cursor: default; }

    .loader { padding: 40px; text-align: center; color: #9e9e9e; }
  `],
  template: `
    <div class="header">
      <h2>180s & Hoge Finish</h2>
      <div class="chips">
        @for (ev of evenings; track ev) {
          <button
            class="chip"
            [class.active]="ev.id === selectedEveningId"
            [class.inhaal]="ev.isInhaalAvond"
            (click)="selectEvening(ev.id)">
            {{ ev.isInhaalAvond ? 'Inhaal' : 'Avond ' + ev.number }}
          </button>
        }
      </div>
    </div>
    
    @if (loading) {
      <div class="loader">Laden…</div>
    }
    
    @if (!loading) {
      <div class="body">
        <p class="info">Voer per speler de 180s en hoogste finish in voor de geselecteerde avond.</p>
        @for (row of rows; track row) {
          <div
            class="card"
            [class.dirty]="row.dirty"
            >
            <div class="player-name">{{ row.name }}</div>
            <div class="stat-group">
              <span class="stat-label">180s</span>
              <div class="stepper">
                <button class="step-btn" type="button" (click)="dec180(row)">−</button>
                <div class="step-val">{{ row.oneEighties }}</div>
                <button class="step-btn" type="button" (click)="inc180(row)">+</button>
              </div>
            </div>
            <div class="stat-group">
              <span class="stat-label">Hoge Finish</span>
              <input
                class="finish-input"
                type="number"
                min="0"
                max="170"
                [(ngModel)]="row.highestFinish"
                (ngModelChange)="markDirty(row)"
                placeholder="0">
            </div>
          </div>
        }
        <button
          class="save-btn"
          [disabled]="saving || !hasDirty"
          (click)="saveAll()">
          {{ saving ? 'Opslaan…' : 'Opslaan' }}
        </button>
      </div>
    }
    `,
})
export class MobileStatsComponent implements OnInit {
  private scheduleService  = inject(ScheduleService);
  private playerService    = inject(PlayerService);
  private seasonService    = inject(SeasonService);
  private eveningStatSvc   = inject(EveningStatService);
  private snackBar         = inject(MatSnackBar);
  private destroyRef       = inject(DestroyRef);

  evenings:          Evening[] = [];
  selectedEveningId = '';
  players:           Player[]  = [];
  rows:              StatRow[] = [];
  loading = true;
  saving  = false;

  get hasDirty(): boolean { return this.rows.some(r => r.dirty); }

  ngOnInit(): void {
    this.playerService.list().subscribe(p => {
      this.players = p;
      this.rebuildRows();
    });
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
    });
  }

  private autoSelect(): void {
    if (this.evenings.length === 0) { this.loading = false; return; }
    const today    = new Date().toISOString().split('T')[0];
    const upcoming = this.evenings.find(e => e.date >= today);
    this.selectedEveningId = upcoming?.id ?? this.evenings[this.evenings.length - 1].id;
    this.loadEveningStats();
  }

  selectEvening(id: string): void {
    this.selectedEveningId = id;
    this.loadEveningStats();
  }

  private loadEveningStats(): void {
    if (!this.selectedEveningId) return;
    this.loading = true;
    this.eveningStatSvc.getByEvening(this.selectedEveningId).subscribe(stats => {
      this.rebuildRows(stats.reduce((m, s) => {
        m[s.playerId] = { oneEighties: s.oneEighties, highestFinish: s.highestFinish };
        return m;
      }, {} as Record<string, { oneEighties: number; highestFinish: number }>));
      this.loading = false;
    });
  }

  private rebuildRows(saved: Record<string, { oneEighties: number; highestFinish: number }> = {}): void {
    this.rows = this.players.map(p => ({
      playerId:      p.id,
      name:          displayName(p.name),
      oneEighties:   saved[p.id]?.oneEighties   ?? 0,
      highestFinish: saved[p.id]?.highestFinish ?? 0,
      dirty:         false,
    }));
  }

  inc180(row: StatRow): void { row.oneEighties++; row.dirty = true; }
  dec180(row: StatRow): void { if (row.oneEighties > 0) { row.oneEighties--; row.dirty = true; } }
  markDirty(row: StatRow):  void { row.dirty = true; }

  saveAll(): void {
    const dirty = this.rows.filter(r => r.dirty);
    if (!dirty.length) return;
    this.saving = true;

    let remaining = dirty.length;
    let hasError  = false;

    for (const row of dirty) {
      this.eveningStatSvc.upsert(
        this.selectedEveningId, row.playerId, row.oneEighties, row.highestFinish,
      ).subscribe({
        next: () => {
          row.dirty = false;
          if (--remaining === 0) this.onSaveDone(hasError);
        },
        error: () => {
          hasError = true;
          if (--remaining === 0) this.onSaveDone(hasError);
        },
      });
    }
  }

  private onSaveDone(hasError: boolean): void {
    this.saving = false;
    this.snackBar.open(hasError ? 'Fout bij opslaan' : 'Opgeslagen', '', { duration: 2000 });
  }
}
