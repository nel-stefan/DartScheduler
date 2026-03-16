import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { distinctUntilChanged, filter, forkJoin } from 'rxjs';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatTabsModule } from '@angular/material/tabs';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { ScheduleService } from '../../services/schedule.service';
import { SeasonService } from '../../services/season.service';
import { ScoreService } from '../../services/score.service';
import { PlayerInfoItem, EveningInfoItem, ScheduleInfo, BuddyPairItem, PlayerStats, Schedule } from '../../models';

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
            MatTabsModule, MatSelectModule, MatFormFieldModule],
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

        <!-- Tab 4: Wedstrijdoverzicht per speler -->
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

      </mat-tab-group>
    </div>
  `,
})
export class InfoComponent implements OnInit {
  private scheduleService = inject(ScheduleService);
  private seasonService   = inject(SeasonService);
  private scoreService    = inject(ScoreService);
  private destroyRef      = inject(DestroyRef);

  info:     ScheduleInfo | null = null;
  schedule: Schedule | null     = null;
  playerRows: PlayerRow[]       = [];
  statRows:   PlayerStats[]     = [];
  selectedPlayerId              = '';

  summaryCols = ['nr', 'name', 'eveningCount', 'totalMatches', 'consecutive', 'buddy'];
  statCols    = ['nr', 'name', 'minTurns', 'avgTurns', 'avgScore', '180s', 'hf'];
  matchCols   = ['evening', 'date', 'opponent', 'score', 'result'];

  ngOnInit(): void {
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      distinctUntilChanged(),
      filter(id => !!id),
    ).subscribe(id => this.load(id));
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

  private load(scheduleId: string): void {
    forkJoin({
      info:     this.scheduleService.getInfo(scheduleId),
      schedule: this.scheduleService.getById(scheduleId),
      stats:    this.scoreService.getStats(scheduleId),
    }).subscribe({
      next: ({ info, schedule, stats }) => {
        this.info     = info;
        this.schedule = schedule;
        this.playerRows = this.buildPlayerRows(info);
        this.statRows = stats
          .filter(s => s.played > 0)
          .sort((a, b) => (parseInt(a.player.nr || '0')) - (parseInt(b.player.nr || '0')));
      },
      error: () => { this.info = null; this.schedule = null; this.playerRows = []; this.statRows = []; },
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
