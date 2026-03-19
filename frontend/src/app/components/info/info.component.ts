import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { distinctUntilChanged, filter, forkJoin } from 'rxjs';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatTableModule } from '@angular/material/table';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatTabsModule } from '@angular/material/tabs';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { ScheduleService } from '../../services/schedule.service';
import { SeasonService } from '../../services/season.service';
import { ScoreService } from '../../services/score.service';
import { SystemService } from '../../services/system.service';
import { PlayerInfoItem, EveningInfoItem, ScheduleInfo, BuddyPairItem, PlayerStats, Schedule, DutyStats } from '../../models';
import { environment } from '../../../environments/environment';

interface PlayerRow {
  player: PlayerInfoItem;
  cells: CellData[];
  totalMatches: number;
  eveningCount: number;
}

interface CellData {
  count: number;
  consecutive: boolean;
}

interface MatchRow {
  eveningNumber: number;
  eveningDate: string;
  opponentName: string;
  myScore: number | null;
  oppScore: number | null;
  result: 'W' | 'V' | 'G' | 'Afgemeld' | '—';
}

@Component({
  selector: 'app-info',
  standalone: true,
  imports: [CommonModule, MatCardModule, MatTableModule, MatIconModule, MatChipsModule,
            MatTabsModule, MatSelectModule, MatFormFieldModule, MatButtonModule],
  styles: [`
    .page { padding: 24px; }
    h2 { margin: 0 0 20px 0; }

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

    table { width: 100%; }

    .buddy-chip-row { display: flex; flex-wrap: wrap; gap: 4px; }
    .evening-chip { display: inline-block; background: #e3f2fd; color: #0277bd; border-radius: 12px;
                    padding: 2px 8px; font-size: 11px; font-weight: 500; }

    .legend { display: flex; gap: 16px; font-size: 12px; color: #616161; margin-bottom: 12px; align-items: center; }
    .legend-item { display: flex; align-items: center; gap: 4px; }
    .legend-box { width: 16px; height: 16px; border-radius: 3px; border: 1px solid #ccc; }

    .result-W { color: #2e7d32; font-weight: 600; }
    .result-V { color: #c62828; font-weight: 600; }
    .result-G { color: #f57f17; font-weight: 600; }
    .result-af { color: #9e9e9e; font-style: italic; }

    .duty-select { min-width: 280px; margin-bottom: 16px; }
    .duty-totals { display: flex; gap: 24px; margin-bottom: 12px; font-size: 13px; }
    .duty-totals span { background: #f5f5f5; border-radius: 6px; padding: 4px 12px; }
    .duty-totals .sec { background: #e3f2fd; color: #0277bd; }
    .duty-totals .cnt { background: #fce4ec; color: #c62828; }
    .duty-section-title { font-size: 12px; font-weight: 600; color: #616161; text-transform: uppercase;
                          letter-spacing: .4px; margin: 12px 0 4px; }
    .duty-table { border-collapse: collapse; width: 100%; font-size: 12px; }
    .duty-table th { background: #f5f5f5; font-weight: 600; text-align: left; padding: 4px 8px;
                     border-bottom: 2px solid #e0e0e0; }
    .duty-table td { padding: 3px 8px; border-bottom: 1px solid #f0f0f0; }
    .duty-table tr:hover td { background: #fafafa; }
    .duty-empty { color: #9e9e9e; font-style: italic; font-size: 12px; padding: 8px 0; }
    .duty-evening-table { border-collapse: collapse; font-size: 12px; margin-bottom: 16px; }
    .duty-evening-table th { background: #f5f5f5; font-weight: 600; text-align: center; padding: 4px 10px;
                              border-bottom: 2px solid #e0e0e0; border-right: 1px solid #e0e0e0; }
    .duty-evening-table td { text-align: center; padding: 3px 10px; border-bottom: 1px solid #f0f0f0;
                              border-right: 1px solid #f0f0f0; }
    .duty-evening-table .sec-cell { color: #0277bd; font-weight: 600; }
    .duty-evening-table .cnt-cell { color: #c62828; font-weight: 600; }
    .duty-evening-table .tot-cell { font-weight: 700; }
    .server-meta { display: flex; align-items: center; gap: 16px; margin-bottom: 16px; }
    .version-chip {
      background: #e8f5e9; color: #2e7d32; border-radius: 12px;
      padding: 4px 12px; font-size: 13px; font-weight: 500;
    }
    .log-box {
      background: #1e1e1e; color: #d4d4d4; font-family: monospace;
      font-size: 12px; line-height: 1.5; padding: 12px 16px;
      border-radius: 6px; max-height: 480px; overflow-y: auto;
      white-space: pre-wrap; word-break: break-all;
    }
    .log-empty { color: #9e9e9e; font-style: italic; font-size: 13px; }
  `],
  template: `
    <div class="page">
      <h2>Seizoen Info</h2>

      <p *ngIf="!info" style="color:#9e9e9e">Selecteer een seizoen om info te zien.</p>

      <mat-tab-group *ngIf="info" animationDuration="150ms" color="primary">

        <!-- Tab 1: Verdeling per avond -->
        <mat-tab label="Verdeling">
          <div style="padding-top:16px">
            <div class="legend">
              <span>Legenda:</span>
              <span class="legend-item">
                <span class="legend-box" style="background:#e8f5e9"></span> Speelt op deze avond
              </span>
              <span class="legend-item">
                <span class="legend-box" style="background:#fff3e0"></span> Opeenvolgende avonden
              </span>
            </div>
            <mat-card>
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
          </div>
        </mat-tab>

        <!-- Tab 2: Spelers -->
        <mat-tab label="Spelers">
          <div style="padding-top:16px">
            <mat-card>
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

                  <ng-container matColumnDef="buddy">
                    <th mat-header-cell *matHeaderCellDef>Koppel</th>
                    <td mat-cell *matCellDef="let row">
                      <ng-container *ngIf="getBuddy(row.player.id) as pair; else noBuddy">
                        <span style="font-size:13px">
                          {{ pair.partnerNr }} {{ pair.partnerName }}
                        </span>
                        <div class="buddy-chip-row" style="margin-top:2px" *ngIf="pair.eveningNrs.length > 0">
                          <span class="evening-chip" *ngFor="let nr of pair.eveningNrs">Av.{{ nr }}</span>
                        </div>
                      </ng-container>
                      <ng-template #noBuddy><span style="color:#bdbdbd;font-size:12px">—</span></ng-template>
                    </td>
                  </ng-container>

                  <tr mat-header-row *matHeaderRowDef="summaryCols"></tr>
                  <tr mat-row *matRowDef="let row; columns: summaryCols;"></tr>
                </table>
              </mat-card-content>
            </mat-card>
          </div>
        </mat-tab>

        <!-- Tab 3: Statistieken -->
        <mat-tab label="Statistieken">
          <div style="padding-top:16px">
            <mat-card *ngIf="statRows.length > 0">
              <mat-card-header><mat-card-title>Beurtstatistieken &amp; Records</mat-card-title></mat-card-header>
              <mat-card-content>
                <table mat-table [dataSource]="statRows" style="width:100%">

                  <ng-container matColumnDef="nr">
                    <th mat-header-cell *matHeaderCellDef style="width:48px">Nr</th>
                    <td mat-cell *matCellDef="let s">{{ s.player.nr }}</td>
                  </ng-container>

                  <ng-container matColumnDef="name">
                    <th mat-header-cell *matHeaderCellDef>Naam</th>
                    <td mat-cell *matCellDef="let s"><strong>{{ s.player.name }}</strong></td>
                  </ng-container>

                  <ng-container matColumnDef="minTurns">
                    <th mat-header-cell *matHeaderCellDef style="width:90px;text-align:center" title="Minste beurten in een gewonnen leg">Min. beurten</th>
                    <td mat-cell *matCellDef="let s" style="text-align:center">{{ s.minTurns || '—' }}</td>
                  </ng-container>

                  <ng-container matColumnDef="avgTurns">
                    <th mat-header-cell *matHeaderCellDef style="width:90px;text-align:center" title="Gemiddeld beurten per gewonnen leg">Gem. beurten</th>
                    <td mat-cell *matCellDef="let s" style="text-align:center">
                      {{ s.avgTurns ? (s.avgTurns | number:'1.1-1') : '—' }}
                    </td>
                  </ng-container>

                  <ng-container matColumnDef="avgScore">
                    <th mat-header-cell *matHeaderCellDef style="width:100px;text-align:center" title="Gemiddelde score per beurt (≈ 501 / gem. beurten)">Gem. score/beurt</th>
                    <td mat-cell *matCellDef="let s" style="text-align:center">
                      {{ s.avgScorePerTurn ? (s.avgScorePerTurn | number:'1.1-1') : '—' }}
                    </td>
                  </ng-container>

                  <ng-container matColumnDef="180s">
                    <th mat-header-cell *matHeaderCellDef style="width:60px;text-align:center">180's</th>
                    <td mat-cell *matCellDef="let s" style="text-align:center;font-weight:600;color:#7b1fa2">
                      {{ s.oneEighties || '—' }}
                    </td>
                  </ng-container>

                  <ng-container matColumnDef="hf">
                    <th mat-header-cell *matHeaderCellDef style="width:90px;text-align:center">Hoogste finish</th>
                    <td mat-cell *matCellDef="let s" style="text-align:center;font-weight:600;color:#0277bd">
                      {{ s.highestFinish || '—' }}
                    </td>
                  </ng-container>

                  <tr mat-header-row *matHeaderRowDef="statCols"></tr>
                  <tr mat-row *matRowDef="let row; columns: statCols;"></tr>
                </table>
              </mat-card-content>
            </mat-card>
            <p *ngIf="statRows.length === 0" style="color:#9e9e9e;padding-top:8px">
              Nog geen gespeelde wedstrijden.
            </p>
          </div>
        </mat-tab>

        <!-- Tab 4: Openstaande wedstrijden -->
        <mat-tab>
          <ng-template mat-tab-label>
            Openstaand
            <span *ngIf="openMatchRows.length > 0"
                  style="margin-left:6px;background:#e53935;color:white;border-radius:10px;
                         padding:1px 7px;font-size:11px;font-weight:600;line-height:18px">
              {{ openMatchRows.length }}
            </span>
          </ng-template>
          <div style="padding-top:16px">
            <p *ngIf="openMatchRows.length === 0" style="color:#9e9e9e;padding-top:8px">
              Alle wedstrijden zijn gespeeld of afgemeld.
            </p>
            <mat-card *ngIf="openMatchRows.length > 0">
              <mat-card-content>
                <div style="display:flex;justify-content:flex-end;margin-bottom:8px">
                  <button mat-stroked-button (click)="printOpenMatches()">
                    <mat-icon>print</mat-icon> Afdrukken
                  </button>
                </div>
                <table mat-table [dataSource]="openMatchRows" style="width:100%">

                  <ng-container matColumnDef="evening">
                    <th mat-header-cell *matHeaderCellDef style="width:80px">Avond</th>
                    <td mat-cell *matCellDef="let r">{{ r.eveningNumber }}</td>
                  </ng-container>

                  <ng-container matColumnDef="date">
                    <th mat-header-cell *matHeaderCellDef style="width:120px">Datum</th>
                    <td mat-cell *matCellDef="let r">{{ r.eveningDate | date:'d MMM yyyy' }}</td>
                  </ng-container>

                  <ng-container matColumnDef="playerA">
                    <th mat-header-cell *matHeaderCellDef>Speler A</th>
                    <td mat-cell *matCellDef="let r"><strong>{{ r.playerAName }}</strong></td>
                  </ng-container>

                  <ng-container matColumnDef="playerB">
                    <th mat-header-cell *matHeaderCellDef>Speler B</th>
                    <td mat-cell *matCellDef="let r"><strong>{{ r.playerBName }}</strong></td>
                  </ng-container>

                  <tr mat-header-row *matHeaderRowDef="openCols"></tr>
                  <tr mat-row *matRowDef="let row; columns: openCols;"></tr>
                </table>
              </mat-card-content>
            </mat-card>
          </div>
        </mat-tab>

        <!-- Tab 5: Wedstrijdoverzicht per speler -->
        <mat-tab label="Wedstrijden">
          <div style="padding-top:16px">
            <mat-form-field style="min-width:280px;margin-bottom:16px" subscriptSizing="dynamic">
              <mat-label>Speler</mat-label>
              <mat-select [(value)]="selectedPlayerId">
                <mat-option value="">— Kies een speler —</mat-option>
                <mat-option *ngFor="let p of sortedPlayers" [value]="p.id">
                  {{ p.nr }} – {{ p.name }}
                </mat-option>
              </mat-select>
            </mat-form-field>

            <div *ngIf="selectedPlayerId" style="display:flex;justify-content:flex-end;gap:8px;margin-bottom:8px">
              <button mat-stroked-button (click)="printPendingMatches()">
                <mat-icon>print</mat-icon> Nog te spelen
              </button>
              <button mat-stroked-button (click)="exportCalendar()">
                <mat-icon>event</mat-icon> Agenda exporteren (.ics)
              </button>
            </div>

            <mat-card *ngIf="selectedPlayerId">
              <mat-card-content>
                <table mat-table [dataSource]="playerMatchRows" style="width:100%">

                  <ng-container matColumnDef="evening">
                    <th mat-header-cell *matHeaderCellDef style="width:80px">Avond</th>
                    <td mat-cell *matCellDef="let r">{{ r.eveningNumber }}</td>
                  </ng-container>

                  <ng-container matColumnDef="date">
                    <th mat-header-cell *matHeaderCellDef style="width:110px">Datum</th>
                    <td mat-cell *matCellDef="let r">{{ r.eveningDate | date:'d MMM yyyy' }}</td>
                  </ng-container>

                  <ng-container matColumnDef="opponent">
                    <th mat-header-cell *matHeaderCellDef>Tegenstander</th>
                    <td mat-cell *matCellDef="let r"><strong>{{ r.opponentName }}</strong></td>
                  </ng-container>

                  <ng-container matColumnDef="score">
                    <th mat-header-cell *matHeaderCellDef style="width:80px;text-align:center">Score</th>
                    <td mat-cell *matCellDef="let r" style="text-align:center">
                      <span *ngIf="r.myScore !== null">{{ r.myScore }} – {{ r.oppScore }}</span>
                      <span *ngIf="r.myScore === null" style="color:#bdbdbd">—</span>
                    </td>
                  </ng-container>

                  <ng-container matColumnDef="result">
                    <th mat-header-cell *matHeaderCellDef style="width:90px;text-align:center">Uitslag</th>
                    <td mat-cell *matCellDef="let r" style="text-align:center">
                      <span [class.result-W]="r.result === 'W'"
                            [class.result-V]="r.result === 'V'"
                            [class.result-G]="r.result === 'G'"
                            [class.result-af]="r.result === 'Afgemeld'">
                        {{ r.result }}
                      </span>
                    </td>
                  </ng-container>

                  <tr mat-header-row *matHeaderRowDef="matchCols"></tr>
                  <tr mat-row *matRowDef="let row; columns: matchCols;"></tr>
                </table>

                <p *ngIf="playerMatchRows.length === 0" style="color:#9e9e9e;text-align:center;padding:24px 0;margin:0">
                  Geen wedstrijden gevonden voor deze speler.
                </p>
              </mat-card-content>
            </mat-card>
          </div>
        </mat-tab>

        <!-- Tab 6: Schrijver / Teller -->
        <mat-tab label="Schrijver/Teller">
          <div style="padding-top:16px">
            <mat-form-field class="duty-select" subscriptSizing="dynamic">
              <mat-label>Speler</mat-label>
              <mat-select [(value)]="selectedDutyPlayerId">
                <mat-option value="">— Kies een speler —</mat-option>
                <mat-option *ngFor="let d of dutyStats" [value]="d.player.id">
                  {{ d.player.nr }} – {{ d.player.name }}
                </mat-option>
              </mat-select>
            </mat-form-field>

            <ng-container *ngIf="selectedDutyPlayer as d">
              <div class="duty-totals">
                <span>Totaal: <strong>{{ d.count }}</strong></span>
                <span class="sec">Schrijver: <strong>{{ d.secretaryCount }}</strong></span>
                <span class="cnt">Teller: <strong>{{ d.counterCount }}</strong></span>
              </div>

              <div class="duty-section-title">Per avond</div>
              <table class="duty-evening-table">
                <thead>
                  <tr>
                    <th>Avond</th>
                    <th>Schrijver</th>
                    <th>Teller</th>
                    <th>Totaal</th>
                  </tr>
                </thead>
                <tbody>
                  <tr *ngFor="let row of dutyByEvening">
                    <td>{{ row.eveningNr }}</td>
                    <td class="sec-cell">{{ row.sec || '—' }}</td>
                    <td class="cnt-cell">{{ row.cnt || '—' }}</td>
                    <td class="tot-cell">{{ row.total }}</td>
                  </tr>
                </tbody>
              </table>

              <div class="duty-section-title">Geschreven wedstrijden</div>
              <p class="duty-empty" *ngIf="d.secretaryMatches.length === 0">Geen.</p>
              <table class="duty-table" *ngIf="d.secretaryMatches.length > 0">
                <thead><tr><th>Avond</th><th>Speler A</th><th></th><th>Speler B</th></tr></thead>
                <tbody>
                  <tr *ngFor="let m of d.secretaryMatches">
                    <td>{{ m.eveningNr || '—' }}</td>
                    <td>{{ m.playerANr }} {{ m.playerAName }}</td>
                    <td style="color:#999;text-align:center">vs</td>
                    <td>{{ m.playerBNr }} {{ m.playerBName }}</td>
                  </tr>
                </tbody>
              </table>

              <div class="duty-section-title" style="margin-top:16px">Getelde wedstrijden</div>
              <p class="duty-empty" *ngIf="d.counterMatches.length === 0">Geen.</p>
              <table class="duty-table" *ngIf="d.counterMatches.length > 0">
                <thead><tr><th>Avond</th><th>Speler A</th><th></th><th>Speler B</th></tr></thead>
                <tbody>
                  <tr *ngFor="let m of d.counterMatches">
                    <td>{{ m.eveningNr || '—' }}</td>
                    <td>{{ m.playerANr }} {{ m.playerAName }}</td>
                    <td style="color:#999;text-align:center">vs</td>
                    <td>{{ m.playerBNr }} {{ m.playerBName }}</td>
                  </tr>
                </tbody>
              </table>
            </ng-container>
          </div>
        </mat-tab>

        <!-- Tab 7: Server logs -->
        <mat-tab label="Server">
          <div style="padding-top:16px">
            <div class="server-meta">
              <span class="version-chip">{{ version }}</span>
              <button mat-stroked-button (click)="refreshLogs()">
                <mat-icon>refresh</mat-icon> Vernieuwen
              </button>

            </div>
            <div *ngIf="logsLoading" style="color:#9e9e9e;font-size:13px">Laden...</div>
            <div *ngIf="!logsLoading && logs.length === 0" class="log-empty">
              Nog geen log regels.
            </div>
            <div *ngIf="!logsLoading && logs.length > 0" class="log-box">{{ logs.join('\n') }}</div>
          </div>
        </mat-tab>

      </mat-tab-group>
    </div>
  `,
})
export class InfoComponent implements OnInit {
  private scheduleService = inject(ScheduleService);
  private seasonService   = inject(SeasonService);
  private scoreService    = inject(ScoreService);
  private systemService   = inject(SystemService);
  private destroyRef      = inject(DestroyRef);

  info:     ScheduleInfo | null = null;
  schedule: Schedule | null     = null;
  playerRows: PlayerRow[]       = [];
  statRows:   PlayerStats[]     = [];
  selectedPlayerId              = '';

  dutyStats:          DutyStats[] = [];
  selectedDutyPlayerId = '';

  get selectedDutyPlayer(): DutyStats | null {
    return this.dutyStats.find(d => d.player.id === this.selectedDutyPlayerId) ?? null;
  }

  get dutyByEvening(): { eveningNr: number; sec: number; cnt: number; total: number }[] {
    const d = this.selectedDutyPlayer;
    if (!d) return [];
    const map = new Map<number, { sec: number; cnt: number }>();
    for (const m of d.secretaryMatches) {
      const nr = m.eveningNr || 0;
      const e = map.get(nr) ?? { sec: 0, cnt: 0 };
      e.sec++;
      map.set(nr, e);
    }
    for (const m of d.counterMatches) {
      const nr = m.eveningNr || 0;
      const e = map.get(nr) ?? { sec: 0, cnt: 0 };
      e.cnt++;
      map.set(nr, e);
    }
    return [...map.entries()]
      .sort((a, b) => a[0] - b[0])
      .map(([eveningNr, { sec, cnt }]) => ({ eveningNr, sec, cnt, total: sec + cnt }));
  }

  version     = environment.version;
  logs:        string[] = [];
  logsLoading = false;

  summaryCols = ['nr', 'name', 'eveningCount', 'totalMatches', 'consecutive', 'buddy'];
  statCols    = ['nr', 'name', 'minTurns', 'avgTurns', 'avgScore', '180s', 'hf'];
  matchCols   = ['evening', 'date', 'opponent', 'score', 'result'];
  openCols    = ['evening', 'date', 'playerA', 'playerB'];

  ngOnInit(): void {
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      distinctUntilChanged(),
      filter(id => !!id),
    ).subscribe(id => this.load(id));
    this.refreshLogs();
  }

  refreshLogs(): void {
    this.logsLoading = true;
    this.systemService.getLogs().subscribe({
      next: ({ logs }) => { this.logs = logs; this.logsLoading = false; },
      error: ()        => { this.logsLoading = false; },
    });
  }

  get sortedPlayers(): PlayerInfoItem[] {
    if (!this.info) return [];
    return [...this.info.players].sort((a, b) => (parseInt(a.nr) || 9999) - (parseInt(b.nr) || 9999));
  }

  get playerMatchRows(): MatchRow[] {
    if (!this.schedule || !this.selectedPlayerId || !this.info) return [];
    const rows: MatchRow[] = [];
    const playerMap = new Map(this.info.players.map(p => [p.id, p]));

    for (const ev of this.schedule.evenings) {
      if (ev.isInhaalAvond) continue;
      for (const m of ev.matches) {
        const isA = m.playerA === this.selectedPlayerId;
        const isB = m.playerB === this.selectedPlayerId;
        if (!isA && !isB) continue;

        const opponentId = isA ? m.playerB : m.playerA;
        const opp = playerMap.get(opponentId);
        const opponentName = opp ? `${opp.nr} ${opp.name}` : opponentId.slice(0, 8);

        const myScore  = isA ? m.scoreA : m.scoreB;
        const oppScore = isA ? m.scoreB : m.scoreA;

        let result: MatchRow['result'] = '—';
        if (m.reportedBy && !m.played) {
          result = 'Afgemeld';
        } else if (m.played && myScore !== null && oppScore !== null) {
          result = myScore > oppScore ? 'W' : myScore < oppScore ? 'V' : 'G';
        }

        rows.push({ eveningNumber: ev.number, eveningDate: ev.date, opponentName, myScore, oppScore, result });
      }
    }

    return rows.sort((a, b) => a.eveningNumber - b.eveningNumber);
  }

  get openMatchRows(): { eveningNumber: number; eveningDate: string; playerAName: string; playerBName: string }[] {
    if (!this.schedule || !this.info) return [];
    const playerMap = new Map(this.info.players.map(p => [p.id, p]));
    const rows: { eveningNumber: number; eveningDate: string; playerAName: string; playerBName: string }[] = [];
    for (const ev of this.schedule.evenings) {
      if (ev.isInhaalAvond) continue;
      for (const m of ev.matches) {
        if (m.played || m.reportedBy) continue;
        const pA = playerMap.get(m.playerA);
        const pB = playerMap.get(m.playerB);
        rows.push({
          eveningNumber: ev.number,
          eveningDate:   ev.date,
          playerAName:   pA ? `${pA.nr} ${pA.name}` : m.playerA.slice(0, 8),
          playerBName:   pB ? `${pB.nr} ${pB.name}` : m.playerB.slice(0, 8),
        });
      }
    }
    return rows.sort((a, b) => a.eveningNumber - b.eveningNumber);
  }

  private load(scheduleId: string): void {
    forkJoin({
      info:     this.scheduleService.getInfo(scheduleId),
      schedule: this.scheduleService.getById(scheduleId),
      stats:    this.scoreService.getStats(scheduleId),
      duties:   this.scoreService.getDutyStats(scheduleId),
    }).subscribe({
      next: ({ info, schedule, stats, duties }) => {
        this.info     = info;
        this.schedule = schedule;
        this.playerRows = this.buildPlayerRows(info);
        this.statRows = stats
          .filter(s => s.played > 0)
          .sort((a, b) => (parseInt(a.player.nr || '0')) - (parseInt(b.player.nr || '0')));
        this.dutyStats = duties
          .filter(d => d.count > 0)
          .sort((a, b) => (parseInt(a.player.nr) || 9999) - (parseInt(b.player.nr) || 9999));
      },
      error: () => { this.info = null; this.schedule = null; this.playerRows = []; this.statRows = []; this.dutyStats = []; },
    });
  }

  private buildPlayerRows(info: ScheduleInfo): PlayerRow[] {
    const lookup = new Map<string, Map<string, number>>();
    for (const cell of info.matrix) {
      if (!lookup.has(cell.playerId)) lookup.set(cell.playerId, new Map());
      lookup.get(cell.playerId)!.set(cell.eveningId, cell.count);
    }

    const evenings = [...info.evenings].sort((a, b) => a.number - b.number);

    return info.players.map(player => {
      const byEvening = lookup.get(player.id) ?? new Map<string, number>();
      const counts    = evenings.map(ev => byEvening.get(ev.id) ?? 0);
      const consecutive = counts.map((c, i) => {
        if (c === 0) return false;
        return (i > 0 && counts[i - 1] > 0) || (i < counts.length - 1 && counts[i + 1] > 0);
      });
      const cells        = counts.map((count, i) => ({ count, consecutive: consecutive[i] }));
      const totalMatches = counts.reduce((s, c) => s + c, 0);
      const eveningCount = counts.filter(c => c > 0).length;
      return { player, cells, totalMatches, eveningCount };
    }).filter(row => row.totalMatches > 0)
      .sort((a, b) => (parseInt(a.player.nr) || 9999) - (parseInt(b.player.nr) || 9999));
  }

  printPendingMatches(): void {
    const player = this.sortedPlayers.find(p => p.id === this.selectedPlayerId);
    if (!player) return;
    const pending = this.playerMatchRows.filter(r => r.result === '—' || r.result === 'Afgemeld');
    const compName = this.schedule?.competitionName ?? '';
    const rowsHtml = pending.map(r => `
      <tr>
        <td>${r.eveningNumber}</td>
        <td>${new Date(r.eveningDate).toLocaleDateString('nl-NL', { day: 'numeric', month: 'long', year: 'numeric' })}</td>
        <td><strong>${r.opponentName}</strong></td>
        <td style="color:${r.result === 'Afgemeld' ? '#c62828' : '#555'}">${r.result === 'Afgemeld' ? 'Afgemeld' : ''}</td>
      </tr>`).join('');
    const html = `<!DOCTYPE html><html><head><meta charset="utf-8">
      <title>Nog te spelen – ${player.nr} ${player.name}</title>
      <style>
        body { font-family: Arial, sans-serif; font-size: 12px; padding: 16px; }
        h2 { font-size: 15px; margin-bottom: 4px; }
        p  { margin: 0 0 12px; color: #555; font-size: 11px; }
        table { border-collapse: collapse; width: 100%; }
        th { background: #f5f5f5; font-weight: 600; text-align: left; padding: 5px 8px; border-bottom: 2px solid #ccc; }
        td { padding: 4px 8px; border-bottom: 1px solid #eee; }
        @media print { @page { size: A4 portrait; margin: 15mm; } }
      </style></head><body>
      <h2>Nog te spelen wedstrijden – ${player.nr} ${player.name}</h2>
      <p>${compName}</p>
      <table><thead><tr><th>Avond</th><th>Datum</th><th>Tegenstander</th><th>Status</th></tr></thead>
      <tbody>${pending.length ? rowsHtml : '<tr><td colspan="3" style="color:#999;padding:12px 8px">Geen openstaande wedstrijden.</td></tr>'}</tbody></table>
      <script>window.onload = () => { window.print(); }<\/script>
      </body></html>`;
    const w = window.open('', '_blank');
    if (w) { w.document.write(html); w.document.close(); }
  }

  exportCalendar(): void {
    const player = this.sortedPlayers.find(p => p.id === this.selectedPlayerId);
    if (!player) return;
    const compName = this.schedule?.competitionName ?? 'Dartclub';

    const pad = (n: number) => String(n).padStart(2, '0');
    const toIcsDate = (iso: string) => {
      const d = new Date(iso);
      return `${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}`;
    };
    // Simple UID generator
    const uid = (i: number) => `dart-${this.selectedPlayerId.slice(0, 8)}-${i}@grolzicht`;

    const events = this.playerMatchRows.map((r, i) => [
      'BEGIN:VEVENT',
      `UID:${uid(i)}`,
      `DTSTART;VALUE=DATE:${toIcsDate(r.eveningDate)}`,
      `DTEND;VALUE=DATE:${toIcsDate(r.eveningDate)}`,
      `SUMMARY:Avond ${r.eveningNumber} – vs ${r.opponentName}`,
      `DESCRIPTION:${compName}\\nAvond ${r.eveningNumber} – ${player.nr} ${player.name} vs ${r.opponentName}`,
      'END:VEVENT',
    ].join('\r\n')).join('\r\n');

    const ics = [
      'BEGIN:VCALENDAR',
      'VERSION:2.0',
      'PRODID:-//DartScheduler//NL',
      'CALSCALE:GREGORIAN',
      'METHOD:PUBLISH',
      events,
      'END:VCALENDAR',
    ].join('\r\n');

    const blob = new Blob([ics], { type: 'text/calendar;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `dart-${player.nr}-${player.name.replace(/[^a-zA-Z0-9]/g, '_')}.ics`;
    a.click();
    URL.revokeObjectURL(url);
  }

  printOpenMatches(): void {
    const rows = this.openMatchRows;
    const name = this.schedule?.competitionName ?? '';
    const rowsHtml = rows.map(r => `
      <tr>
        <td>${r.eveningNumber}</td>
        <td>${new Date(r.eveningDate).toLocaleDateString('nl-NL', { day: 'numeric', month: 'long', year: 'numeric' })}</td>
        <td><strong>${r.playerAName}</strong></td>
        <td style="color:#999;text-align:center">vs</td>
        <td><strong>${r.playerBName}</strong></td>
      </tr>`).join('');
    const html = `<!DOCTYPE html><html><head><meta charset="utf-8">
      <title>Openstaande wedstrijden</title>
      <style>
        body { font-family: Arial, sans-serif; font-size: 12px; padding: 16px; }
        h2 { font-size: 15px; margin-bottom: 12px; }
        table { border-collapse: collapse; width: 100%; }
        th { background: #f5f5f5; font-weight: 600; text-align: left; padding: 5px 8px; border-bottom: 2px solid #ccc; }
        td { padding: 4px 8px; border-bottom: 1px solid #eee; }
        @media print { @page { size: A4 portrait; margin: 15mm; } }
      </style></head><body>
      <h2>Openstaande wedstrijden${name ? ' — ' + name : ''}</h2>
      <table><thead><tr><th>Avond</th><th>Datum</th><th>Speler A</th><th></th><th>Speler B</th></tr></thead>
      <tbody>${rowsHtml}</tbody></table>
      <script>window.onload = () => { window.print(); }<\/script>
      </body></html>`;
    const w = window.open('', '_blank');
    if (w) { w.document.write(html); w.document.close(); }
  }

  getBuddy(playerId: string): { partnerNr: string; partnerName: string; eveningNrs: number[] } | null {
    if (!this.info) return null;
    const pair = this.info.buddyPairs.find(p => p.playerAId === playerId || p.playerBId === playerId);
    if (!pair) return null;
    const isA = pair.playerAId === playerId;
    return {
      partnerNr:   isA ? pair.playerBNr   : pair.playerANr,
      partnerName: isA ? pair.playerBName : pair.playerAName,
      eveningNrs:  pair.eveningNrs,
    };
  }

  getStreaks(row: PlayerRow): number[][] {
    const evenings = this.info ? [...this.info.evenings].sort((a, b) => a.number - b.number) : [];
    const activeNrs: number[] = row.cells
      .map((cell, i) => cell.count > 0 ? evenings[i]?.number : null)
      .filter((n): n is number => n !== null);

    const streaks: number[][] = [];
    let current: number[] = [];
    for (let i = 0; i < activeNrs.length; i++) {
      if (current.length === 0 || activeNrs[i] === activeNrs[i - 1] + 1) {
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
