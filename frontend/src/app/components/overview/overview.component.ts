import { Component, inject, OnInit, Inject, DestroyRef } from '@angular/core';
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
}

@Component({
  selector: 'app-score-dialog',
  standalone: true,
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
              <mat-option *ngFor="let n of turnsOptions" [value]="n">{{ n }}</mat-option>
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
              <mat-option *ngFor="let n of turnsOptions" [value]="n">{{ n }}</mat-option>
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
              <mat-option *ngFor="let n of turnsOptions" [value]="n">{{ n }}</mat-option>
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
              <mat-option *ngFor="let p of data.players" [value]="p.nr">{{ p.nr }} – {{ p.name }}</mat-option>
            </mat-select>
          </mat-form-field>
          <mat-form-field style="flex:1;min-width:130px" subscriptSizing="dynamic">
            <mat-label>Teller</mat-label>
            <mat-select formControlName="counterNr">
              <mat-option value="">—</mat-option>
              <mat-option *ngFor="let p of data.players" [value]="p.nr">{{ p.nr }} – {{ p.name }}</mat-option>
            </mat-select>
          </mat-form-field>
        </div>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="dialogRef.close()">Annuleren</button>
      <button mat-raised-button color="primary" [disabled]="!isValid()" (click)="submit()">Opslaan</button>
    </mat-dialog-actions>
  `,
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
    });
  }
}

// ---------------------------------------------------------------------------
// AbsentDialog — report one player absent for an evening
// ---------------------------------------------------------------------------

interface AbsentDialogData { evening: Evening; players: Player[]; }

@Component({
  selector: 'app-absent-dialog',
  standalone: true,
  imports: [CommonModule, FormsModule, MatDialogModule, MatButtonModule, MatFormFieldModule,
            MatSelectModule, MatInputModule, MatIconModule],
  template: `
    <h2 mat-dialog-title>Speler afmelden — avond {{ data.evening.number }}</h2>
    <mat-dialog-content style="min-width:420px;padding-top:8px">

      <mat-form-field style="width:100%">
        <mat-label>Afwezige speler</mat-label>
        <mat-select [(ngModel)]="selectedPlayerId" (ngModelChange)="onPlayerChange()">
          <mat-option *ngFor="let p of selectablePlayers" [value]="p.id">
            {{ p.nr }} — {{ p.name }}
          </mat-option>
        </mat-select>
      </mat-form-field>

      <mat-form-field style="width:100%;margin-top:4px">
        <mat-label>Afgemeld door</mat-label>
        <input matInput [(ngModel)]="reportedBy" placeholder="naam of nr">
      </mat-form-field>

      <div *ngIf="selectedPlayerId" style="margin-top:12px">
        <div style="font-size:13px;font-weight:500;margin-bottom:6px;color:#333">
          Wedstrijden die als afgemeld worden gemarkeerd:
        </div>
        <div *ngIf="affectedMatches.length === 0" style="color:#9e9e9e;font-size:13px">
          Geen openstaande wedstrijden gevonden.
        </div>
        <div *ngFor="let m of affectedMatches"
             style="display:flex;align-items:center;gap:8px;padding:4px 8px;background:#fff3e0;border-radius:4px;margin-bottom:4px;font-size:13px">
          <mat-icon style="font-size:16px;width:16px;height:16px;color:#e65100">warning</mat-icon>
          <strong>{{ playerLabel(m.playerA) }}</strong>
          <span style="color:#999">vs</span>
          <strong>{{ playerLabel(m.playerB) }}</strong>
        </div>
      </div>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annuleren</button>
      <button mat-raised-button color="warn"
              [disabled]="!selectedPlayerId || affectedMatches.length === 0"
              (click)="confirm()">
        <mat-icon>person_off</mat-icon> Afmelden
      </button>
    </mat-dialog-actions>
  `,
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
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule, MatCardModule, MatTableModule, MatTabsModule,
    MatSnackBarModule, MatDialogModule, MatIconModule,
    MatTooltipModule, MatSelectModule, MatFormFieldModule, MatInputModule,
    FormsModule, ReactiveFormsModule,
    AbsentDialogComponent, EveningStatDialogComponent,
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
      <h2 class="schedule-title">{{ schedule?.competitionName ?? 'DartScheduler' }}</h2>

      <button mat-stroked-button *ngIf="schedule" (click)="printSchedule()">
        <mat-icon>print</mat-icon> Afdrukken
      </button>
      <mat-form-field *ngIf="schedule" style="min-width:200px" subscriptSizing="dynamic">
        <mat-label>Ga naar avond</mat-label>
        <mat-select [(ngModel)]="activeTab">
          <mat-option *ngFor="let ev of schedule.evenings; let i = index" [value]="i">
            {{ ev.isInhaalAvond ? 'Inhaalavond' : 'Avond ' + ev.number }} — {{ ev.date | date:'d MMM' }}
          </mat-option>
        </mat-select>
      </mat-form-field>
    </div>

    <div *ngIf="!schedule" class="empty-state">
      <mat-icon>sports_bar</mat-icon>
      <p>Nog geen schema. Ga naar <strong>Beheer</strong> om spelers te importeren en een schema te genereren.</p>
    </div>

    <mat-tab-group *ngIf="schedule" animationDuration="150ms" [selectedIndex]="activeTab"
                   (selectedIndexChange)="activeTab = $event">
      <mat-tab *ngFor="let ev of schedule.evenings">

        <!-- Tab-inhoud -->
        <mat-card style="border-radius:8px;">
          <mat-card-header style="padding-bottom:0">
            <div class="card-header-row" style="width:100%">
              <div>
                <mat-card-title>
                  <span *ngIf="ev.isInhaalAvond" style="color:#7b1fa2;margin-right:6px">
                    <mat-icon style="vertical-align:middle;font-size:18px">replay</mat-icon>
                    Inhaalavond
                  </span>
                  <span *ngIf="!ev.isInhaalAvond">Avond {{ ev.number }}</span>
                  &mdash; {{ ev.date | date:'EEEE d MMMM yyyy' }}
                </mat-card-title>
                <mat-card-subtitle *ngIf="!ev.isInhaalAvond">
                  {{ playedCount(ev) }} / {{ ev.matches.length }} wedstrijden gespeeld
                </mat-card-subtitle>
                <mat-card-subtitle *ngIf="ev.isInhaalAvond" style="color:#7b1fa2">
                  {{ openCount(ev) }} openstaande wedstrijden
                </mat-card-subtitle>
              </div>
              <div style="display:flex;gap:4px;align-items:center">
                <button mat-stroked-button (click)="openAbsentDialog(ev)" matTooltip="Speler afmelden"
                        *ngIf="ev.matches.length > 0">
                  <mat-icon>person_off</mat-icon> Afmelden
                </button>
                <button mat-stroked-button (click)="openStatDialog(ev)" matTooltip="180s / Hoge Finish invoeren"
                        *ngIf="ev.matches.length > 0">
                  <mat-icon>emoji_events</mat-icon> 180 / HF
                </button>
                <button mat-icon-button (click)="exportEvening(ev.id)" matTooltip="Exporteren naar Excel">
                  <mat-icon>file_download</mat-icon>
                </button>
                <button mat-icon-button (click)="printEvening(ev.id)" matTooltip="Afdrukken">
                  <mat-icon>print</mat-icon>
                </button>
                <button mat-icon-button color="warn" (click)="deleteEvening(ev)" matTooltip="Avond verwijderen"
                        *ngIf="ev.isInhaalAvond">
                  <mat-icon>delete</mat-icon>
                </button>
              </div>
            </div>
          </mat-card-header>
          <mat-card-content>
            <p *ngIf="ev.isInhaalAvond && ev.matches.length === 0"
               style="color:#757575;text-align:center;padding:24px 0;margin:0">
              Geen uitgestelde wedstrijden gevonden voor deze inhaalavond.
            </p>
            <mat-form-field *ngIf="ev.isInhaalAvond && ev.matches.length > 0"
                            style="width:100%;margin-bottom:8px" subscriptSizing="dynamic">
              <mat-label>Zoek op nr (bijv. 12 of 12 34)</mat-label>
              <input matInput [(ngModel)]="catchUpSearch" placeholder="nr A   nr B">
              <button matSuffix mat-icon-button *ngIf="catchUpSearch" (click)="catchUpSearch=''">
                <mat-icon>close</mat-icon>
              </button>
              <mat-icon matPrefix style="margin-right:4px">search</mat-icon>
            </mat-form-field>
            <table mat-table [dataSource]="ev.isInhaalAvond ? filteredMatches(ev) : ev.matches" *ngIf="ev.matches.length > 0">
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
                  <span *ngIf="m.played">{{ m.scoreA }} – {{ m.scoreB }}</span>
                  <span *ngIf="!m.played && m.reportedBy"
                        style="color:#e65100;font-size:12px;font-weight:500">
                    <mat-icon style="font-size:14px;vertical-align:middle;height:14px;width:14px">schedule</mat-icon>
                    Afgemeld: {{ m.reportedBy }}
                  </span>
                  <span *ngIf="!m.played && !m.reportedBy" style="color:#bdbdbd">—</span>
                </td>
              </ng-container>
              <ng-container matColumnDef="actions">
                <th mat-header-cell *matHeaderCellDef></th>
                <td mat-cell *matCellDef="let m" style="text-align:right">
                  <button mat-stroked-button color="primary" *ngIf="!m.played"
                    (click)="openScore(m)">
                    <mat-icon>edit</mat-icon> Score
                  </button>
                  <button mat-button color="accent" *ngIf="m.played"
                    (click)="openScore(m)" matTooltip="Score wijzigen">
                    <mat-icon>check_circle</mat-icon> Wijzigen
                  </button>
                </td>
              </ng-container>
              <tr mat-header-row *matHeaderRowDef="matchCols"></tr>
              <tr mat-row *matRowDef="let row; columns: matchCols;"
                  [class.match-played]="row.played"></tr>
            </table>
          </mat-card-content>
        </mat-card>
      </mat-tab>
    </mat-tab-group>

    <!-- Print-only schedule matrix -->
    <div class="print-only" *ngIf="schedule && printData">
      <p class="print-schedule-title">{{ schedule.competitionName }} — {{ schedule.season }}</p>

      <!-- Page 1: first half of evenings -->
      <table class="print-schedule-table">
        <thead>
          <tr>
            <th class="row-nr">#</th>
            <th *ngFor="let ev of printData.half1">
              {{ ev.number }}<br><span style="font-weight:400">{{ ev.date | date:'d MMM' }}</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr *ngFor="let row of printData.rows1; let i = index">
            <td class="row-nr">{{ i + 1 }}</td>
            <ng-container *ngFor="let col of printData.half1; let ci = index">
              <td *ngIf="col.isCatchUp && i === 0"
                  [attr.rowspan]="printData.rowCount"
                  style="writing-mode:vertical-lr;text-align:center;vertical-align:middle;font-weight:600;letter-spacing:3px;color:#7b1fa2;font-size:7pt">
                INHAALAVOND
              </td>
              <td *ngIf="!col.isCatchUp">{{ row[ci] }}</td>
            </ng-container>
          </tr>
        </tbody>
      </table>

      <!-- Page 2: second half of evenings -->
      <div style="page-break-before:always"></div>
      <p class="print-schedule-title">{{ schedule.competitionName }} — {{ schedule.season }}</p>
      <table class="print-schedule-table">
        <thead>
          <tr>
            <th class="row-nr">#</th>
            <th *ngFor="let ev of printData.half2">
              {{ ev.number }}<br><span style="font-weight:400">{{ ev.date | date:'d MMM' }}</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr *ngFor="let row of printData.rows2; let i = index">
            <td class="row-nr">{{ i + 1 }}</td>
            <ng-container *ngFor="let col of printData.half2; let ci = index">
              <td *ngIf="col.isCatchUp && i === 0"
                  [attr.rowspan]="printData.rowCount"
                  style="writing-mode:vertical-lr;text-align:center;vertical-align:middle;font-weight:600;letter-spacing:3px;color:#7b1fa2;font-size:7pt">
                INHAALAVOND
              </td>
              <td *ngIf="!col.isCatchUp">{{ row[ci] }}</td>
            </ng-container>
          </tr>
        </tbody>
      </table>
    </div>
  `,
})
export class OverviewComponent implements OnInit {
  private scheduleService  = inject(ScheduleService);
  private playerService    = inject(PlayerService);
  private scoreService     = inject(ScoreService);
  private seasonService    = inject(SeasonService);
  private snackBar         = inject(MatSnackBar);
  private dialog           = inject(MatDialog);
  private destroyRef       = inject(DestroyRef);

  schedule: Schedule | null = null;
  players:  Player[] = [];
  activeTab = 0;
  matchCols = ['playerA', 'vs', 'playerB', 'score', 'actions'];
  catchUpSearch = '';

  ngOnInit(): void {
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      filter(id => !!id),
      distinctUntilChanged(),
    ).subscribe(id => {
      if (this.schedule?.id !== id) this.loadScheduleById(id);
    });
    this.playerService.list().subscribe({ next: (ps) => (this.players = ps), error: () => {} });
  }

  loadScheduleById(id: string, preserveTab = false): void {
    this.scheduleService.getById(id).subscribe({
      next: (s) => {
        this.schedule = s;
        if (!preserveTab) this.activeTab = this.firstUpcomingTab(s.evenings);
        console.log('[overview] schedule loaded', s.id, `evenings: ${s.evenings.length}`);
        s.evenings.forEach(ev => {
          if (ev.isInhaalAvond) {
            console.log(`[overview] inhaalavond ev#${ev.number} (${ev.date}) matches:`, ev.matches);
          }
        });
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
    const p = this.players.find((p) => p.id === id);
    if (!p) return id.slice(0, 8);
    return p.nr ? `${p.nr} ${p.name}` : p.name;
  }

  playerNr(id: string): string {
    return this.players.find((p) => p.id === id)?.nr ?? '';
  }

  filteredMatches(ev: Evening): Match[] {
    const q = this.catchUpSearch.trim().toLowerCase();
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

  openScore(match: Match): void {
    const ref = this.dialog.open(ScoreDialogComponent, {
      data: {
        match,
        nameA: this.playerName(match.playerA),
        nameB: this.playerName(match.playerB),
        nrA:   this.playerNr(match.playerA),
        nrB:   this.playerNr(match.playerB),
        players: this.players,
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
    } | undefined) => {
      if (!result) return;
      console.log('[openScore] submitting result', match.id, result);
      this.scoreService.submitResult(match.id, result).subscribe({
        next: () => {
          this.snackBar.open('Resultaat opgeslagen!', 'OK', { duration: 2000 });
          if (this.schedule) this.loadScheduleById(this.schedule.id, true);
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
    if (!this.schedule) return null;
    const evs = this.schedule.evenings;
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
      data: { evening: ev, players: this.players } as AbsentDialogData,
      minWidth: '420px',
    });
    ref.afterClosed().subscribe((result: { playerId: string; reportedBy: string } | undefined) => {
      if (!result) return;
      this.scoreService.reportAbsent(ev.id, result.playerId, result.reportedBy).subscribe({
        next: () => {
          this.snackBar.open('Speler afgemeld', 'OK', { duration: 2000 });
          if (this.schedule) this.loadScheduleById(this.schedule.id, true);
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    });
  }

  openStatDialog(ev: Evening): void {
    const playerIdsOnEvening = new Set(ev.matches.flatMap(m => [m.playerA, m.playerB]));
    const players = this.players
      .filter(p => playerIdsOnEvening.has(p.id))
      .map(p => ({ id: p.id, name: p.name }));
    this.dialog.open(EveningStatDialogComponent, {
      data: {
        scheduleId: this.schedule!.id,
        players,
      } as EveningStatDialogData,
    }).afterClosed().subscribe(saved => {
      if (saved) this.snackBar.open('Opgeslagen', '', { duration: 2000 });
    });
  }

  exportEvening(eveningId: string): void {
    window.open(`${environment.apiBaseUrl}/export/evening/${eveningId}/excel`, '_blank');
  }

  printEvening(eveningId: string): void {
    window.open(`${environment.apiBaseUrl}/export/evening/${eveningId}/print`, '_blank');
  }

  deleteEvening(ev: Evening): void {
    if (!confirm(`Inhaalavond ${ev.number} verwijderen?`)) return;
    this.scheduleService.deleteEvening(this.schedule!.id, ev.id).subscribe({
      next: () => {
        this.snackBar.open('Avond verwijderd', 'OK', { duration: 2000 });
        this.loadScheduleById(this.schedule!.id);
      },
      error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
    });
  }

}
