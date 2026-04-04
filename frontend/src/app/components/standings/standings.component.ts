import { Component, inject, OnInit, DestroyRef, signal, computed } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { distinctUntilChanged } from 'rxjs';

import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatTabsModule } from '@angular/material/tabs';
import { MatIconModule } from '@angular/material/icon';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatTooltipModule } from '@angular/material/tooltip';
import { ScoreService } from '../../services/score.service';
import { SeasonService } from '../../services/season.service';
import { PlayerStats, DutyStats } from '../../models';
import { EveningStatDialogComponent, EveningStatDialogData } from '../evening-stat-dialog.component';

@Component({
  selector: 'app-standings',
  imports: [
    MatButtonModule,
    MatCardModule,
    MatDialogModule,
    MatTableModule,
    MatTabsModule,
    MatIconModule,
    MatTooltipModule,
  ],
  styles: [
    `
      .section-title {
        font-size: 18px;
        font-weight: 500;
        margin: 0 0 12px 0;
      }
      table {
        width: 100%;
      }
      .rank-col {
        width: 40px;
        font-weight: 600;
        color: #616161;
      }
      .pts-col {
        font-weight: 600;
        color: #2e7d32;
      }

      .records-row {
        display: flex;
        gap: 16px;
        margin-bottom: 20px;
        flex-wrap: wrap;
      }
      .record-card {
        display: flex;
        align-items: center;
        gap: 12px;
        background: #fff;
        border-radius: 8px;
        padding: 12px 20px;
        box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
        min-width: 220px;
      }
      .record-icon {
        font-size: 28px;
        width: 28px;
        height: 28px;
      }
      .record-label {
        font-size: 11px;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        color: #757575;
      }
      .record-name {
        font-size: 14px;
        font-weight: 500;
        color: #212121;
      }
      .record-value {
        font-size: 20px;
        font-weight: 700;
      }

      .print-only {
        display: none;
      }

      .print-table {
        width: 100%;
        border-collapse: collapse;
        font-size: 13pt;
      }
      .print-table th,
      .print-table td {
        border: 1px solid #ccc;
        padding: 4px 8px;
        text-align: left;
        line-height: 1.3;
      }
      .print-table th {
        background: #f0f0f0;
        font-weight: 600;
      }
      .print-table td.center {
        text-align: center;
      }
      .print-section-title {
        font-size: 15pt;
        font-weight: 600;
        margin: 8px 0 3px 0;
      }

      @media print {
        @page {
          margin: 12mm;
        }
        .screen-only {
          display: none !important;
        }
        .print-only {
          display: block !important;
        }
      }
    `,
  ],
  template: `
    <div style="padding:24px">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px">
        <h2 style="margin:0">Klassement</h2>
        <button mat-stroked-button (click)="print()" class="screen-only"><mat-icon>print</mat-icon> Afdrukken</button>
      </div>

      <!-- Records: minste beurten + hoogste finish (overkoepelend) -->
      @if (minTurnsRecord() || highestFinishRecord()) {
        <div class="records-row screen-only">
          @if (minTurnsRecord()) {
            <div class="record-card">
              <mat-icon class="record-icon" style="color:#0277bd">speed</mat-icon>
              <div class="record-body">
                <div class="record-label">Minste beurten</div>
                <div class="record-name">{{ minTurnsRecord()!.player.name }}</div>
                <div class="record-value" style="color:#0277bd">{{ minTurnsRecord()!.minTurns }}</div>
              </div>
            </div>
          }
          @if (highestFinishRecord()) {
            <div class="record-card">
              <mat-icon class="record-icon" style="color:#e65100">star</mat-icon>
              <div class="record-body">
                <div class="record-label">Hoogste finish</div>
                <div class="record-name">{{ highestFinishRecord()!.player.name }}</div>
                <div class="record-value" style="color:#e65100">{{ highestFinishRecord()!.highestFinish }}</div>
              </div>
            </div>
          }
        </div>
      }

      <!-- Screen: tabs -->
      <div class="screen-only">
        @if (!loading()) {
          <mat-tab-group animationDuration="150ms" color="primary" [selectedIndex]="0">
            @for (cls of sortedClasses(); track cls) {
              <mat-tab [label]="cls.label">
                <mat-card style="border-radius:0 0 8px 8px;border-top:none">
                  <mat-card-content>
                    <table mat-table [dataSource]="cls.stats" style="width:100%">
                      <ng-container matColumnDef="rank">
                        <th mat-header-cell *matHeaderCellDef class="rank-col">#</th>
                        <td mat-cell *matCellDef="let s; let i = index" class="rank-col">{{ i + 1 }}</td>
                      </ng-container>
                      <ng-container matColumnDef="nr">
                        <th mat-header-cell *matHeaderCellDef style="width:48px">Nr</th>
                        <td mat-cell *matCellDef="let s">{{ s.player.nr }}</td>
                      </ng-container>
                      <ng-container matColumnDef="name">
                        <th mat-header-cell *matHeaderCellDef style="cursor:pointer;user-select:none"
                            (click)="sortStandings('name')">
                          Naam
                          @if (standingsSortCol() === 'name') {
                            <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">
                              {{ standingsSortDir() === 'asc' ? 'arrow_upward' : 'arrow_downward' }}
                            </mat-icon>
                          }
                        </th>
                        <td mat-cell *matCellDef="let s">
                          <strong>{{ s.player.name }}</strong>
                        </td>
                      </ng-container>
                      <ng-container matColumnDef="wins">
                        <th mat-header-cell *matHeaderCellDef
                            style="width:60px;text-align:center;cursor:pointer;user-select:none"
                            (click)="sortStandings('wins')">
                          <mat-icon style="color:#2e7d32;vertical-align:middle;font-size:18px">emoji_events</mat-icon>
                          @if (standingsSortCol() === 'wins') {
                            <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">
                              {{ standingsSortDir() === 'asc' ? 'arrow_upward' : 'arrow_downward' }}
                            </mat-icon>
                          }
                        </th>
                        <td mat-cell *matCellDef="let s" style="text-align:center" class="pts-col">{{ s.wins }}</td>
                      </ng-container>
                      <ng-container matColumnDef="losses">
                        <th mat-header-cell *matHeaderCellDef
                            style="width:60px;text-align:center;cursor:pointer;user-select:none"
                            (click)="sortStandings('losses')">
                          <mat-icon style="color:#c62828;vertical-align:middle;font-size:18px">cancel</mat-icon>
                          @if (standingsSortCol() === 'losses') {
                            <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">
                              {{ standingsSortDir() === 'asc' ? 'arrow_upward' : 'arrow_downward' }}
                            </mat-icon>
                          }
                        </th>
                        <td mat-cell *matCellDef="let s" style="text-align:center;color:#c62828">{{ s.losses }}</td>
                      </ng-container>
                      <ng-container matColumnDef="pf">
                        <th mat-header-cell *matHeaderCellDef
                            style="width:60px;text-align:center;cursor:pointer;user-select:none"
                            title="Gewonnen legs" (click)="sortStandings('pf')">
                          + punten
                          @if (standingsSortCol() === 'pf') {
                            <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">
                              {{ standingsSortDir() === 'asc' ? 'arrow_upward' : 'arrow_downward' }}
                            </mat-icon>
                          }
                        </th>
                        <td mat-cell *matCellDef="let s" style="text-align:center">{{ s.pointsFor }}</td>
                      </ng-container>
                      <ng-container matColumnDef="pa">
                        <th mat-header-cell *matHeaderCellDef
                            style="width:60px;text-align:center;cursor:pointer;user-select:none"
                            title="Verloren legs" (click)="sortStandings('pa')">
                          - punten
                          @if (standingsSortCol() === 'pa') {
                            <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">
                              {{ standingsSortDir() === 'asc' ? 'arrow_upward' : 'arrow_downward' }}
                            </mat-icon>
                          }
                        </th>
                        <td mat-cell *matCellDef="let s" style="text-align:center">{{ s.pointsAgainst }}</td>
                      </ng-container>
                      <ng-container matColumnDef="180s">
                        <th mat-header-cell *matHeaderCellDef
                            style="width:52px;text-align:center;cursor:pointer;user-select:none"
                            title="Aantal 180's" (click)="sortStandings('180s')">
                          180
                          @if (standingsSortCol() === '180s') {
                            <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">
                              {{ standingsSortDir() === 'asc' ? 'arrow_upward' : 'arrow_downward' }}
                            </mat-icon>
                          }
                        </th>
                        <td mat-cell *matCellDef="let s" style="text-align:center;font-weight:600;color:#7b1fa2">
                          {{ s.oneEighties || '—' }}
                        </td>
                      </ng-container>
                      <ng-container matColumnDef="edit">
                        <th mat-header-cell *matHeaderCellDef style="width:40px"></th>
                        <td mat-cell *matCellDef="let s">
                          <button
                            mat-icon-button
                            (click)="openStatDialog(s)"
                            matTooltip="180s / HF aanpassen"
                            style="width:32px;height:32px;font-size:18px"
                            class="screen-only"
                          >
                            <mat-icon style="font-size:18px;width:18px;height:18px">edit</mat-icon>
                          </button>
                        </td>
                      </ng-container>
                      <tr mat-header-row *matHeaderRowDef="matchCols"></tr>
                      <tr mat-row *matRowDef="let row; columns: matchCols"></tr>
                    </table>
                    @if (cls.stats.length === 0) {
                      <p style="color:#9e9e9e;text-align:center;padding:24px 0">
                        Nog geen gespeelde wedstrijden in deze klasse.
                      </p>
                    }
                  </mat-card-content>
                </mat-card>
              </mat-tab>
            }

            <!-- Duty stats tab -->
            <mat-tab label="Schrijver / Teller">
              <mat-card style="border-radius:0 0 8px 8px;border-top:none">
                <mat-card-content>
                  <p style="color:#616161;font-size:13px;margin:12px 0 8px 0">
                    Totaal aantal keer als schrijver of teller ingezet (gecombineerd).
                  </p>
                  <table mat-table [dataSource]="dutyStats()" style="width:100%">
                    <ng-container matColumnDef="rank">
                      <th mat-header-cell *matHeaderCellDef class="rank-col">#</th>
                      <td mat-cell *matCellDef="let s; let i = index" class="rank-col">{{ i + 1 }}</td>
                    </ng-container>

                    <ng-container matColumnDef="nr">
                      <th mat-header-cell *matHeaderCellDef style="width:48px">Nr</th>
                      <td mat-cell *matCellDef="let s">{{ s.player.nr }}</td>
                    </ng-container>

                    <ng-container matColumnDef="name">
                      <th mat-header-cell *matHeaderCellDef>Naam</th>
                      <td mat-cell *matCellDef="let s">
                        <strong>{{ s.player.name }}</strong>
                      </td>
                    </ng-container>

                    <ng-container matColumnDef="count">
                      <th mat-header-cell *matHeaderCellDef style="width:80px;text-align:center">Keer</th>
                      <td mat-cell *matCellDef="let s" style="text-align:center;font-weight:600">{{ s.count }}</td>
                    </ng-container>

                    <tr mat-header-row *matHeaderRowDef="dutyCols"></tr>
                    <tr mat-row *matRowDef="let row; columns: dutyCols"></tr>
                  </table>

                  @if (dutyStats().length === 0) {
                    <p style="color:#9e9e9e;text-align:center;padding:24px 0">
                      Nog geen schrijvers of tellers geregistreerd.
                    </p>
                  }
                </mat-card-content>
              </mat-card>
            </mat-tab>
          </mat-tab-group>
        }
        <!-- /if !loading -->
      </div>
      <!-- /screen-only -->

      <!-- Print: flat sections -->
      <div class="print-only">
        <!-- Records -->
        @if (minTurnsRecord() || highestFinishRecord()) {
          <div style="display:flex;gap:24px;margin-bottom:12px">
            @if (minTurnsRecord()) {
              <div>
                <span style="font-size:13pt;color:#757575">Minste beurten: </span>
                <strong style="font-size:13pt">{{ minTurnsRecord()!.player.name }}</strong>
                <span style="font-size:13pt"> ({{ minTurnsRecord()!.minTurns }})</span>
              </div>
            }
            @if (highestFinishRecord()) {
              <div>
                <span style="font-size:13pt;color:#757575">Hoogste finish: </span>
                <strong style="font-size:13pt">{{ highestFinishRecord()!.player.name }}</strong>
                <span style="font-size:13pt"> ({{ highestFinishRecord()!.highestFinish }})</span>
              </div>
            }
          </div>
        }

        <!-- Page 1: all class standings -->
        @for (cls of classes(); track cls) {
          <div>
            <h3 class="print-section-title">{{ cls.label }}</h3>
            <table class="print-table">
              <thead>
                <tr>
                  <th style="width:32px">#</th>
                  <th style="width:40px">Nr</th>
                  <th>Naam</th>
                  <th class="center" style="width:60px">+ punten</th>
                  <th class="center" style="width:60px">- punten</th>
                  <th class="center" style="width:44px">180</th>
                </tr>
              </thead>
              <tbody>
                @for (s of cls.stats; track s; let i = $index) {
                  <tr>
                    <td>{{ i + 1 }}</td>
                    <td>{{ s.player.nr }}</td>
                    <td>
                      <strong>{{ s.player.name }}</strong>
                    </td>
                    <td class="center">{{ s.pointsFor }}</td>
                    <td class="center">{{ s.pointsAgainst }}</td>
                    <td class="center" style="font-weight:600">{{ s.oneEighties || '—' }}</td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        }

        <!-- Page 2: duty stats -->
        <div style="page-break-before:always"></div>
        <h3 class="print-section-title">Schrijver / Teller</h3>
        <p style="font-size:10pt;color:#616161;margin:0 0 8px 0">
          Totaal aantal keer als schrijver of teller ingezet (gecombineerd).
        </p>
        <table class="print-table">
          <thead>
            <tr>
              <th style="width:32px">#</th>
              <th style="width:40px">Nr</th>
              <th>Naam</th>
              <th class="center" style="width:56px">Keer</th>
            </tr>
          </thead>
          <tbody>
            @for (s of dutyStats(); track s; let i = $index) {
              <tr>
                <td>{{ i + 1 }}</td>
                <td>{{ s.player.nr }}</td>
                <td>
                  <strong>{{ s.player.name }}</strong>
                </td>
                <td class="center" style="font-weight:600">{{ s.count }}</td>
              </tr>
            }
          </tbody>
        </table>
      </div>
      <!-- /print-only -->
    </div>
  `,
})
export class StandingsComponent implements OnInit {
  private scoreService = inject(ScoreService);
  private seasonService = inject(SeasonService);
  private dialog = inject(MatDialog);
  private destroyRef = inject(DestroyRef);

  classes = signal<{ label: string; stats: PlayerStats[] }[]>([]);
  dutyStats = signal<DutyStats[]>([]);
  allStats = signal<PlayerStats[]>([]);
  loading = signal(true);
  minTurnsRecord = signal<PlayerStats | null>(null);
  highestFinishRecord = signal<PlayerStats | null>(null);

  matchCols = ['rank', 'nr', 'name', 'wins', 'losses', 'pf', 'pa', '180s', 'edit'];
  dutyCols = ['rank', 'nr', 'name', 'count'];

  standingsSortCol = signal<'name' | 'wins' | 'losses' | 'pf' | 'pa' | '180s'>('pf');
  standingsSortDir = signal<'asc' | 'desc'>('desc');

  sortStandings(col: 'name' | 'wins' | 'losses' | 'pf' | 'pa' | '180s'): void {
    if (this.standingsSortCol() === col) {
      this.standingsSortDir.set(this.standingsSortDir() === 'asc' ? 'desc' : 'asc');
    } else {
      this.standingsSortCol.set(col);
      this.standingsSortDir.set(col === 'name' ? 'asc' : 'desc');
    }
  }

  sortedClasses = computed(() => {
    const col = this.standingsSortCol();
    const dir = this.standingsSortDir() === 'asc' ? 1 : -1;
    return this.classes().map((cls) => ({
      label: cls.label,
      stats: [...cls.stats].sort((a, b) => {
        if (col === 'name')   return dir * a.player.name.localeCompare(b.player.name);
        if (col === 'wins')   return dir * (a.wins - b.wins);
        if (col === 'losses') return dir * (a.losses - b.losses);
        if (col === 'pf')     return dir * (a.pointsFor - b.pointsFor);
        if (col === 'pa')     return dir * (a.pointsAgainst - b.pointsAgainst);
        if (col === '180s')   return dir * ((a.oneEighties ?? 0) - (b.oneEighties ?? 0));
        return 0;
      }),
    }));
  });

  ngOnInit(): void {
    this.seasonService.selectedId$
      .pipe(takeUntilDestroyed(this.destroyRef), distinctUntilChanged())
      .subscribe(() => this.loadStats());
  }

  private loadStats(): void {
    const sid = this.seasonService.selectedId$.value || undefined;
    // Hide the tab group while data is in flight so that MatTabGroup is only
    // instantiated once the full tab list (Klasse tabs + static tab) is ready.
    // This avoids the race where the static "Schrijver / Teller" tab temporarily
    // occupies index 0 and then triggers selectedIndexChange when Klasse tabs are
    // prepended, causing the two-way binding to overwrite our desired index.
    this.loading.set(true);
    this.scoreService.getStats(sid).subscribe((s) => {
      this.allStats.set(s);
      this.classes.set(this.buildClasses(s));
      const withMinTurns = s.filter((x) => x.minTurns > 0);
      this.minTurnsRecord.set(
        withMinTurns.length ? withMinTurns.reduce((b, x) => (x.minTurns < b.minTurns ? x : b)) : null
      );
      const withHF = s.filter((x) => x.highestFinish > 0);
      this.highestFinishRecord.set(
        withHF.length ? withHF.reduce((b, x) => (x.highestFinish > b.highestFinish ? x : b)) : null
      );
      this.loading.set(false);
    });
    this.scoreService.getDutyStats(sid).subscribe((d) => {
      this.dutyStats.set(d.sort((a, b) => b.count - a.count));
    });
  }

  openStatDialog(stat: PlayerStats): void {
    const sid = this.seasonService.selectedId$.value;
    if (!sid) return;
    this.dialog
      .open(EveningStatDialogComponent, {
        data: {
          scheduleId: sid,
          players: [{ id: stat.player.id, name: stat.player.name }],
          preselectedPlayerId: stat.player.id,
        } as EveningStatDialogData,
      })
      .afterClosed()
      .subscribe((saved) => {
        if (saved) this.loadStats();
      });
  }

  private buildClasses(allStats: PlayerStats[]): { label: string; stats: PlayerStats[] }[] {
    const byClass = new Map<string, PlayerStats[]>();
    for (const stat of allStats) {
      const cls = stat.player.class || '';
      if (!byClass.has(cls)) byClass.set(cls, []);
      byClass.get(cls)!.push(stat);
    }
    const noClass = byClass.get('') ?? [];
    byClass.delete('');
    if (byClass.size === 0) {
      return [{ label: 'Alle spelers', stats: this.sortedStats(allStats) }];
    }
    const result = [...byClass.keys()].sort().map((c) => ({
      label: `Klasse ${c}`,
      stats: this.sortedStats(byClass.get(c)!),
    }));
    if (noClass.length > 0) {
      result.push({ label: 'Overig', stats: this.sortedStats(noClass) });
    }
    return result;
  }

  private sortedStats(stats: PlayerStats[]): PlayerStats[] {
    return [...stats].sort((a, b) => {
      if (b.pointsFor !== a.pointsFor) return b.pointsFor - a.pointsFor;
      const diffA = a.pointsFor - a.pointsAgainst;
      const diffB = b.pointsFor - b.pointsAgainst;
      return diffB - diffA;
    });
  }

  print(): void {
    window.print();
  }
}
