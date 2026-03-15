import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { distinctUntilChanged, filter } from 'rxjs';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { ScheduleService } from '../../services/schedule.service';
import { SeasonService } from '../../services/season.service';
import { PlayerInfoItem, EveningInfoItem, ScheduleInfo, BuddyPairItem } from '../../models';

interface PlayerRow {
  player: PlayerInfoItem;
  cells: CellData[];  // one per evening, in order
  totalMatches: number;
  eveningCount: number;  // how many evenings with at least 1 match
}

interface CellData {
  count: number;
  consecutive: boolean;  // this evening is part of a consecutive streak
}

@Component({
  selector: 'app-info',
  standalone: true,
  imports: [CommonModule, MatCardModule, MatTableModule, MatIconModule, MatChipsModule],
  styles: [`
    .page { padding: 24px; }
    h2 { margin: 0 0 20px 0; }
    h3 { margin: 20px 0 12px 0; font-size: 16px; font-weight: 500; color: #424242; }

    /* Matrix grid */
    .matrix-wrapper { overflow-x: auto; }
    .matrix-table { border-collapse: collapse; font-size: 12px; white-space: nowrap; }
    .matrix-table th, .matrix-table td { border: 1px solid #e0e0e0; padding: 4px 6px; }
    .matrix-table th { background: #f5f5f5; font-weight: 500; text-align: center; }
    .th-player { text-align: left !important; min-width: 140px; position: sticky; left: 0; background: #f5f5f5; z-index: 1; }
    .td-player { position: sticky; left: 0; background: white; z-index: 1; font-size: 12px; }
    .td-player strong { font-size: 11px; color: #616161; margin-right: 4px; }
    .cell-empty { background: #fafafa; color: #bdbdbd; text-align: center; }
    .cell-played { background: #e8f5e9; color: #2e7d32; text-align: center; font-weight: 600; }
    .cell-consec { background: #fff3e0; color: #e65100; text-align: center; font-weight: 600; }
    .th-total { background: #ede7f6; min-width: 50px; }
    .td-total { text-align: center; font-weight: 600; color: #4527a0; background: #ede7f6; }

    /* Per-player summary table */
    table { width: 100%; }

    /* Buddy pairs */
    .buddy-pair-row { display: flex; align-items: center; gap: 12px; padding: 8px 0; border-bottom: 1px solid #f0f0f0; }
    .buddy-names { min-width: 240px; font-size: 13px; }
    .buddy-name-a { font-weight: 600; }
    .buddy-chip-row { display: flex; flex-wrap: wrap; gap: 6px; }
    .evening-chip { display: inline-block; background: #e3f2fd; color: #0277bd; border-radius: 12px;
                    padding: 2px 10px; font-size: 12px; font-weight: 500; }
    .no-shared { color: #9e9e9e; font-size: 13px; font-style: italic; }

    .legend { display: flex; gap: 16px; font-size: 12px; color: #616161; margin-bottom: 12px; align-items: center; }
    .legend-item { display: flex; align-items: center; gap: 4px; }
    .legend-box { width: 16px; height: 16px; border-radius: 3px; border: 1px solid #ccc; }
  `],
  template: `
    <div class="page">
      <h2>Seizoen Info</h2>

      <ng-container *ngIf="!info">
        <p style="color:#9e9e9e">Selecteer een seizoen om info te zien.</p>
      </ng-container>

      <ng-container *ngIf="info">

        <!-- Legend -->
        <div class="legend">
          <span>Legenda:</span>
          <span class="legend-item">
            <span class="legend-box" style="background:#e8f5e9"></span> Speelt op deze avond
          </span>
          <span class="legend-item">
            <span class="legend-box" style="background:#fff3e0"></span> Opeenvolgende avonden
          </span>
        </div>

        <!-- Matrix -->
        <mat-card style="margin-bottom:24px">
          <mat-card-header><mat-card-title>Verdeling per avond</mat-card-title></mat-card-header>
          <mat-card-content>
            <div class="matrix-wrapper">
              <table class="matrix-table">
                <thead>
                  <tr>
                    <th class="th-player">Speler</th>
                    <th *ngFor="let ev of info.evenings" style="text-align:center">
                      Av.{{ ev.number }}<br>
                      <span style="font-weight:400;font-size:10px;color:#757575">{{ ev.date | date:'d/M' }}</span>
                    </th>
                    <th class="th-total">Totaal</th>
                  </tr>
                </thead>
                <tbody>
                  <tr *ngFor="let row of playerRows">
                    <td class="td-player">
                      <strong>{{ row.player.nr }}</strong>{{ row.player.name }}
                    </td>
                    <td *ngFor="let cell of row.cells"
                        [class.cell-empty]="cell.count === 0"
                        [class.cell-played]="cell.count > 0 && !cell.consecutive"
                        [class.cell-consec]="cell.count > 0 && cell.consecutive">
                      {{ cell.count || '' }}
                    </td>
                    <td class="td-total">{{ row.totalMatches }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Per-player summary -->
        <mat-card style="margin-bottom:24px">
          <mat-card-header><mat-card-title>Overzicht per speler</mat-card-title></mat-card-header>
          <mat-card-content>
            <table mat-table [dataSource]="playerRows" style="width:100%">

              <ng-container matColumnDef="nr">
                <th mat-header-cell *matHeaderCellDef style="width:48px">Nr</th>
                <td mat-cell *matCellDef="let row">{{ row.player.nr }}</td>
              </ng-container>

              <ng-container matColumnDef="name">
                <th mat-header-cell *matHeaderCellDef>Naam</th>
                <td mat-cell *matCellDef="let row"><strong>{{ row.player.name }}</strong></td>
              </ng-container>

              <ng-container matColumnDef="eveningCount">
                <th mat-header-cell *matHeaderCellDef style="width:90px;text-align:center">Avonden</th>
                <td mat-cell *matCellDef="let row" style="text-align:center">{{ row.eveningCount }}</td>
              </ng-container>

              <ng-container matColumnDef="totalMatches">
                <th mat-header-cell *matHeaderCellDef style="width:100px;text-align:center">Wedstrijden</th>
                <td mat-cell *matCellDef="let row" style="text-align:center;font-weight:600">{{ row.totalMatches }}</td>
              </ng-container>

              <ng-container matColumnDef="consecutive">
                <th mat-header-cell *matHeaderCellDef style="width:160px">Opeenvolgende avonden</th>
                <td mat-cell *matCellDef="let row">
                  <span *ngFor="let streak of getStreaks(row)" style="margin-right:6px;font-size:12px;color:#e65100">
                    Av.{{ streak.join(', ') }}
                  </span>
                  <span *ngIf="getStreaks(row).length === 0" style="color:#bdbdbd;font-size:12px">—</span>
                </td>
              </ng-container>

              <tr mat-header-row *matHeaderRowDef="summaryCols"></tr>
              <tr mat-row *matRowDef="let row; columns: summaryCols;"></tr>
            </table>
          </mat-card-content>
        </mat-card>

        <!-- Buddy pairs -->
        <mat-card *ngIf="info.buddyPairs.length > 0">
          <mat-card-header><mat-card-title>Koppels & gedeelde avonden</mat-card-title></mat-card-header>
          <mat-card-content>
            <div class="buddy-pair-row" *ngFor="let pair of info.buddyPairs">
              <div class="buddy-names">
                <span class="buddy-name-a">{{ pair.playerANr }} {{ pair.playerAName }}</span>
                <span style="color:#9e9e9e;margin:0 6px">↔</span>
                <span>{{ pair.playerBNr }} {{ pair.playerBName }}</span>
              </div>
              <div class="buddy-chip-row" *ngIf="pair.eveningNrs.length > 0">
                <span class="evening-chip" *ngFor="let nr of pair.eveningNrs">Av.{{ nr }}</span>
              </div>
              <span class="no-shared" *ngIf="pair.eveningNrs.length === 0">Geen gedeelde avonden</span>
            </div>
          </mat-card-content>
        </mat-card>

        <p *ngIf="info.buddyPairs.length === 0" style="color:#9e9e9e;font-size:13px">
          Geen koppels geconfigureerd voor dit seizoen.
        </p>

      </ng-container>
    </div>
  `,
})
export class InfoComponent implements OnInit {
  private scheduleService = inject(ScheduleService);
  private seasonService   = inject(SeasonService);
  private destroyRef      = inject(DestroyRef);

  info: ScheduleInfo | null = null;
  playerRows: PlayerRow[] = [];

  summaryCols = ['nr', 'name', 'eveningCount', 'totalMatches', 'consecutive'];

  ngOnInit(): void {
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      distinctUntilChanged(),
      filter(id => !!id),
    ).subscribe(id => this.load(id));
  }

  private load(scheduleId: string): void {
    this.scheduleService.getInfo(scheduleId).subscribe({
      next: (info) => {
        this.info = info;
        this.playerRows = this.buildPlayerRows(info);
      },
      error: () => {
        this.info = null;
        this.playerRows = [];
      },
    });
  }

  private buildPlayerRows(info: ScheduleInfo): PlayerRow[] {
    // Build lookup: playerId → { eveningId → count }
    const lookup = new Map<string, Map<string, number>>();
    for (const cell of info.matrix) {
      if (!lookup.has(cell.playerId)) lookup.set(cell.playerId, new Map());
      lookup.get(cell.playerId)!.set(cell.eveningId, cell.count);
    }

    // Sort evenings by number
    const evenings = [...info.evenings].sort((a, b) => a.number - b.number);

    return info.players.map(player => {
      const byEvening = lookup.get(player.id) ?? new Map<string, number>();

      // Build count array per evening
      const counts = evenings.map(ev => byEvening.get(ev.id) ?? 0);

      // Detect consecutive evenings: mark index i as consecutive if counts[i-1] > 0 or counts[i+1] > 0
      const consecutive = counts.map((c, i) => {
        if (c === 0) return false;
        const prevHas = i > 0 && counts[i - 1] > 0;
        const nextHas = i < counts.length - 1 && counts[i + 1] > 0;
        return prevHas || nextHas;
      });

      const cells: CellData[] = counts.map((count, i) => ({ count, consecutive: consecutive[i] }));
      const totalMatches = counts.reduce((s, c) => s + c, 0);
      const eveningCount = counts.filter(c => c > 0).length;

      return { player, cells, totalMatches, eveningCount };
    }).filter(row => row.totalMatches > 0)
      .sort((a, b) => {
        const nrA = parseInt(a.player.nr) || 9999;
        const nrB = parseInt(b.player.nr) || 9999;
        return nrA - nrB;
      });
  }

  getStreaks(row: PlayerRow): number[][] {
    const evenings = this.info ? [...this.info.evenings].sort((a, b) => a.number - b.number) : [];
    // Find groups of consecutive evenings where the player plays
    const activeNrs: number[] = row.cells
      .map((cell, i) => cell.count > 0 ? evenings[i]?.number : null)
      .filter((n): n is number => n !== null);

    const streaks: number[][] = [];
    let current: number[] = [];
    for (let i = 0; i < activeNrs.length; i++) {
      if (current.length === 0 || activeNrs[i] === activeNrs[i - 1] + 1) {
        // But we need to check against sorted evening numbers, not activeNrs indices
        // activeNrs already contains the actual evening numbers in order
        current.push(activeNrs[i]);
      } else {
        if (current.length >= 2) streaks.push(current);
        current = [activeNrs[i]];
      }
    }
    if (current.length >= 2) streaks.push(current);
    return streaks;
  }
}
