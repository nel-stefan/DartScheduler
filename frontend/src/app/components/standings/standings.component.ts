import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { distinctUntilChanged } from 'rxjs';
import { CommonModule } from '@angular/common';
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
    imports: [CommonModule, MatButtonModule, MatCardModule, MatDialogModule,
        MatTableModule, MatTabsModule, MatIconModule, MatTooltipModule],
    styles: [`
    .section-title {
      font-size: 18px;
      font-weight: 500;
      margin: 0 0 12px 0;
    }
    table { width: 100%; }
    .rank-col { width: 40px; font-weight: 600; color: #616161; }
    .pts-col  { font-weight: 600; color: #2e7d32; }


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
      box-shadow: 0 1px 4px rgba(0,0,0,.12);
      min-width: 220px;
    }
    .record-icon { font-size: 28px; width: 28px; height: 28px; }
    .record-label { font-size: 11px; text-transform: uppercase; letter-spacing: .5px; color: #757575; }
    .record-name  { font-size: 14px; font-weight: 500; color: #212121; }
    .record-value { font-size: 20px; font-weight: 700; }

    .print-only { display: none; }

    .print-table {
      width: 100%;
      border-collapse: collapse;
      font-size: 9pt;
    }
    .print-table th, .print-table td {
      border: 1px solid #ccc;
      padding: 2px 6px;
      text-align: left;
      line-height: 1.3;
    }
    .print-table th { background: #f0f0f0; font-weight: 600; }
    .print-table td.center { text-align: center; }
    .print-section-title { font-size: 11pt; font-weight: 600; margin: 8px 0 3px 0; }

    @media print {
      @page { margin: 10mm; }
      .screen-only { display: none !important; }
      .print-only  { display: block !important; }

    }
  `],
    template: `
    <div style="padding:24px">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px">
        <h2 style="margin:0">Klassement</h2>
        <button mat-stroked-button (click)="print()" class="screen-only">
          <mat-icon>print</mat-icon> Afdrukken
        </button>
      </div>

      <!-- Records: minste beurten + hoogste finish (overkoepelend) -->
      <div class="records-row screen-only" *ngIf="minTurnsRecord || highestFinishRecord">
        <div class="record-card" *ngIf="minTurnsRecord">
          <mat-icon class="record-icon" style="color:#0277bd">speed</mat-icon>
          <div class="record-body">
            <div class="record-label">Minste beurten</div>
            <div class="record-name">{{ minTurnsRecord.player.name }}</div>
            <div class="record-value" style="color:#0277bd">{{ minTurnsRecord.minTurns }}</div>
          </div>
        </div>
        <div class="record-card" *ngIf="highestFinishRecord">
          <mat-icon class="record-icon" style="color:#e65100">star</mat-icon>
          <div class="record-body">
            <div class="record-label">Hoogste finish</div>
            <div class="record-name">{{ highestFinishRecord.player.name }}</div>
            <div class="record-value" style="color:#e65100">{{ highestFinishRecord.highestFinish }}</div>
          </div>
        </div>
      </div>

      <!-- Screen: tabs -->
      <div class="screen-only">
      <mat-tab-group animationDuration="150ms" color="primary" [(selectedIndex)]="selectedTabIndex">

        <mat-tab *ngFor="let cls of classes" [label]="cls.label">
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
                  <th mat-header-cell *matHeaderCellDef>Naam</th>
                  <td mat-cell *matCellDef="let s"><strong>{{ s.player.name }}</strong></td>
                </ng-container>

<ng-container matColumnDef="wins">
                  <th mat-header-cell *matHeaderCellDef style="width:60px;text-align:center">
                    <mat-icon style="color:#2e7d32;vertical-align:middle;font-size:18px">emoji_events</mat-icon>
                  </th>
                  <td mat-cell *matCellDef="let s" style="text-align:center" class="pts-col">{{ s.wins }}</td>
                </ng-container>

                <ng-container matColumnDef="losses">
                  <th mat-header-cell *matHeaderCellDef style="width:60px;text-align:center">
                    <mat-icon style="color:#c62828;vertical-align:middle;font-size:18px">cancel</mat-icon>
                  </th>
                  <td mat-cell *matCellDef="let s" style="text-align:center;color:#c62828">{{ s.losses }}</td>
                </ng-container>

                <ng-container matColumnDef="pf">
                  <th mat-header-cell *matHeaderCellDef style="width:60px;text-align:center"
                      title="Gewonnen legs">+ punten</th>
                  <td mat-cell *matCellDef="let s" style="text-align:center">{{ s.pointsFor }}</td>
                </ng-container>

                <ng-container matColumnDef="pa">
                  <th mat-header-cell *matHeaderCellDef style="width:60px;text-align:center"
                      title="Verloren legs">- punten</th>
                  <td mat-cell *matCellDef="let s" style="text-align:center">{{ s.pointsAgainst }}</td>
                </ng-container>

                <ng-container matColumnDef="180s">
                  <th mat-header-cell *matHeaderCellDef style="width:52px;text-align:center" title="Aantal 180's">180</th>
                  <td mat-cell *matCellDef="let s" style="text-align:center;font-weight:600;color:#7b1fa2">
                    {{ s.oneEighties || '—' }}
                  </td>
                </ng-container>

                <ng-container matColumnDef="edit">
                  <th mat-header-cell *matHeaderCellDef style="width:40px"></th>
                  <td mat-cell *matCellDef="let s">
                    <button mat-icon-button (click)="openStatDialog(s)" matTooltip="180s / HF aanpassen"
                            style="width:32px;height:32px;font-size:18px" class="screen-only">
                      <mat-icon style="font-size:18px;width:18px;height:18px">edit</mat-icon>
                    </button>
                  </td>
                </ng-container>

                <tr mat-header-row *matHeaderRowDef="matchCols"></tr>
                <tr mat-row *matRowDef="let row; columns: matchCols;"></tr>
              </table>

              <p *ngIf="cls.stats.length === 0" style="color:#9e9e9e;text-align:center;padding:24px 0">
                Nog geen gespeelde wedstrijden in deze klasse.
              </p>
            </mat-card-content>
          </mat-card>
        </mat-tab>

        <!-- Duty stats tab -->
        <mat-tab label="Schrijver / Teller">
          <mat-card style="border-radius:0 0 8px 8px;border-top:none">
            <mat-card-content>
              <p style="color:#616161;font-size:13px;margin:12px 0 8px 0">
                Totaal aantal keer als schrijver of teller ingezet (gecombineerd).
              </p>
              <table mat-table [dataSource]="dutyStats" style="width:100%">

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
                  <td mat-cell *matCellDef="let s"><strong>{{ s.player.name }}</strong></td>
                </ng-container>

                <ng-container matColumnDef="count">
                  <th mat-header-cell *matHeaderCellDef style="width:80px;text-align:center">Keer</th>
                  <td mat-cell *matCellDef="let s" style="text-align:center;font-weight:600">{{ s.count }}</td>
                </ng-container>

                <tr mat-header-row *matHeaderRowDef="dutyCols"></tr>
                <tr mat-row *matRowDef="let row; columns: dutyCols;"></tr>
              </table>

              <p *ngIf="dutyStats.length === 0" style="color:#9e9e9e;text-align:center;padding:24px 0">
                Nog geen schrijvers of tellers geregistreerd.
              </p>
            </mat-card-content>
          </mat-card>
        </mat-tab>

      </mat-tab-group>
      </div><!-- /screen-only -->

      <!-- Print: flat sections -->
      <div class="print-only">
        <!-- Records -->
        <div *ngIf="minTurnsRecord || highestFinishRecord" style="display:flex;gap:24px;margin-bottom:12px">
          <div *ngIf="minTurnsRecord">
            <span style="font-size:9pt;color:#757575">Minste beurten: </span>
            <strong>{{ minTurnsRecord.player.name }}</strong>
            <span style="font-size:9pt"> ({{ minTurnsRecord.minTurns }})</span>
          </div>
          <div *ngIf="highestFinishRecord">
            <span style="font-size:9pt;color:#757575">Hoogste finish: </span>
            <strong>{{ highestFinishRecord.player.name }}</strong>
            <span style="font-size:9pt"> ({{ highestFinishRecord.highestFinish }})</span>
          </div>
        </div>

        <!-- Page 1: all class standings -->
        <div *ngFor="let cls of classes">
          <h3 class="print-section-title">{{ cls.label }}</h3>
          <table class="print-table">
            <thead>
              <tr>
                <th style="width:32px">#</th>
                <th style="width:40px">Nr</th>
                <th>Naam</th>
                <th class="center" style="width:64px">Gewonnen</th>
                <th class="center" style="width:64px">Verloren</th>
                <th class="center" style="width:60px">+ punten</th>
                <th class="center" style="width:60px">- punten</th>
                <th class="center" style="width:44px">180</th>
              </tr>
            </thead>
            <tbody>
              <tr *ngFor="let s of cls.stats; let i = index">
                <td>{{ i + 1 }}</td>
                <td>{{ s.player.nr }}</td>
                <td><strong>{{ s.player.name }}</strong></td>
                <td class="center" style="font-weight:600">{{ s.wins }}</td>
                <td class="center">{{ s.losses }}</td>
                <td class="center">{{ s.pointsFor }}</td>
                <td class="center">{{ s.pointsAgainst }}</td>
                <td class="center" style="font-weight:600">{{ s.oneEighties || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>

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
            <tr *ngFor="let s of dutyStats; let i = index">
              <td>{{ i + 1 }}</td>
              <td>{{ s.player.nr }}</td>
              <td><strong>{{ s.player.name }}</strong></td>
              <td class="center" style="font-weight:600">{{ s.count }}</td>
            </tr>
          </tbody>
        </table>
      </div><!-- /print-only -->

    </div>
  `
})
export class StandingsComponent implements OnInit {
  private scoreService  = inject(ScoreService);
  private seasonService = inject(SeasonService);
  private dialog        = inject(MatDialog);
  private destroyRef    = inject(DestroyRef);

  classes:   { label: string; stats: PlayerStats[] }[] = [];
  dutyStats: DutyStats[] = [];
  allStats:  PlayerStats[] = [];
  selectedTabIndex = 0;
  minTurnsRecord:     PlayerStats | null = null;
  highestFinishRecord: PlayerStats | null = null;

  matchCols = ['rank', 'nr', 'name', 'wins', 'losses', 'pf', 'pa', '180s', 'edit'];
  dutyCols  = ['rank', 'nr', 'name', 'count'];

  ngOnInit(): void {
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      distinctUntilChanged(),
    ).subscribe(() => this.loadStats());
  }

  private loadStats(): void {
    const sid = this.seasonService.selectedId$.value || undefined;
    this.scoreService.getStats(sid).subscribe((s) => {
      this.allStats = s;
      this.classes = this.buildClasses(s);
      const withMinTurns = s.filter(x => x.minTurns > 0);
      this.minTurnsRecord = withMinTurns.length ? withMinTurns.reduce((b, x) => x.minTurns < b.minTurns ? x : b) : null;
      const withHF = s.filter(x => x.highestFinish > 0);
      this.highestFinishRecord = withHF.length ? withHF.reduce((b, x) => x.highestFinish > b.highestFinish ? x : b) : null;
      // Defer reset to after MatTabGroup has processed the new @ContentChildren tabs.
      // Setting the index synchronously causes a race: MatTabGroup fires selectedIndexChange
      // while rebuilding its tab list, which overwrites our value via the two-way binding.
      setTimeout(() => { this.selectedTabIndex = 0; });
    });
    this.scoreService.getDutyStats(sid).subscribe((d) => {
      this.dutyStats = d.sort((a, b) => b.count - a.count);
    });
  }

  openStatDialog(stat: PlayerStats): void {
    const sid = this.seasonService.selectedId$.value;
    if (!sid) return;
    this.dialog.open(EveningStatDialogComponent, {
      data: {
        scheduleId: sid,
        players: [{ id: stat.player.id, name: stat.player.name }],
        preselectedPlayerId: stat.player.id,
      } as EveningStatDialogData,
    }).afterClosed().subscribe(saved => {
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
    const result = [...byClass.keys()].sort().map(c => ({
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
      if (b.wins !== a.wins) return b.wins - a.wins;
      const diffA = a.pointsFor - a.pointsAgainst;
      const diffB = b.pointsFor - b.pointsAgainst;
      return diffB - diffA;
    });
  }


  print(): void {
    window.print();
  }
}
