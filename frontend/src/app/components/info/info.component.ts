import { Component, inject, OnInit, DestroyRef, signal } from '@angular/core';
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
import { PlayerInfoItem, ScheduleInfo, PlayerStats, Schedule, DutyStats } from '../../models';

interface PlayerRow {
  player: PlayerInfoItem;
  cells: CellData[];
  totalMatches: number;
  eveningCount: number;
}

interface CellData {
  count: number;
  level: 'none' | 'ok' | 'soft' | 'orange' | 'hard';
  consec: boolean; // part of a 2-evening run — shown bold but not a color violation
}

interface PlayerStatusRow {
  player: PlayerInfoItem;
  gespeeld: number;
  teSpelen: number;
  inTeHalen: number;
}

interface MatchRow {
  eveningNumber: number;
  eveningDate: string;
  playerAName: string;
  playerBName: string;
  opponentName: string;
  myScore: number | null;
  oppScore: number | null;
  result: 'W' | 'V' | 'G' | 'Afgemeld' | '—';
  reportedBy: string; // who reported the absence (empty if not cancelled)
  isCatchUp: boolean;
  playedDate: string; // set for catch-up matches that have been played
}

@Component({
    selector: 'app-info',
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
    .cell-none   { background: #fafafa; color: #bdbdbd; text-align: center; }
    .cell-ok     { background: #e8f5e9; color: #2e7d32; text-align: center; font-weight: 600; }
    .cell-soft   { background: #ffe066; color: #7a6000; text-align: center; font-weight: 600; }
    .cell-orange { background: #ff9900; color: #6b3e00; text-align: center; font-weight: 600; }
    .cell-hard   { background: #e53935; color: #ffffff; text-align: center; font-weight: 600; }
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
  `],
    template: `
    <div class="page">
      <h2>Seizoen Info</h2>
    
      @if (!info()) {
        <p style="color:#9e9e9e">Selecteer een seizoen om info te zien.</p>
      }

      @if (info()) {
        <mat-tab-group animationDuration="150ms" color="primary">
          <!-- Tab 1: Verdeling per avond -->
          <mat-tab label="Verdeling">
            <div style="padding-top:16px">
              <div class="legend">
                <span>Legenda:</span>
                <span class="legend-item">
                  <span class="legend-box" style="background:#fafafa;border-color:#bdbdbd"></span> Geen wedstrijden
                </span>
                <span class="legend-item">
                  <span class="legend-box" style="background:#e8f5e9"></span> Ok (2+ wedstrijden)
                </span>
                <span class="legend-item">
                  <span class="legend-box" style="background:#ffe066"></span> Soft (1 wedstrijd / buddy mismatch 1e keer)
                </span>
                <span class="legend-item">
                  <span style="text-decoration:underline;font-size:13px;min-width:16px;text-align:center">2</span> 2 opeenvolgende avonden (onderstreept)
                </span>
                <span class="legend-item">
                  <span class="legend-box" style="background:#e53935"></span> Hard (&gt;4 wedstrijden / 3+ opeenvolgende / buddy mismatch 2e keer / &gt;3 avonden gap)
                </span>
              </div>
              <mat-card>
                <mat-card-content>
                  <div class="matrix-wrapper">
                    <table class="matrix-table">
                      <thead>
                        <tr>
                          <th class="th-player">Speler</th>
                          @for (ev of info()!.evenings; track ev) {
                            <th style="text-align:center">
                              Av.{{ ev.number }}<br>
                              <span style="font-weight:400;font-size:10px;color:#757575">{{ ev.date | date:'d/M' }}</span>
                            </th>
                          }
                          <th class="th-total">Totaal</th>
                        </tr>
                      </thead>
                      <tbody>
                        @for (row of playerRows(); track row) {
                          <tr>
                            <td class="td-player">
                              <strong>{{ row.player.nr }}</strong>{{ row.player.name }}
                            </td>
                            @for (cell of row.cells; track cell) {
                              <td [class]="'cell-' + cell.level" [style.text-decoration]="cell.consec ? 'underline' : null">
                                {{ cell.count || '' }}
                              </td>
                            }
                            <td class="td-total">{{ row.totalMatches }}</td>
                          </tr>
                        }
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
                  <table mat-table [dataSource]="playerRows()" style="width:100%">
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
                        @for (streak of getStreaks(row); track streak) {
                          <span style="margin-right:6px;font-size:12px;color:#e65100">
                            Av.{{ streak.join(', ') }}
                          </span>
                        }
                        @if (getStreaks(row).length === 0) {
                          <span style="color:#bdbdbd;font-size:12px">—</span>
                        }
                      </td>
                    </ng-container>
                    <ng-container matColumnDef="buddy">
                      <th mat-header-cell *matHeaderCellDef>Koppel</th>
                      <td mat-cell *matCellDef="let row">
                        @if (getBuddy(row.player.id); as pair) {
                          <span style="font-size:13px">
                            {{ pair.partnerNr }} {{ pair.partnerName }}
                          </span>
                          @if (pair.eveningNrs.length > 0) {
                            <div class="buddy-chip-row" style="margin-top:2px">
                              @for (nr of pair.eveningNrs; track nr) {
                                <span class="evening-chip">Av.{{ nr }}</span>
                              }
                            </div>
                          }
                        } @else {
                          <span style="color:#bdbdbd;font-size:12px">—</span>
                        }
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
              @if (statRows().length > 0) {
                <mat-card>
                  <mat-card-header><mat-card-title>Beurtstatistieken &amp; Records</mat-card-title></mat-card-header>
                  <mat-card-content>
                    <table mat-table [dataSource]="statRows()" style="width:100%">
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
              }
              @if (statRows().length === 0) {
                <p style="color:#9e9e9e;padding-top:8px">
                  Nog geen gespeelde wedstrijden.
                </p>
              }
            </div>
          </mat-tab>
          <!-- Tab 4: Openstaande wedstrijden -->
          <mat-tab>
            <ng-template mat-tab-label>
              Openstaand
              @if (openMatchRows.length > 0) {
                <span
                  style="margin-left:6px;background:#e53935;color:white;border-radius:10px;
                         padding:1px 7px;font-size:11px;font-weight:600;line-height:18px">
                  {{ openMatchRows.length }}
                </span>
              }
            </ng-template>
            <div style="padding-top:16px">
              @if (openMatchRows.length === 0) {
                <p style="color:#9e9e9e;padding-top:8px">
                  Alle wedstrijden zijn gespeeld of afgemeld.
                </p>
              }
              @if (openMatchRows.length > 0) {
                <mat-card>
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
              }
            </div>
          </mat-tab>
          <!-- Tab 5: Wedstrijdoverzicht per speler -->
          <mat-tab label="Wedstrijden">
            <div style="padding-top:16px">
              <mat-form-field style="min-width:280px;margin-bottom:16px" subscriptSizing="dynamic">
                <mat-label>Speler</mat-label>
                <mat-select [value]="selectedPlayerId()" (valueChange)="selectedPlayerId.set($event)">
                  <mat-option value="">— Kies een speler —</mat-option>
                  @for (p of sortedPlayers; track p) {
                    <mat-option [value]="p.id">
                      {{ p.nr }} – {{ p.name }}
                    </mat-option>
                  }
                </mat-select>
              </mat-form-field>
              @if (selectedPlayerId()) {
                <div style="display:flex;justify-content:flex-end;gap:8px;margin-bottom:8px">
                  <button mat-stroked-button (click)="printPendingMatches()">
                    <mat-icon>print</mat-icon> Nog te spelen
                  </button>
                  <button mat-stroked-button (click)="exportCalendar()">
                    <mat-icon>event</mat-icon> Agenda exporteren (.ics)
                  </button>
                  <button mat-stroked-button (click)="emailPendingMatches()"
                          [disabled]="!selectedPlayerEmail()"
                          [title]="selectedPlayerEmail() ? 'Mail openstaande wedstrijden naar ' + selectedPlayerEmail() : 'Geen e-mailadres bekend'">
                    <mat-icon>email</mat-icon> Mailen
                  </button>
                </div>
              }
              @if (selectedPlayerId()) {
                <mat-card>
                  <mat-card-content>
                    <table mat-table [dataSource]="playerMatchRows" style="width:100%">
                      <ng-container matColumnDef="evening">
                        <th mat-header-cell *matHeaderCellDef style="width:80px">Avond</th>
                        <td mat-cell *matCellDef="let r">{{ r.eveningNumber }}</td>
                      </ng-container>
                      <ng-container matColumnDef="date">
                        <th mat-header-cell *matHeaderCellDef style="width:110px">Datum</th>
                        <td mat-cell *matCellDef="let r">
                          <span [style.color]="r.isCatchUp ? '#7b1fa2' : null">
                            {{ r.eveningDate | date:'d MMM yyyy' }}
                            @if (r.isCatchUp) {
                              <span style="font-size:10px;font-weight:600;margin-left:2px">inhaal</span>
                            }
                          </span>
                        </td>
                      </ng-container>
                      <ng-container matColumnDef="playedDate">
                        <th mat-header-cell *matHeaderCellDef style="width:110px">Gespeeld op</th>
                        <td mat-cell *matCellDef="let r">
                          @if (r.isCatchUp && r.playedDate) {
                            <span style="color:#7b1fa2;font-weight:500">
                              {{ r.playedDate | date:'d MMM yyyy' }}
                            </span>
                          }
                          @if (!r.isCatchUp || !r.playedDate) {
                            <span style="color:#bdbdbd">—</span>
                          }
                        </td>
                      </ng-container>
                      <ng-container matColumnDef="opponent">
                        <th mat-header-cell *matHeaderCellDef>Tegenstander</th>
                        <td mat-cell *matCellDef="let r"><strong>{{ r.opponentName }}</strong></td>
                      </ng-container>
                      <ng-container matColumnDef="score">
                        <th mat-header-cell *matHeaderCellDef style="width:80px;text-align:center">Score</th>
                        <td mat-cell *matCellDef="let r" style="text-align:center">
                          @if (r.myScore !== null) {
                            <span>{{ r.myScore }} – {{ r.oppScore }}</span>
                          }
                          @if (r.myScore === null) {
                            <span style="color:#bdbdbd">—</span>
                          }
                        </td>
                      </ng-container>
                      <ng-container matColumnDef="result">
                        <th mat-header-cell *matHeaderCellDef style="width:110px;text-align:center">Uitslag</th>
                        <td mat-cell *matCellDef="let r" style="text-align:center">
                          <span [class.result-W]="r.result === 'W'"
                            [class.result-V]="r.result === 'V'"
                            [class.result-G]="r.result === 'G'"
                            [class.result-af]="r.result === 'Afgemeld'">
                            {{ r.result }}
                          </span>
                          @if (r.reportedBy) {
                            <br>
                            <span style="font-size:11px;color:#9e9e9e">door: {{ r.reportedBy }}</span>
                          }
                        </td>
                      </ng-container>
                      <tr mat-header-row *matHeaderRowDef="matchCols"></tr>
                      <tr mat-row *matRowDef="let row; columns: matchCols;"></tr>
                    </table>
                    @if (playerMatchRows.length === 0) {
                      <p style="color:#9e9e9e;text-align:center;padding:24px 0;margin:0">
                        Geen wedstrijden gevonden voor deze speler.
                      </p>
                    }
                  </mat-card-content>
                </mat-card>
              }
            </div>
          </mat-tab>
          <!-- Tab 6: Speler overzicht -->
          <mat-tab label="Overzicht">
            <div style="padding-top:16px">
              <mat-card>
                <mat-card-content>
                  <table mat-table [dataSource]="playerStatusRows" style="width:100%">
                    <ng-container matColumnDef="nr">
                      <th mat-header-cell *matHeaderCellDef style="width:48px;cursor:pointer;user-select:none"
                          (click)="sortStatus('nr')">
                        Nr
                        @if (statusSortCol() === 'nr') {
                          <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">
                            {{ statusSortDir() === 'asc' ? 'arrow_upward' : 'arrow_downward' }}
                          </mat-icon>
                        }
                      </th>
                      <td mat-cell *matCellDef="let r">{{ r.player.nr }}</td>
                    </ng-container>
                    <ng-container matColumnDef="name">
                      <th mat-header-cell *matHeaderCellDef style="cursor:pointer;user-select:none"
                          (click)="sortStatus('name')">
                        Naam
                        @if (statusSortCol() === 'name') {
                          <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">
                            {{ statusSortDir() === 'asc' ? 'arrow_upward' : 'arrow_downward' }}
                          </mat-icon>
                        }
                      </th>
                      <td mat-cell *matCellDef="let r"><strong>{{ r.player.name }}</strong></td>
                    </ng-container>
                    <ng-container matColumnDef="gespeeld">
                      <th mat-header-cell *matHeaderCellDef
                          style="width:100px;text-align:center;cursor:pointer;user-select:none"
                          (click)="sortStatus('gespeeld')">
                        Gespeeld
                        @if (statusSortCol() === 'gespeeld') {
                          <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">
                            {{ statusSortDir() === 'asc' ? 'arrow_upward' : 'arrow_downward' }}
                          </mat-icon>
                        }
                      </th>
                      <td mat-cell *matCellDef="let r" style="text-align:center;color:#2e7d32;font-weight:600">{{ r.gespeeld }}</td>
                    </ng-container>
                    <ng-container matColumnDef="teSpelen">
                      <th mat-header-cell *matHeaderCellDef
                          style="width:120px;text-align:center;cursor:pointer;user-select:none"
                          (click)="sortStatus('teSpelen')">
                        Nog te spelen
                        @if (statusSortCol() === 'teSpelen') {
                          <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">
                            {{ statusSortDir() === 'asc' ? 'arrow_upward' : 'arrow_downward' }}
                          </mat-icon>
                        }
                      </th>
                      <td mat-cell *matCellDef="let r" style="text-align:center">{{ r.teSpelen }}</td>
                    </ng-container>
                    <ng-container matColumnDef="inTeHalen">
                      <th mat-header-cell *matHeaderCellDef
                          style="width:130px;text-align:center;cursor:pointer;user-select:none"
                          (click)="sortStatus('inTeHalen')">
                        Nog in te halen
                        @if (statusSortCol() === 'inTeHalen') {
                          <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">
                            {{ statusSortDir() === 'asc' ? 'arrow_upward' : 'arrow_downward' }}
                          </mat-icon>
                        }
                      </th>
                      <td mat-cell *matCellDef="let r" style="text-align:center"
                          [style.color]="r.inTeHalen > 0 ? '#c62828' : null"
                          [style.font-weight]="r.inTeHalen > 0 ? '600' : null">
                        {{ r.inTeHalen || '—' }}
                      </td>
                    </ng-container>
                    <tr mat-header-row *matHeaderRowDef="statusCols"></tr>
                    <tr mat-row *matRowDef="let row; columns: statusCols;"></tr>
                  </table>
                </mat-card-content>
              </mat-card>
            </div>
          </mat-tab>
          <!-- Tab 7: Schrijver / Teller -->
          <mat-tab label="Schrijver/Teller">
            <div style="padding-top:16px">
              <mat-form-field class="duty-select" subscriptSizing="dynamic">
                <mat-label>Speler</mat-label>
                <mat-select [value]="selectedDutyPlayerId()" (valueChange)="selectedDutyPlayerId.set($event)">
                  <mat-option value="">— Kies een speler —</mat-option>
                  @for (d of dutyStats(); track d) {
                    <mat-option [value]="d.player.id">
                      {{ d.player.nr }} – {{ d.player.name }}
                    </mat-option>
                  }
                </mat-select>
              </mat-form-field>
              @if (selectedDutyPlayer; as d) {
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
                    @for (row of dutyByEvening; track row) {
                      <tr>
                        <td>{{ row.eveningNr }}</td>
                        <td class="sec-cell">{{ row.sec || '—' }}</td>
                        <td class="cnt-cell">{{ row.cnt || '—' }}</td>
                        <td class="tot-cell">{{ row.total }}</td>
                      </tr>
                    }
                  </tbody>
                </table>
                <div class="duty-section-title">Geschreven wedstrijden</div>
                @if (d.secretaryMatches.length === 0) {
                  <p class="duty-empty">Geen.</p>
                }
                @if (d.secretaryMatches.length > 0) {
                  <table class="duty-table">
                    <thead><tr><th>Avond</th><th>Speler A</th><th></th><th>Speler B</th></tr></thead>
                    <tbody>
                      @for (m of d.secretaryMatches; track m) {
                        <tr>
                          <td>{{ m.eveningNr || '—' }}</td>
                          <td>{{ m.playerANr }} {{ m.playerAName }}</td>
                          <td style="color:#999;text-align:center">vs</td>
                          <td>{{ m.playerBNr }} {{ m.playerBName }}</td>
                        </tr>
                      }
                    </tbody>
                  </table>
                }
                <div class="duty-section-title" style="margin-top:16px">Getelde wedstrijden</div>
                @if (d.counterMatches.length === 0) {
                  <p class="duty-empty">Geen.</p>
                }
                @if (d.counterMatches.length > 0) {
                  <table class="duty-table">
                    <thead><tr><th>Avond</th><th>Speler A</th><th></th><th>Speler B</th></tr></thead>
                    <tbody>
                      @for (m of d.counterMatches; track m) {
                        <tr>
                          <td>{{ m.eveningNr || '—' }}</td>
                          <td>{{ m.playerANr }} {{ m.playerAName }}</td>
                          <td style="color:#999;text-align:center">vs</td>
                          <td>{{ m.playerBNr }} {{ m.playerBName }}</td>
                        </tr>
                      }
                    </tbody>
                  </table>
                }
              }
            </div>
          </mat-tab>
        </mat-tab-group>
      }
    </div>
    `
})
export class InfoComponent implements OnInit {
  private scheduleService = inject(ScheduleService);
  private seasonService   = inject(SeasonService);
  private scoreService    = inject(ScoreService);
  private destroyRef      = inject(DestroyRef);

  info        = signal<ScheduleInfo | null>(null);
  schedule    = signal<Schedule | null>(null);
  playerRows  = signal<PlayerRow[]>([]);
  statRows    = signal<PlayerStats[]>([]);
  selectedPlayerId = signal('');

  dutyStats           = signal<DutyStats[]>([]);
  selectedDutyPlayerId = signal('');

  get selectedDutyPlayer(): DutyStats | null {
    return this.dutyStats().find(d => d.player.id === this.selectedDutyPlayerId()) ?? null;
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

  summaryCols = ['nr', 'name', 'eveningCount', 'totalMatches', 'consecutive', 'buddy'];
  statCols    = ['nr', 'name', 'minTurns', 'avgTurns', 'avgScore', '180s', 'hf'];
  matchCols   = ['evening', 'date', 'playedDate', 'opponent', 'score', 'result'];
  openCols    = ['evening', 'date', 'playerA', 'playerB'];
  statusCols  = ['nr', 'name', 'gespeeld', 'teSpelen', 'inTeHalen'];
  statusSortCol = signal<'nr' | 'name' | 'gespeeld' | 'teSpelen' | 'inTeHalen'>('nr');
  statusSortDir = signal<'asc' | 'desc'>('asc');

  sortStatus(col: 'nr' | 'name' | 'gespeeld' | 'teSpelen' | 'inTeHalen'): void {
    if (this.statusSortCol() === col) {
      this.statusSortDir.set(this.statusSortDir() === 'asc' ? 'desc' : 'asc');
    } else {
      this.statusSortCol.set(col);
      this.statusSortDir.set(col === 'nr' || col === 'name' ? 'asc' : 'desc');
    }
  }

  ngOnInit(): void {
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      distinctUntilChanged(),
      filter(id => !!id),
    ).subscribe(id => this.load(id));
  }

  get sortedPlayers(): PlayerInfoItem[] {
    if (!this.info()) return [];
    return [...this.info()!.players].sort((a, b) => (parseInt(a.nr) || 9999) - (parseInt(b.nr) || 9999));
  }

  get playerMatchRows(): MatchRow[] {
    if (!this.schedule() || !this.selectedPlayerId() || !this.info()) return [];
    const rows: MatchRow[] = [];
    const playerMap = new Map(this.info()!.players.map(p => [p.id, p]));

    for (const ev of this.schedule()!.evenings) {
      if (ev.isInhaalAvond) continue; // catch-up matches are the same objects as in their regular evening

      for (const m of ev.matches) {
        const isA = m.playerA === this.selectedPlayerId();
        const isB = m.playerB === this.selectedPlayerId();
        if (!isA && !isB) continue;

        const opponentId = isA ? m.playerB : m.playerA;
        const opp = playerMap.get(opponentId);
        const opponentName = opp ? `${opp.nr} ${opp.name}` : opponentId.slice(0, 8);
        const selfName = `${playerMap.get(this.selectedPlayerId())?.nr ?? ''} ${playerMap.get(this.selectedPlayerId())?.name ?? ''}`.trim();

        const myScore  = isA ? m.scoreA : m.scoreB;
        const oppScore = isA ? m.scoreB : m.scoreA;

        let result: MatchRow['result'] = '—';
        if (m.reportedBy && !m.played) {
          result = 'Afgemeld';
        } else if (m.played && myScore !== null && oppScore !== null) {
          result = myScore > oppScore ? 'W' : myScore < oppScore ? 'V' : 'G';
        }

        rows.push({
          eveningNumber: ev.number,
          eveningDate: ev.date,
          playerAName: isA ? selfName : opponentName,
          playerBName: isA ? opponentName : selfName,
          opponentName,
          myScore,
          oppScore,
          result,
          reportedBy: result === 'Afgemeld' ? (m.reportedBy ?? '') : '',
          isCatchUp: !!m.playedDate,
          playedDate: m.playedDate ?? '',
        });
      }
    }

    return rows.sort((a, b) => a.eveningNumber - b.eveningNumber);
  }

  get openMatchRows(): { eveningNumber: number; eveningDate: string; playerAName: string; playerBName: string }[] {
    if (!this.schedule() || !this.info()) return [];
    const playerMap = new Map(this.info()!.players.map(p => [p.id, p]));
    const rows: { eveningNumber: number; eveningDate: string; playerAName: string; playerBName: string }[] = [];
    for (const ev of this.schedule()!.evenings) {
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

  get playerStatusRows(): PlayerStatusRow[] {
    if (!this.schedule() || !this.info()) return [];
    const playerMap = new Map(this.info()!.players.map(p => [p.id, p]));
    const counts = new Map<string, { gespeeld: number; teSpelen: number; inTeHalen: number }>();
    for (const p of this.info()!.players) {
      counts.set(p.id, { gespeeld: 0, teSpelen: 0, inTeHalen: 0 });
    }
    for (const ev of this.schedule()!.evenings) {
      if (ev.isInhaalAvond) continue;
      for (const m of ev.matches) {
        for (const pid of [m.playerA, m.playerB]) {
          const c = counts.get(pid);
          if (!c) continue;
          if (m.played) {
            c.gespeeld++;
          } else if (m.reportedBy) {
            c.inTeHalen++;
          } else {
            c.teSpelen++;
          }
        }
      }
    }
    const rows = [...counts.entries()]
      .map(([id, c]) => ({ player: playerMap.get(id)!, ...c }))
      .filter(r => r.player);

    const col = this.statusSortCol();
    const dir = this.statusSortDir() === 'asc' ? 1 : -1;
    rows.sort((a, b) => {
      if (col === 'name') return dir * a.player.name.localeCompare(b.player.name);
      if (col === 'nr')   return dir * ((parseInt(a.player.nr) || 9999) - (parseInt(b.player.nr) || 9999));
      return dir * (a[col] - b[col]);
    });
    return rows;
  }

  private load(scheduleId: string): void {
    forkJoin({
      info:     this.scheduleService.getInfo(scheduleId),
      schedule: this.scheduleService.getById(scheduleId),
      stats:    this.scoreService.getStats(scheduleId),
      duties:   this.scoreService.getDutyStats(scheduleId),
    }).subscribe({
      next: ({ info, schedule, stats, duties }) => {
        this.info.set(info);
        this.schedule.set(schedule);
        this.playerRows.set(this.buildPlayerRows(info));
        this.statRows.set(stats
          .filter(s => s.played > 0)
          .sort((a, b) => (parseInt(a.player.nr || '0')) - (parseInt(b.player.nr || '0'))));
        this.dutyStats.set(duties
          .filter(d => d.count > 0)
          .sort((a, b) => (parseInt(a.player.nr) || 9999) - (parseInt(b.player.nr) || 9999)));
      },
      error: () => { this.info.set(null); this.schedule.set(null); this.playerRows.set([]); this.statRows.set([]); this.dutyStats.set([]); },
    });
  }

  private buildBuddyViolationMap(info: ScheduleInfo, evenings: { id: string; number: number }[]): Map<string, Map<number, 'soft' | 'hard'>> {
    const eveningIndexById = new Map<string, number>(evenings.map((ev, i) => [ev.id, i]));
    const lookup = new Map<string, Map<string, number>>();
    for (const cell of info.matrix) {
      if (!lookup.has(cell.playerId)) lookup.set(cell.playerId, new Map());
      lookup.get(cell.playerId)!.set(cell.eveningId, cell.count);
    }

    // playerId -> eveningIndex -> violation level
    const result = new Map<string, Map<number, 'soft' | 'hard'>>();
    const ensurePlayer = (id: string) => {
      if (!result.has(id)) result.set(id, new Map());
      return result.get(id)!;
    };

    for (const pair of info.buddyPairs) {
      const countsA = lookup.get(pair.playerAId) ?? new Map<string, number>();
      const countsB = lookup.get(pair.playerBId) ?? new Map<string, number>();

      // Find evenings where exactly one of the two has count > 0
      const mismatches: number[] = [];
      for (const [evId, idx] of eveningIndexById) {
        const cA = countsA.get(evId) ?? 0;
        const cB = countsB.get(evId) ?? 0;
        if ((cA > 0) !== (cB > 0)) {
          mismatches.push(idx);
        }
      }
      mismatches.sort((a, b) => a - b);

      for (let m = 0; m < mismatches.length; m++) {
        const idx = mismatches[m];
        const level: 'soft' | 'hard' = m === 0 ? 'soft' : 'hard';
        // Only mark the player that actually plays (count > 0) on that evening
        const evId = evenings[idx]?.id;
        if (!evId) continue;
        if ((countsA.get(evId) ?? 0) > 0) {
          const mapA = ensurePlayer(pair.playerAId);
          // Only escalate, never downgrade
          if (!mapA.has(idx) || (mapA.get(idx) === 'soft' && level === 'hard')) {
            mapA.set(idx, level);
          }
        }
        if ((countsB.get(evId) ?? 0) > 0) {
          const mapB = ensurePlayer(pair.playerBId);
          if (!mapB.has(idx) || (mapB.get(idx) === 'soft' && level === 'hard')) {
            mapB.set(idx, level);
          }
        }
      }
    }

    return result;
  }

  private buildPlayerRows(info: ScheduleInfo): PlayerRow[] {
    const lookup = new Map<string, Map<string, number>>();
    for (const cell of info.matrix) {
      if (!lookup.has(cell.playerId)) lookup.set(cell.playerId, new Map());
      lookup.get(cell.playerId)!.set(cell.eveningId, cell.count);
    }

    const evenings = [...info.evenings].sort((a, b) => a.number - b.number);
    const buddyMap = this.buildBuddyViolationMap(info, evenings);

    return info.players.map(player => {
      const byEvening = lookup.get(player.id) ?? new Map<string, number>();
      const counts    = evenings.map(ev => byEvening.get(ev.id) ?? 0);
      const playerBuddyMap = buddyMap.get(player.id) ?? new Map<number, 'soft' | 'hard'>();

      // Compute consecutive run lengths for each cell
      const runLengths = counts.map((c, i) => {
        if (c === 0) return 0;
        // Find start of this run
        let start = i;
        while (start > 0 && counts[start - 1] > 0) start--;
        return i - start + 1; // 1-based position within the run
      });

      // Compute gap violations: mark the first active evening after a gap > 4 as hard.
      const gapHardIndices = new Set<number>();
      let lastActiveIdx = -1;
      for (let i = 0; i < counts.length; i++) {
        if (counts[i] > 0) {
          if (lastActiveIdx >= 0 && i - lastActiveIdx > 4) {
            gapHardIndices.add(i);
          }
          lastActiveIdx = i;
        }
      }

      const cells: CellData[] = counts.map((count, i) => {
        if (count === 0) return { count, level: 'none' as const, consec: false };

        const runPos     = runLengths[i];
        const consec     = runPos === 2; // exactly 2 consecutive — bold only, no color change
        const isSolo     = count === 1;
        const buddyLevel = playerBuddyMap.get(i);
        const isGapHard  = gapHardIndices.has(i);

        // Hard: count > 4, 3rd+ consecutive evening, buddy hard, or gap violation
        if (count > 4 || runPos >= 3 || buddyLevel === 'hard' || isGapHard) {
          return { count, level: 'hard' as const, consec };
        }

        // Soft: solo evening or first buddy mismatch
        if (isSolo || buddyLevel === 'soft') {
          return { count, level: 'soft' as const, consec };
        }

        return { count, level: 'ok' as const, consec };
      });

      const totalMatches = counts.reduce((s, c) => s + c, 0);
      const eveningCount = counts.filter(c => c > 0).length;
      return { player, cells, totalMatches, eveningCount };
    }).filter(row => row.totalMatches > 0)
      .sort((a, b) => (parseInt(a.player.nr) || 9999) - (parseInt(b.player.nr) || 9999));
  }

  printPendingMatches(): void {
    const player = this.sortedPlayers.find(p => p.id === this.selectedPlayerId());
    if (!player) return;
    const pending = this.playerMatchRows.filter(r => r.result === '—' || r.result === 'Afgemeld');
    const compName = this.schedule()?.competitionName ?? '';
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
    const player = this.sortedPlayers.find(p => p.id === this.selectedPlayerId());
    if (!player) return;
    const compName = this.schedule()?.competitionName ?? 'Dartclub';

    const pad = (n: number) => String(n).padStart(2, '0');
    const toIcsDate = (iso: string) => {
      const d = new Date(iso);
      return `${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}`;
    };
    // Simple UID generator
    const uid = (i: number) => `dart-${this.selectedPlayerId().slice(0, 8)}-${i}@grolzicht`;

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

  selectedPlayerEmail(): string {
    const player = this.sortedPlayers.find(p => p.id === this.selectedPlayerId());
    return player?.email ?? '';
  }

  emailPendingMatches(): void {
    const player = this.sortedPlayers.find(p => p.id === this.selectedPlayerId());
    if (!player?.email) return;

    const pending = this.playerMatchRows.filter(r => r.result === '—');
    const compName = this.schedule()?.competitionName ?? 'Dartcompetitie';
    const playerLabel = `${player.nr} ${player.name}`;

    const subject = encodeURIComponent(`Openstaande wedstrijden – ${playerLabel} – ${compName}`);

    const pad = (n: number) => String(n).padStart(2, '0');
    const fmtDate = (iso: string) => {
      const d = new Date(iso);
      return `${pad(d.getDate())}-${pad(d.getMonth() + 1)}-${d.getFullYear()}`;
    };

    // Group by evening (same date + number)
    const byEvening = new Map<number, typeof pending>();
    for (const r of pending) {
      const group = byEvening.get(r.eveningNumber) ?? [];
      group.push(r);
      byEvening.set(r.eveningNumber, group);
    }
    const evenings = [...byEvening.entries()].sort((a, b) => a[0] - b[0]);

    let body = `Beste ${player.name},\n\n`;
    body += `Hieronder vind je een overzicht van jouw nog te spelen wedstrijden voor ${compName}.\n\n`;

    if (evenings.length === 0) {
      body += 'Je hebt geen openstaande wedstrijden meer.\n';
    } else {
      for (const [, rows] of evenings) {
        const n = rows.length;
        body += `${fmtDate(rows[0].eveningDate)}: ${n} wedstrijd${n !== 1 ? 'en' : ''}\n`;
        for (const r of rows) {
          body += `  \u2022 ${r.playerAName} - ${r.playerBName}\n`;
        }
      }
    }

    body += '\nMet sportieve groet,\nDart Scheduler';

    window.location.href = `mailto:${player.email}?subject=${subject}&body=${encodeURIComponent(body)}`;
  }

  printOpenMatches(): void {
    const rows = this.openMatchRows;
    const name = this.schedule()?.competitionName ?? '';
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
    if (!this.info()) return null;
    const pair = this.info()!.buddyPairs.find(p => p.playerAId === playerId || p.playerBId === playerId);
    if (!pair) return null;
    const isA = pair.playerAId === playerId;
    return {
      partnerNr:   isA ? pair.playerBNr   : pair.playerANr,
      partnerName: isA ? pair.playerBName : pair.playerAName,
      eveningNrs:  pair.eveningNrs,
    };
  }

  getStreaks(row: PlayerRow): number[][] {
    const evenings = this.info() ? [...this.info()!.evenings].sort((a, b) => a.number - b.number) : [];
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
