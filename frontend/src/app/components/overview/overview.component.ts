import { Component, inject, OnInit, Inject, DestroyRef, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { filter, distinctUntilChanged } from 'rxjs';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialog, MatDialogModule, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatTabsModule } from '@angular/material/tabs';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatIconModule } from '@angular/material/icon';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatRadioModule } from '@angular/material/radio';
import { MatDividerModule } from '@angular/material/divider';
import { MatSelectModule } from '@angular/material/select';
import { ScheduleService } from '../../services/schedule.service';
import { PlayerService } from '../../services/player.service';
import { ScoreService } from '../../services/score.service';
import { SeasonService } from '../../services/season.service';
import { Schedule, Player, Match, Evening } from '../../models';
import { environment } from '../../../environments/environment';
import { EveningStatDialogComponent, EveningStatDialogData } from '../evening-stat-dialog.component';

// ---------------------------------------------------------------------------
// Score-dialog (leg-entry for best of 3)
// ---------------------------------------------------------------------------

export interface ScoreDialogData {
  match: Match;
  nameA: string;
  nameB: string;
  nrA: string;
  nrB: string;
  players: Player[];
  isInhaalAvond: boolean;
  evenings: Evening[];
  defaultPlayedDate: string;
}

@Component({
    selector: 'app-score-dialog',
    imports: [CommonModule, ReactiveFormsModule, MatDialogModule, MatButtonModule,
        MatFormFieldModule, MatInputModule, MatRadioModule, MatSelectModule, MatDividerModule],
    styles: [`
    .matchup { font-size: 13px; text-align: center; margin-bottom: 12px; color: #444; }
    .matchup strong { font-size: 14px; color: #111; }
    .matchup .vs { color: #999; margin: 0 6px; }
    .matchup .bof { font-size: 11px; color: #aaa; margin-left: 6px; }
    .legs-grid {
      display: grid;
      grid-template-columns: 48px 1fr 80px;
      align-items: center;
      gap: 6px 10px;
      margin-bottom: 4px;
      font-size: 13px;
    }
    .leg-label { font-size: 11px; font-weight: 600; color: #666; text-transform: uppercase; letter-spacing: .4px; }
    .radio-group { display: flex; gap: 16px; align-items: center; }
    .radio-group label { display: flex; align-items: center; gap: 5px; font-size: 13px; cursor: pointer; }
    .radio-group input[type=radio] { cursor: pointer; accent-color: #1976d2; }
    .turns-field { width: 80px; }
    .admin-row { display: flex; gap: 8px; flex-wrap: wrap; margin-top: 10px; }
  `],
    template: `
    <h2 mat-dialog-title style="font-size:16px;margin-bottom:4px">Score invoeren</h2>
    <mat-dialog-content style="min-width:400px;padding-top:0">
      <div class="matchup">
        <strong>{{ data.nameA }}</strong>
        <span class="vs">vs</span>
        <strong>{{ data.nameB }}</strong>
        <span class="bof">best of 3</span>
      </div>
      <form [formGroup]="form">
        <div class="legs-grid">
          <!-- headers -->
          <span></span>
          <span style="font-size:11px;color:#999;font-weight:600">Winnaar</span>
          <span style="font-size:11px;color:#999;font-weight:600">Beurten</span>
          <!-- Partij 1 -->
          <span class="leg-label">P1</span>
          <div class="radio-group">
            <label><input type="radio" [value]="data.match.playerA" formControlName="leg1Winner"> {{ data.nameA }}</label>
            <label><input type="radio" [value]="data.match.playerB" formControlName="leg1Winner"> {{ data.nameB }}</label>
          </div>
          <mat-form-field class="turns-field" subscriptSizing="dynamic">
            <mat-select formControlName="leg1Turns">
              <mat-option [value]="null">—</mat-option>
              @for (n of turnsOptions; track n) {
                <mat-option [value]="n">{{ n }}</mat-option>
              }
            </mat-select>
          </mat-form-field>
          <!-- Partij 2 -->
          <span class="leg-label">P2</span>
          <div class="radio-group">
            <label><input type="radio" [value]="data.match.playerA" formControlName="leg2Winner"> {{ data.nameA }}</label>
            <label><input type="radio" [value]="data.match.playerB" formControlName="leg2Winner"> {{ data.nameB }}</label>
          </div>
          <mat-form-field class="turns-field" subscriptSizing="dynamic">
            <mat-select formControlName="leg2Turns">
              <mat-option [value]="null">—</mat-option>
              @for (n of turnsOptions; track n) {
                <mat-option [value]="n">{{ n }}</mat-option>
              }
            </mat-select>
          </mat-form-field>
          <!-- Partij 3 -->
          <span class="leg-label">P3</span>
          <div class="radio-group">
            <label><input type="radio" [value]="data.match.playerA" formControlName="leg3Winner"> {{ data.nameA }}</label>
            <label><input type="radio" [value]="data.match.playerB" formControlName="leg3Winner"> {{ data.nameB }}</label>
          </div>
          <mat-form-field class="turns-field" subscriptSizing="dynamic">
            <mat-select formControlName="leg3Turns">
              <mat-option [value]="null">—</mat-option>
              @for (n of turnsOptions; track n) {
                <mat-option [value]="n">{{ n }}</mat-option>
              }
            </mat-select>
          </mat-form-field>
        </div>
        <!-- Administrative fields -->
        <mat-divider style="margin:10px 0 8px"></mat-divider>
        <div class="admin-row">
          <mat-form-field style="flex:1;min-width:130px" subscriptSizing="dynamic">
            <mat-label>Schrijver</mat-label>
            <mat-select formControlName="secretaryNr">
              <mat-option value="">—</mat-option>
              @for (p of data.players; track p) {
                <mat-option [value]="p.nr">{{ p.nr }} – {{ p.name }}</mat-option>
              }
            </mat-select>
          </mat-form-field>
          <mat-form-field style="flex:1;min-width:130px" subscriptSizing="dynamic">
            <mat-label>Teller</mat-label>
            <mat-select formControlName="counterNr">
              <mat-option value="">—</mat-option>
              @for (p of data.players; track p) {
                <mat-option [value]="p.nr">{{ p.nr }} – {{ p.name }}</mat-option>
              }
            </mat-select>
          </mat-form-field>
        </div>
        @if (data.isInhaalAvond) {
          <div style="margin-top:8px">
            <mat-form-field style="width:100%" subscriptSizing="dynamic">
              <mat-label>Datum gespeeld</mat-label>
              <mat-select formControlName="playedDate">
                <mat-option value="">— kies avond —</mat-option>
                @for (ev of data.evenings; track ev) {
                  <mat-option [value]="ev.date">
                    {{ ev.isInhaalAvond ? 'Inhaalavond' : 'Avond ' + ev.number }} — {{ ev.date | date:'d MMM yyyy' }}
                  </mat-option>
                }
              </mat-select>
            </mat-form-field>
          </div>
        }
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="dialogRef.close()">Annuleren</button>
      <button mat-raised-button color="primary" [disabled]="!isValid()" (click)="submit()">Opslaan</button>
    </mat-dialog-actions>
    `
})
export class ScoreDialogComponent {
  dialogRef = inject(MatDialogRef<ScoreDialogComponent>);
  fb = inject(FormBuilder);

  constructor(@Inject(MAT_DIALOG_DATA) public data: ScoreDialogData) {}

  turnsOptions = Array.from({ length: 18 }, (_, i) => i + 3); // 3..20

  form = this.fb.group({
    leg1Winner: [this.data.match.leg1Winner || '', Validators.required],
    leg1Turns:  [this.data.match.leg1Turns  || null as number | null],
    leg2Winner: [this.data.match.leg2Winner || '', Validators.required],
    leg2Turns:  [this.data.match.leg2Turns  || null as number | null],
    leg3Winner: [this.data.match.leg3Winner || ''],
    leg3Turns:  [this.data.match.leg3Turns  || null as number | null],
    secretaryNr: [this.data.match.secretaryNr || ''],
    counterNr:   [this.data.match.counterNr   || ''],
    playedDate:  [this.data.match.playedDate || this.data.defaultPlayedDate || ''],
  });

  isValid(): boolean {
    const v = this.form.value;
    return !!(v.leg1Winner && v.leg2Winner && v.leg3Winner);
  }

  submit(): void {
    if (!this.isValid()) return;
    const v = this.form.value;
    this.dialogRef.close({
      leg1Winner:     v.leg1Winner ?? '',
      leg1Turns:      v.leg1Turns ?? 0,
      leg2Winner:     v.leg2Winner ?? '',
      leg2Turns:      v.leg2Turns ?? 0,
      leg3Winner:     v.leg3Winner ?? '',
      leg3Turns:      v.leg3Turns ?? 0,
      playerA180s:          0,
      playerB180s:          0,
      playerAHighestFinish: 0,
      playerBHighestFinish: 0,
      reportedBy:     '',
      rescheduleDate: '',
      secretaryNr:    v.secretaryNr ?? '',
      counterNr:      v.counterNr ?? '',
      playedDate:     v.playedDate ?? '',
    });
  }
}

// ---------------------------------------------------------------------------
// AbsentDialog — report one player absent for an evening
// ---------------------------------------------------------------------------

interface AbsentDialogData { evening: Evening; players: Player[]; }

@Component({
    selector: 'app-absent-dialog',
    imports: [CommonModule, FormsModule, MatDialogModule, MatButtonModule, MatFormFieldModule,
        MatSelectModule, MatInputModule, MatIconModule],
    template: `
    <h2 mat-dialog-title>Speler afmelden — avond {{ data.evening.number }}</h2>
    <mat-dialog-content style="min-width:420px;padding-top:8px">
    
      <mat-form-field style="width:100%">
        <mat-label>Afwezige speler</mat-label>
        <mat-select [(ngModel)]="selectedPlayerId" (ngModelChange)="onPlayerChange()">
          @for (p of selectablePlayers; track p) {
            <mat-option [value]="p.id">
              {{ p.nr }} — {{ p.name }}
            </mat-option>
          }
        </mat-select>
      </mat-form-field>
    
      <mat-form-field style="width:100%;margin-top:4px">
        <mat-label>Afgemeld door</mat-label>
        <input matInput [(ngModel)]="reportedBy" placeholder="naam of nr">
      </mat-form-field>
    
      @if (selectedPlayerId) {
        <div style="margin-top:12px">
          <div style="font-size:13px;font-weight:500;margin-bottom:6px;color:#333">
            Wedstrijden die als afgemeld worden gemarkeerd:
          </div>
          @if (affectedMatches.length === 0) {
            <div style="color:#9e9e9e;font-size:13px">
              Geen openstaande wedstrijden gevonden.
            </div>
          }
          @for (m of affectedMatches; track m) {
            <div
              style="display:flex;align-items:center;gap:8px;padding:4px 8px;background:#fff3e0;border-radius:4px;margin-bottom:4px;font-size:13px">
              <mat-icon style="font-size:16px;width:16px;height:16px;color:#e65100">warning</mat-icon>
              <strong>{{ playerLabel(m.playerA) }}</strong>
              <span style="color:#999">vs</span>
              <strong>{{ playerLabel(m.playerB) }}</strong>
            </div>
          }
        </div>
      }
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annuleren</button>
      <button mat-raised-button color="warn"
        [disabled]="!selectedPlayerId || affectedMatches.length === 0"
        (click)="confirm()">
        <mat-icon>person_off</mat-icon> Afmelden
      </button>
    </mat-dialog-actions>
    `
})
export class AbsentDialogComponent {
  data      = inject<AbsentDialogData>(MAT_DIALOG_DATA);
  dialogRef = inject(MatDialogRef<AbsentDialogComponent>);

  selectedPlayerId = '';
  reportedBy = '';

  get selectablePlayers(): Player[] {
    const ids = new Set<string>();
    for (const m of this.data.evening.matches ?? []) {
      if (!m.played) { ids.add(m.playerA); ids.add(m.playerB); }
    }
    return this.data.players
      .filter(p => ids.has(p.id))
      .sort((a, b) => parseInt(a.nr || '0') - parseInt(b.nr || '0'));
  }

  get affectedMatches(): Match[] {
    if (!this.selectedPlayerId) return [];
    return (this.data.evening.matches ?? []).filter(
      m => !m.played && (m.playerA === this.selectedPlayerId || m.playerB === this.selectedPlayerId)
    );
  }

  playerLabel(id: string): string {
    const p = this.data.players.find(p => p.id === id);
    return p ? `${p.nr} ${p.name}` : id.slice(0, 6);
  }

  onPlayerChange(): void {
    const p = this.data.players.find(p => p.id === this.selectedPlayerId);
    if (p) this.reportedBy = p.nr ? `${p.nr} ${p.name}` : p.name;
  }

  confirm(): void {
    this.dialogRef.close({ playerId: this.selectedPlayerId, reportedBy: this.reportedBy });
  }
}

// ---------------------------------------------------------------------------
// Overview-component
// ---------------------------------------------------------------------------

@Component({
    selector: 'app-overview',
    imports: [
        CommonModule,
        MatButtonModule, MatCardModule, MatTableModule, MatTabsModule,
        MatSnackBarModule, MatDialogModule, MatIconModule,
        MatTooltipModule, MatSelectModule, MatFormFieldModule, MatInputModule,
        FormsModule, ReactiveFormsModule,
    ],
    styles: [`
    .schedule-header {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 20px;
      flex-wrap: wrap;
    }
    .schedule-title { margin: 0; font-size: 22px; font-weight: 500; }
    table { width: 100%; }
    :host ::ng-deep .mat-mdc-tab-header { display: none; }
    .score-cell { font-weight: 600; color: #2e7d32; }
    .vs-cell { color: #9e9e9e; width: 36px; text-align: center; }
    .empty-state {
      padding: 48px 0;
      text-align: center;
      color: #757575;
    }
    .empty-state mat-icon { font-size: 64px; width: 64px; height: 64px; color: #bdbdbd; }
    .card-header-row { display: flex; align-items: flex-start; justify-content: space-between; }

    .print-only { display: none; }
    .print-schedule-table {
      width: 100%;
      border-collapse: collapse;
      border: 1px solid #bbb;
      font-size: 8pt;
      table-layout: fixed;
    }
    .print-schedule-table th, .print-schedule-table td {
      border: 1px solid #bbb;
      padding: 2px 4px;
      text-align: center;
      white-space: nowrap;
      overflow: hidden;
    }
    .print-schedule-table th { background: #f0f0f0; font-weight: 600; }
    .print-schedule-table td.row-nr { width: 24px; color: #666; font-size: 7pt; }
    .print-schedule-title { font-size: 11pt; font-weight: 600; margin: 0 0 4px 0; }

    @media print {
      @page { size: A4 landscape; margin: 8mm; }
      .schedule-header { display: none !important; }
      mat-tab-group  { display: none !important; }
      .print-only    { display: block !important; }
    }
  `],
    template: `
    <div class="schedule-header">
      <h2 class="schedule-title">{{ schedule()?.competitionName ?? 'DartScheduler' }}</h2>

      @if (schedule()) {
        <button mat-stroked-button (click)="printSchedule()">
          <mat-icon>print</mat-icon> Afdrukken
        </button>
      }
      @if (schedule()) {
        <mat-form-field style="min-width:320px" subscriptSizing="dynamic">
          <mat-label>Avond</mat-label>
          <mat-select [ngModel]="activeTab()" (ngModelChange)="activeTab.set($event)">
            <mat-select-trigger>
              @if (schedule()!.evenings[activeTab()]; as ev) {
                <span [style.background]="eveningColor(ev)"
                style="display:inline-block;width:8px;height:8px;border-radius:50%;margin-right:7px;vertical-align:middle;flex-shrink:0"></span>
                {{ ev.isInhaalAvond ? 'Inhaalavond' : 'Avond ' + ev.number }}
                <span style="color:#9e9e9e;margin-left:4px">— {{ ev.date | date:'d MMM' }}</span>
              }
            </mat-select-trigger>
            @for (ev of schedule()!.evenings; track ev; let i = $index) {
              <mat-option [value]="i">
                <span style="display:flex;align-items:center;gap:8px;width:100%">
                  <span [style.background]="eveningColor(ev)"
                  style="width:8px;height:8px;border-radius:50%;flex-shrink:0;display:inline-block"></span>
                  <span style="flex:1">{{ ev.isInhaalAvond ? 'Inhaalavond' : 'Avond ' + ev.number }}</span>
                  <span style="color:#9e9e9e;font-size:12px">{{ ev.date | date:'d MMM' }}</span>
                  <span [style.color]="eveningColor(ev)" style="font-size:12px;min-width:32px;text-align:right;font-weight:500">
                    {{ eveningCountLabel(ev) }}
                  </span>
                </span>
              </mat-option>
            }
          </mat-select>
        </mat-form-field>
      }
    </div>
    
    @if (!schedule()) {
      <div class="empty-state">
        <mat-icon>sports_bar</mat-icon>
        <p>Nog geen schema. Ga naar <strong>Beheer</strong> om spelers te importeren en een schema te genereren.</p>
      </div>
    }
    
    @if (schedule()) {
      <mat-tab-group animationDuration="150ms" [selectedIndex]="activeTab()"
        (selectedIndexChange)="activeTab.set($event)">
        @for (ev of schedule()!.evenings; track ev) {
          <mat-tab>
            <!-- Tab-inhoud -->
            <mat-card style="border-radius:8px;">
              <mat-card-header style="padding-bottom:0">
                <div class="card-header-row" style="width:100%">
                  <div>
                    <mat-card-title>
                      @if (ev.isInhaalAvond) {
                        <span style="color:#7b1fa2;margin-right:6px">
                          <mat-icon style="vertical-align:middle;font-size:18px">replay</mat-icon>
                          Inhaalavond
                        </span>
                      }
                      @if (!ev.isInhaalAvond) {
                        <span>Avond {{ ev.number }}</span>
                      }
                      &mdash; {{ ev.date | date:'EEEE d MMMM yyyy' }}
                    </mat-card-title>
                    @if (!ev.isInhaalAvond) {
                      <mat-card-subtitle>
                        {{ playedCount(ev) }} / {{ ev.matches.length }} wedstrijden gespeeld
                      </mat-card-subtitle>
                    }
                    @if (ev.isInhaalAvond) {
                      <mat-card-subtitle style="color:#7b1fa2">
                        {{ openCount(ev) }} openstaande wedstrijden
                      </mat-card-subtitle>
                    }
                  </div>
                  <div style="display:flex;gap:4px;align-items:center">
                    @if (ev.matches.length > 0) {
                      <button mat-stroked-button (click)="openAbsentDialog(ev)" matTooltip="Speler afmelden"
                        >
                        <mat-icon>person_off</mat-icon> Afmelden
                      </button>
                    }
                    @if (ev.matches.length > 0) {
                      <button mat-stroked-button (click)="openStatDialog(ev)" matTooltip="180s / Hoge Finish invoeren"
                        >
                        <mat-icon>emoji_events</mat-icon> 180 / HF
                      </button>
                    }
                    <button mat-icon-button (click)="exportEvening(ev.id)" matTooltip="Exporteren naar Excel">
                      <mat-icon>file_download</mat-icon>
                    </button>
                    <button mat-icon-button (click)="exportEveningPdf(ev.id)" matTooltip="Exporteren naar PDF">
                      <mat-icon>picture_as_pdf</mat-icon>
                    </button>
                    <button mat-icon-button (click)="printEvening(ev.id)" matTooltip="Afdrukken">
                      <mat-icon>print</mat-icon>
                    </button>
                    @if (ev.isInhaalAvond) {
                      <button mat-icon-button color="warn" (click)="deleteEvening(ev)" matTooltip="Avond verwijderen"
                        >
                        <mat-icon>delete</mat-icon>
                      </button>
                    }
                  </div>
                </div>
              </mat-card-header>
              <mat-card-content>
                @if (ev.isInhaalAvond && ev.matches.length === 0) {
                  <p
                    style="color:#757575;text-align:center;padding:24px 0;margin:0">
                    Geen uitgestelde wedstrijden gevonden voor deze inhaalavond.
                  </p>
                }
                @if (ev.isInhaalAvond && ev.matches.length > 0) {
                  <mat-form-field
                    style="width:100%;margin-bottom:8px" subscriptSizing="dynamic">
                    <mat-label>Zoek op nr (bijv. 12 of 12 34)</mat-label>
                    <input matInput [ngModel]="catchUpSearch()" (ngModelChange)="catchUpSearch.set($event)" placeholder="nr A   nr B">
                    @if (catchUpSearch()) {
                      <button matSuffix mat-icon-button (click)="catchUpSearch.set('')">
                        <mat-icon>close</mat-icon>
                      </button>
                    }
                    <mat-icon matPrefix style="margin-right:4px">search</mat-icon>
                  </mat-form-field>
                }
                @if (ev.matches.length > 0) {
                  <table mat-table [dataSource]="ev.isInhaalAvond ? filteredMatches(ev) : ev.matches">
                    <ng-container matColumnDef="playerA">
                      <th mat-header-cell *matHeaderCellDef>Speler A</th>
                      <td mat-cell *matCellDef="let m">{{ playerName(m.playerA) }}</td>
                    </ng-container>
                    <ng-container matColumnDef="vs">
                      <th mat-header-cell *matHeaderCellDef></th>
                      <td mat-cell *matCellDef="let m" class="vs-cell">vs</td>
                    </ng-container>
                    <ng-container matColumnDef="playerB">
                      <th mat-header-cell *matHeaderCellDef>Speler B</th>
                      <td mat-cell *matCellDef="let m">{{ playerName(m.playerB) }}</td>
                    </ng-container>
                    <ng-container matColumnDef="score">
                      <th mat-header-cell *matHeaderCellDef>Score</th>
                      <td mat-cell *matCellDef="let m" class="score-cell">
                        @if (m.played) {
                          <span>{{ m.scoreA }} – {{ m.scoreB }}</span>
                        }
                        @if (!m.played && m.reportedBy) {
                          <span
                            style="color:#e65100;font-size:12px;font-weight:500">
                            <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">schedule</mat-icon>
                            Afgemeld: {{ m.reportedBy }}
                          </span>
                        }
                        @if (!m.played && !m.reportedBy) {
                          <span style="color:#bdbdbd">—</span>
                        }
                      </td>
                    </ng-container>
                    <ng-container matColumnDef="actions">
                      <th mat-header-cell *matHeaderCellDef></th>
                      <td mat-cell *matCellDef="let m" style="text-align:right">
                        @if (!m.played) {
                          <button mat-stroked-button color="primary"
                            (click)="openScore(m, ev.isInhaalAvond)">
                            <mat-icon>edit</mat-icon> Score
                          </button>
                        }
                        @if (m.played) {
                          <button mat-button color="accent"
                            (click)="openScore(m, ev.isInhaalAvond)" matTooltip="Score wijzigen">
                            <mat-icon>check_circle</mat-icon> Wijzigen
                          </button>
                        }
                      </td>
                    </ng-container>
                    <tr mat-header-row *matHeaderRowDef="matchCols"></tr>
                    <tr mat-row *matRowDef="let row; columns: matchCols;"
                    [class.match-played]="row.played"></tr>
                  </table>
                }
              </mat-card-content>
            </mat-card>
          </mat-tab>
        }
      </mat-tab-group>
    }
    
    <!-- Print-only schedule matrix -->
    @if (schedule() && printData) {
      <div class="print-only">
        <p class="print-schedule-title">{{ schedule()!.competitionName }} — {{ schedule()!.season }}</p>
        <!-- Page 1: first half of evenings -->
        <table class="print-schedule-table">
          <thead>
            <tr>
              <th class="row-nr">#</th>
              @for (ev of printData.half1; track ev) {
                <th>
                  {{ ev.number }}<br><span style="font-weight:400">{{ ev.date | date:'d MMM' }}</span>
                </th>
              }
            </tr>
          </thead>
          <tbody>
            @for (row of printData.rows1; track row; let i = $index) {
              <tr>
                <td class="row-nr">{{ i + 1 }}</td>
                @for (col of printData.half1; track col; let ci = $index) {
                  @if (col.isCatchUp && i === 0) {
                    <td
                      [attr.rowspan]="printData.rowCount"
                      style="writing-mode:vertical-lr;text-align:center;vertical-align:middle;font-weight:600;letter-spacing:3px;color:#7b1fa2;font-size:7pt">
                      INHAALAVOND
                    </td>
                  }
                  @if (!col.isCatchUp) {
                    <td>{{ row[ci] }}</td>
                  }
                }
              </tr>
            }
          </tbody>
        </table>
        <!-- Page 2: second half of evenings -->
        <div style="page-break-before:always"></div>
        <p class="print-schedule-title">{{ schedule()!.competitionName }} — {{ schedule()!.season }}</p>
        <table class="print-schedule-table">
          <thead>
            <tr>
              <th class="row-nr">#</th>
              @for (ev of printData.half2; track ev) {
                <th>
                  {{ ev.number }}<br><span style="font-weight:400">{{ ev.date | date:'d MMM' }}</span>
                </th>
              }
            </tr>
          </thead>
          <tbody>
            @for (row of printData.rows2; track row; let i = $index) {
              <tr>
                <td class="row-nr">{{ i + 1 }}</td>
                @for (col of printData.half2; track col; let ci = $index) {
                  @if (col.isCatchUp && i === 0) {
                    <td
                      [attr.rowspan]="printData.rowCount"
                      style="writing-mode:vertical-lr;text-align:center;vertical-align:middle;font-weight:600;letter-spacing:3px;color:#7b1fa2;font-size:7pt">
                      INHAALAVOND
                    </td>
                  }
                  @if (!col.isCatchUp) {
                    <td>{{ row[ci] }}</td>
                  }
                }
              </tr>
            }
          </tbody>
        </table>
      </div>
    }
    `
})
export class OverviewComponent implements OnInit {
  private scheduleService  = inject(ScheduleService);
  private playerService    = inject(PlayerService);
  private scoreService     = inject(ScoreService);
  private seasonService    = inject(SeasonService);
  private snackBar         = inject(MatSnackBar);
  private dialog           = inject(MatDialog);
  private destroyRef       = inject(DestroyRef);

  schedule = signal<Schedule | null>(null);
  players  = signal<Player[]>([]);
  activeTab = signal(0);
  matchCols = ['playerA', 'vs', 'playerB', 'score', 'actions'];
  catchUpSearch = signal('');
  lastCatchUpPlayedDate = signal('');

  ngOnInit(): void {
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      filter(id => !!id),
      distinctUntilChanged(),
    ).subscribe(id => {
      if (this.schedule()?.id !== id) this.loadScheduleById(id);
    });
    this.playerService.list().subscribe({ next: (ps) => (this.players.set(ps)), error: () => {} });
  }

  loadScheduleById(id: string, preserveTab = false): void {
    this.scheduleService.getById(id).subscribe({
      next: (s) => {
        this.schedule.set(s);
        if (!preserveTab) this.activeTab.set(this.firstUpcomingTab(s.evenings));
      },
      error: () => {},
    });
  }

  private firstUpcomingTab(evenings: Evening[]): number {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const idx = evenings.findIndex(ev => new Date(ev.date) >= today);
    return idx >= 0 ? idx : evenings.length - 1;
  }

  playerName(id: string): string {
    const p = this.players().find((p) => p.id === id);
    if (!p) return id.slice(0, 8);
    return p.nr ? `${p.nr} ${p.name}` : p.name;
  }

  playerNr(id: string): string {
    return this.players().find((p) => p.id === id)?.nr ?? '';
  }

  filteredMatches(ev: Evening): Match[] {
    const q = this.catchUpSearch().trim().toLowerCase();
    if (!q) return ev.matches;
    const tokens = q.split(/\s+/);
    return ev.matches.filter(m => {
      const nrA = this.playerNr(m.playerA).toLowerCase();
      const nrB = this.playerNr(m.playerB).toLowerCase();
      return tokens.every(t => nrA.includes(t) || nrB.includes(t));
    });
  }

  playedCount(ev: Evening): number {
    return ev.matches?.filter(m => m.played).length ?? 0;
  }

  openCount(ev: Evening):  number  { return (ev.matches ?? []).filter(m => !m.played).length; }

  eveningColor(ev: Evening): string {
    if (ev.isInhaalAvond) return '#7b1fa2';
    const played = this.playedCount(ev);
    const total  = ev.matches?.length ?? 0;
    if (total === 0 || played === 0) return '#9e9e9e';
    if (played === total) return '#2e7d32';
    return '#f57c00';
  }

  eveningCountLabel(ev: Evening): string {
    if (ev.isInhaalAvond) return `${this.openCount(ev)} open`;
    const total = ev.matches?.length ?? 0;
    return `${this.playedCount(ev)}/${total}`;
  }

  openScore(match: Match, isInhaalAvond = false): void {
    const ref = this.dialog.open(ScoreDialogComponent, {
      data: {
        match,
        nameA: this.playerName(match.playerA),
        nameB: this.playerName(match.playerB),
        nrA:   this.playerNr(match.playerA),
        nrB:   this.playerNr(match.playerB),
        players: this.players(),
        isInhaalAvond,
        evenings: this.schedule()?.evenings ?? [],
        defaultPlayedDate: this.lastCatchUpPlayedDate(),
      } as ScoreDialogData,
    });
    ref.afterClosed().subscribe((result: {
      leg1Winner: string; leg1Turns: number;
      leg2Winner: string; leg2Turns: number;
      leg3Winner: string; leg3Turns: number;
      playerA180s: number; playerB180s: number;
      playerAHighestFinish: number; playerBHighestFinish: number;
      reportedBy: string; rescheduleDate: string;
      secretaryNr: string; counterNr: string;
      playedDate: string;
    } | undefined) => {
      if (!result) return;
      this.scoreService.submitResult(match.id, { ...result, playedDate: result.playedDate ?? '' }).subscribe({
        next: () => {
          if (isInhaalAvond && result.playedDate) this.lastCatchUpPlayedDate.set(result.playedDate);
          this.snackBar.open('Resultaat opgeslagen!', 'OK', { duration: 2000 });
          if (this.schedule()) this.loadScheduleById(this.schedule()!.id, true);
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    });
  }

  get printData(): {
    half1: { number: number; date: string; isCatchUp: boolean }[]; rows1: string[][];
    half2: { number: number; date: string; isCatchUp: boolean }[]; rows2: string[][];
    rowCount: number;
  } | null {
    if (!this.schedule()) return null;
    const evs = this.schedule()!.evenings;
    const mid = Math.ceil(evs.length / 2);
    const maxRows = Math.max(0, ...evs
      .filter(e => !e.isInhaalAvond)
      .map(ev => ev.matches?.length ?? 0));
    const buildRows = (cols: Evening[]): string[][] =>
      Array.from({ length: maxRows }, (_, ri) =>
        cols.map(ev => {
          if (ev.isInhaalAvond) return '';
          const m = ev.matches?.[ri];
          if (!m) return '';
          const a = this.playerNr(m.playerA);
          const b = this.playerNr(m.playerB);
          return a && b ? `${a} - ${b}` : '';
        })
      );
    const toCol = (ev: Evening) => ({ number: ev.number, date: ev.date, isCatchUp: ev.isInhaalAvond });
    return {
      half1: evs.slice(0, mid).map(toCol),
      rows1: buildRows(evs.slice(0, mid)),
      half2: evs.slice(mid).map(toCol),
      rows2: buildRows(evs.slice(mid)),
      rowCount: maxRows,
    };
  }

  printSchedule(): void {
    window.print();
  }

  openAbsentDialog(ev: Evening): void {
    const ref = this.dialog.open(AbsentDialogComponent, {
      data: { evening: ev, players: this.players() } as AbsentDialogData,
      minWidth: '420px',
    });
    ref.afterClosed().subscribe((result: { playerId: string; reportedBy: string } | undefined) => {
      if (!result) return;
      this.scoreService.reportAbsent(ev.id, result.playerId, result.reportedBy).subscribe({
        next: () => {
          this.snackBar.open('Speler afgemeld', 'OK', { duration: 2000 });
          if (this.schedule()) this.loadScheduleById(this.schedule()!.id, true);
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    });
  }

  openStatDialog(ev: Evening): void {
    const playerIdsOnEvening = new Set(ev.matches.flatMap(m => [m.playerA, m.playerB]));
    const players = this.players()
      .filter(p => playerIdsOnEvening.has(p.id))
      .map(p => ({ id: p.id, name: p.name }));
    this.dialog.open(EveningStatDialogComponent, {
      data: {
        scheduleId: this.schedule()!.id,
        players,
      } as EveningStatDialogData,
    }).afterClosed().subscribe(saved => {
      if (saved) this.snackBar.open('Opgeslagen', '', { duration: 2000 });
    });
  }

  exportEvening(eveningId: string): void {
    window.open(`${environment.apiBaseUrl}/export/evening/${eveningId}/excel`, '_blank');
  }

  exportEveningPdf(eveningId: string): void {
    window.open(`${environment.apiBaseUrl}/export/evening/${eveningId}/pdf`, '_blank');
  }

  printEvening(eveningId: string): void {
    window.open(`${environment.apiBaseUrl}/export/evening/${eveningId}/print`, '_blank');
  }

  deleteEvening(ev: Evening): void {
    if (!confirm(`Inhaalavond ${ev.number} verwijderen?`)) return;
    this.scheduleService.deleteEvening(this.schedule()!.id, ev.id).subscribe({
      next: () => {
        this.snackBar.open('Avond verwijderd', 'OK', { duration: 2000 });
        this.loadScheduleById(this.schedule()!.id);
      },
      error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
    });
  }

}
