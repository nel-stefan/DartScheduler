import { Component, inject, OnInit, Inject, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { filter, distinctUntilChanged } from 'rxjs';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialog, MatDialogModule, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatTabsModule } from '@angular/material/tabs';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatChipsModule } from '@angular/material/chips';
import { MatIconModule } from '@angular/material/icon';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatRadioModule } from '@angular/material/radio';
import { MatSelectModule } from '@angular/material/select';
import { ScheduleService } from '../../services/schedule.service';
import { PlayerService } from '../../services/player.service';
import { ScoreService } from '../../services/score.service';
import { SeasonService } from '../../services/season.service';
import { Schedule, Player, Match, Evening, GenerateScheduleRequest } from '../../models';
import { environment } from '../../../environments/environment';

// ---------------------------------------------------------------------------
// Generate-dialog
// ---------------------------------------------------------------------------

@Component({
  selector: 'app-generate-dialog',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatDialogModule, MatButtonModule,
            MatFormFieldModule, MatInputModule],
  template: `
    <h2 mat-dialog-title>Schema genereren</h2>
    <mat-dialog-content>
      <form [formGroup]="form" style="display:flex;flex-direction:column;gap:12px;min-width:320px;padding-top:8px">
        <mat-form-field><mat-label>Naam competitie</mat-label>
          <input matInput formControlName="competitionName">
        </mat-form-field>
        <mat-form-field><mat-label>Seizoen (bijv. 2026)</mat-label>
          <input matInput formControlName="season" placeholder="2026">
        </mat-form-field>
        <mat-form-field><mat-label>Aantal avonden</mat-label>
          <input matInput type="number" formControlName="numEvenings">
        </mat-form-field>
        <mat-form-field><mat-label>Startdatum (YYYY-MM-DD)</mat-label>
          <input matInput formControlName="startDate" placeholder="2026-04-01">
        </mat-form-field>
        <mat-form-field><mat-label>Interval (dagen)</mat-label>
          <input matInput type="number" formControlName="intervalDays">
        </mat-form-field>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annuleren</button>
      <button mat-raised-button color="primary" [disabled]="form.invalid" (click)="submit()">Genereren</button>
    </mat-dialog-actions>
  `,
})
export class GenerateDialogComponent {
  private dialogRef = inject(MatDialogRef<GenerateDialogComponent>);
  fb = inject(FormBuilder);

  form = this.fb.group({
    competitionName: ['Liga 2026', Validators.required],
    season:          ['2026', Validators.required],
    numEvenings: [20, [Validators.required, Validators.min(1)]],
    startDate: ['2026-04-01', Validators.required],
    intervalDays: [7, [Validators.required, Validators.min(1)]],
  });

  submit(): void {
    if (this.form.valid) this.dialogRef.close(this.form.value as GenerateScheduleRequest);
  }
}

// ---------------------------------------------------------------------------
// Import-season-dialog
// ---------------------------------------------------------------------------

@Component({
  selector: 'app-import-season-dialog',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatDialogModule, MatButtonModule,
            MatFormFieldModule, MatInputModule, MatIconModule],
  template: `
    <h2 mat-dialog-title>Oud seizoen importeren</h2>
    <mat-dialog-content>
      <form [formGroup]="form" style="display:flex;flex-direction:column;gap:12px;min-width:340px;padding-top:8px">
        <mat-form-field><mat-label>Naam competitie</mat-label>
          <input matInput formControlName="competitionName">
        </mat-form-field>
        <mat-form-field><mat-label>Seizoen (bijv. 2025)</mat-label>
          <input matInput formControlName="season" placeholder="2025">
        </mat-form-field>
        <div>
          <input #fileInput type="file" accept=".xlsx,.xls" hidden (change)="onFile($event)">
          <button mat-stroked-button type="button" (click)="fileInput.click()">
            <mat-icon>upload_file</mat-icon> Excel kiezen
          </button>
          <span *ngIf="file" style="margin-left:8px;color:#555;font-size:13px">{{ file.name }}</span>
        </div>
      </form>
      <p style="color:#757575;font-size:12px;margin-top:12px">
        Ondersteunde formaten:<br>
        • <strong>Speelschema matrix</strong>: rij 1 = datums, cellen = "nr - nr" of "nr (naam) - nr (naam)".
        Kolommen met "INHAAL" worden als inhaalavonden geïmporteerd.<br>
        • <strong>Platte tabel</strong>: kolommen avond, datum, nr a, naam a, nr b, naam b, leg1, beurten1, …
      </p>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annuleren</button>
      <button mat-raised-button color="primary" [disabled]="form.invalid || !file" (click)="submit()">Importeren</button>
    </mat-dialog-actions>
  `,
})
export class ImportSeasonDialogComponent {
  private dialogRef = inject(MatDialogRef<ImportSeasonDialogComponent>);
  fb = inject(FormBuilder);
  file: File | null = null;

  form = this.fb.group({
    competitionName: ['', Validators.required],
    season:          ['', Validators.required],
  });

  onFile(e: Event): void {
    this.file = (e.target as HTMLInputElement).files?.[0] ?? null;
    if (this.file && !this.form.value.competitionName) {
      const name = this.file.name.replace(/\.xlsx?$/i, '');
      this.form.patchValue({ competitionName: name, season: name.replace(/\D/g, '') });
    }
  }

  submit(): void {
    if (this.form.valid && this.file) {
      this.dialogRef.close({ file: this.file, ...this.form.value });
    }
  }
}

// ---------------------------------------------------------------------------
// Add-inhaalavond-dialog
// ---------------------------------------------------------------------------

@Component({
  selector: 'app-add-inhaal-dialog',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatDialogModule, MatButtonModule,
            MatFormFieldModule, MatInputModule],
  template: `
    <h2 mat-dialog-title>Inhaalavond toevoegen</h2>
    <mat-dialog-content>
      <form [formGroup]="form" style="display:flex;flex-direction:column;gap:12px;min-width:300px;padding-top:8px">
        <mat-form-field>
          <mat-label>Datum (JJJJ-MM-DD)</mat-label>
          <input matInput formControlName="date" placeholder="2026-03-22">
        </mat-form-field>
      </form>
      <p style="color:#757575;font-size:12px;margin-top:8px">
        Alle ongespeelde wedstrijden van avonden vóór deze datum worden naar deze inhaalavond verplaatst.
      </p>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annuleren</button>
      <button mat-raised-button color="primary" [disabled]="form.invalid" (click)="submit()">Toevoegen</button>
    </mat-dialog-actions>
  `,
})
export class AddInhaalAvondDialogComponent {
  private dialogRef = inject(MatDialogRef<AddInhaalAvondDialogComponent>);
  fb = inject(FormBuilder);

  form = this.fb.group({
    date: ['', [Validators.required, Validators.pattern(/^\d{4}-\d{2}-\d{2}$/)]],
  });

  submit(): void {
    if (this.form.valid) this.dialogRef.close(this.form.value.date as string);
  }
}

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
            MatFormFieldModule, MatInputModule, MatRadioModule, MatSelectModule],
  styles: [`
    .leg-row { display: flex; align-items: center; gap: 16px; margin-bottom: 8px; flex-wrap: wrap; }
    .leg-label { font-weight: 600; min-width: 60px; }
    .leg-note { font-size: 11px; color: #757575; margin-left: 4px; }
    .admin-row { display: flex; gap: 12px; flex-wrap: wrap; margin-top: 8px; }
  `],
  template: `
    <h2 mat-dialog-title>Wedstrijdresultaat invoeren</h2>
    <mat-dialog-content style="min-width:420px">
      <div style="margin-bottom:16px;font-size:15px;text-align:center">
        <strong>{{ data.nrA }} {{ data.nameA }}</strong>
        &nbsp;vs&nbsp;
        <strong>{{ data.nrB }} {{ data.nameB }}</strong>
        &nbsp;&nbsp;<span style="color:#757575;font-size:12px">Best of 3</span>
      </div>
      <form [formGroup]="form" style="padding-top:4px">
        <!-- Leg 1 -->
        <div class="leg-row">
          <span class="leg-label">Partij 1</span>
          <mat-radio-group formControlName="leg1Winner" style="display:flex;gap:12px">
            <label><input type="radio" [value]="data.match.playerA" formControlName="leg1Winner"> {{ data.nameA }}</label>
            <label><input type="radio" [value]="data.match.playerB" formControlName="leg1Winner"> {{ data.nameB }}</label>
          </mat-radio-group>
          <mat-form-field style="width:110px">
            <mat-label>Aantal beurten</mat-label>
            <input matInput type="number" formControlName="leg1Turns" min="1">
          </mat-form-field>
        </div>
        <!-- Leg 2 -->
        <div class="leg-row">
          <span class="leg-label">Partij 2</span>
          <mat-radio-group formControlName="leg2Winner" style="display:flex;gap:12px">
            <label><input type="radio" [value]="data.match.playerA" formControlName="leg2Winner"> {{ data.nameA }}</label>
            <label><input type="radio" [value]="data.match.playerB" formControlName="leg2Winner"> {{ data.nameB }}</label>
          </mat-radio-group>
          <mat-form-field style="width:110px">
            <mat-label>Aantal beurten</mat-label>
            <input matInput type="number" formControlName="leg2Turns" min="1">
          </mat-form-field>
        </div>
        <!-- Leg 3 -->
        <div class="leg-row">
          <span class="leg-label">Partij 3</span>
          <mat-radio-group formControlName="leg3Winner" style="display:flex;gap:12px">
            <label><input type="radio" [value]="data.match.playerA" formControlName="leg3Winner"> {{ data.nameA }}</label>
            <label><input type="radio" [value]="data.match.playerB" formControlName="leg3Winner"> {{ data.nameB }}</label>
          </mat-radio-group>
          <mat-form-field style="width:110px">
            <mat-label>Aantal beurten</mat-label>
            <input matInput type="number" formControlName="leg3Turns" min="1">
          </mat-form-field>
        </div>
        <!-- Administrative fields -->
        <div class="admin-row">
          <mat-form-field style="flex:1;min-width:150px">
            <mat-label>Afgemeld door</mat-label>
            <input matInput formControlName="reportedBy">
          </mat-form-field>
          <mat-form-field style="flex:1;min-width:130px">
            <mat-label>Vooruitgooi datum</mat-label>
            <input matInput formControlName="rescheduleDate" placeholder="DD-MM-JJJJ">
          </mat-form-field>
        </div>
        <div class="admin-row">
          <mat-form-field style="flex:1;min-width:160px">
            <mat-label>Schrijver</mat-label>
            <mat-select formControlName="secretaryNr">
              <mat-option *ngFor="let p of data.players" [value]="p.nr">{{ p.nr }} – {{ p.name }}</mat-option>
            </mat-select>
          </mat-form-field>
          <mat-form-field style="flex:1;min-width:160px">
            <mat-label>Teller</mat-label>
            <mat-select formControlName="counterNr">
              <mat-option *ngFor="let p of data.players" [value]="p.nr">{{ p.nr }} – {{ p.name }}</mat-option>
            </mat-select>
          </mat-form-field>
        </div>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annuleren</button>
      <button mat-raised-button color="primary" [disabled]="!isValid()" (click)="submit()">Opslaan</button>
    </mat-dialog-actions>
  `,
})
export class ScoreDialogComponent {
  private dialogRef = inject(MatDialogRef<ScoreDialogComponent>);
  fb = inject(FormBuilder);

  constructor(@Inject(MAT_DIALOG_DATA) public data: ScoreDialogData) {}

  form = this.fb.group({
    leg1Winner: ['', Validators.required],
    leg1Turns:  [null as number | null, Validators.min(1)],
    leg2Winner: ['', Validators.required],
    leg2Turns:  [null as number | null, Validators.min(1)],
    leg3Winner: ['', Validators.required],
    leg3Turns:  [null as number | null, Validators.min(1)],
    reportedBy:     [''],
    rescheduleDate: [''],
    secretaryNr:    [''],
    counterNr:      [''],
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
      reportedBy:     v.reportedBy ?? '',
      rescheduleDate: v.rescheduleDate ?? '',
      secretaryNr:    v.secretaryNr ?? '',
      counterNr:      v.counterNr ?? '',
    });
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
    MatSnackBarModule, MatDialogModule, MatChipsModule, MatIconModule,
    MatTooltipModule, MatSelectModule, MatFormFieldModule, MatInputModule,
    ReactiveFormsModule,
    AddInhaalAvondDialogComponent,
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
    .tab-label { display: flex; align-items: center; gap: 6px; }
    .dot {
      width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0;
    }
    .dot-done    { background: #2e7d32; }
    .dot-partial { background: #f57c00; }
    .dot-open    { background: #bdbdbd; }
    .dot-inhaal  { background: #7b1fa2; }
    table { width: 100%; }
    .score-cell { font-weight: 600; color: #2e7d32; }
    .vs-cell { color: #9e9e9e; width: 36px; text-align: center; }
    .empty-state {
      padding: 48px 0;
      text-align: center;
      color: #757575;
    }
    .empty-state mat-icon { font-size: 64px; width: 64px; height: 64px; color: #bdbdbd; }
    .card-header-row { display: flex; align-items: flex-start; justify-content: space-between; }
  `],
  template: `
    <div class="schedule-header">
      <h2 class="schedule-title">{{ schedule?.competitionName ?? 'DartScheduler' }}</h2>

      <button mat-raised-button color="primary" (click)="openGenerate()">
        <mat-icon>auto_awesome</mat-icon> Schema genereren
      </button>
      <button mat-stroked-button (click)="openImportSeason()">
        <mat-icon>history</mat-icon> Oud seizoen importeren
      </button>
      <button mat-stroked-button color="accent" *ngIf="schedule" (click)="openAddInhaalAvond()">
        <mat-icon>replay</mat-icon> Inhaalavond toevoegen
      </button>
    </div>

    <div *ngIf="!schedule" class="empty-state">
      <mat-icon>sports_bar</mat-icon>
      <p>Nog geen schema. Importeer eerst spelers via <strong>Spelers</strong>, dan genereer je hier een schema.</p>
    </div>

    <mat-tab-group *ngIf="schedule" animationDuration="150ms" [selectedIndex]="activeTab"
                   (selectedIndexChange)="activeTab = $event" color="primary" backgroundColor="primary">
      <mat-tab *ngFor="let ev of schedule.evenings">
        <ng-template mat-tab-label>
          <span class="tab-label">
            <span class="dot"
              [class.dot-inhaal]="ev.isInhaalAvond"
              [class.dot-done]="!ev.isInhaalAvond && allPlayed(ev)"
              [class.dot-partial]="!ev.isInhaalAvond && somePlayed(ev)"
              [class.dot-open]="!ev.isInhaalAvond && nonePlayed(ev)">
            </span>
            {{ ev.number }}
          </span>
        </ng-template>

        <!-- Tab-inhoud -->
        <mat-card style="border-radius:0 0 8px 8px; border-top: none;">
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
                  {{ playedCount(ev) }} / {{ ev.matches?.length ?? 0 }} wedstrijden gespeeld
                </mat-card-subtitle>
                <mat-card-subtitle *ngIf="ev.isInhaalAvond" style="color:#7b1fa2">
                  Vrije avond voor uitgestelde wedstrijden
                </mat-card-subtitle>
              </div>
              <button mat-icon-button (click)="exportEvening(ev.id)" matTooltip="Exporteren naar Excel"
                      *ngIf="!ev.isInhaalAvond">
                <mat-icon>file_download</mat-icon>
              </button>
            </div>
          </mat-card-header>
          <mat-card-content *ngIf="ev.isInhaalAvond" style="padding:16px 0 8px 0;color:#757575;text-align:center">
            <mat-icon style="font-size:40px;width:40px;height:40px;color:#ce93d8">replay</mat-icon>
            <p style="margin:8px 0 0 0">Op deze avond kunnen uitgestelde wedstrijden worden ingehaald.</p>
          </mat-card-content>
          <mat-card-content *ngIf="!ev.isInhaalAvond">
            <table mat-table [dataSource]="ev.matches ?? []">
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
                  {{ m.played ? (m.scoreA + ' – ' + m.scoreB) : '—' }}
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
  `,
})
export class OverviewComponent implements OnInit {
  private scheduleService = inject(ScheduleService);
  private playerService   = inject(PlayerService);
  private scoreService    = inject(ScoreService);
  private seasonService   = inject(SeasonService);
  private snackBar        = inject(MatSnackBar);
  private dialog          = inject(MatDialog);
  private destroyRef      = inject(DestroyRef);

  schedule: Schedule | null = null;
  players:  Player[] = [];
  activeTab = 0;
  matchCols = ['playerA', 'vs', 'playerB', 'score', 'actions'];

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

  loadScheduleById(id: string): void {
    this.scheduleService.getById(id).subscribe({
      next: (s) => { this.schedule = s; this.activeTab = 0; },
      error: () => {},
    });
  }

  playerName(id: string): string {
    const p = this.players.find((p) => p.id === id);
    if (!p) return id.slice(0, 8);
    return p.nr ? `${p.nr} ${p.name}` : p.name;
  }

  playerNr(id: string): string {
    return this.players.find((p) => p.id === id)?.nr ?? '';
  }

  playedCount(ev: Evening): number {
    return ev.matches?.filter(m => m.played).length ?? 0;
  }

  allPlayed(ev: Evening):  boolean { return (ev.matches?.length ?? 0) > 0 && this.playedCount(ev) === (ev.matches?.length ?? 0); }
  nonePlayed(ev: Evening): boolean { return this.playedCount(ev) === 0; }
  somePlayed(ev: Evening): boolean { return !this.allPlayed(ev) && !this.nonePlayed(ev); }

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
      reportedBy: string; rescheduleDate: string;
      secretaryNr: string; counterNr: string;
    } | undefined) => {
      if (result == null) return;
      this.scoreService.submitResult(match.id, result).subscribe({
        next: () => {
          this.snackBar.open('Resultaat opgeslagen!', 'OK', { duration: 2000 });
          if (this.schedule) this.loadScheduleById(this.schedule.id);
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    });
  }

  exportEvening(eveningId: string): void {
    window.open(`${environment.apiBaseUrl}/export/evening/${eveningId}/excel`, '_blank');
  }

  openGenerate(): void {
    const ref = this.dialog.open(GenerateDialogComponent);
    ref.afterClosed().subscribe((req: GenerateScheduleRequest | undefined) => {
      if (!req) return;
      this.scheduleService.generate(req).subscribe({
        next: (s) => {
          this.snackBar.open('Schema gegenereerd!', 'OK', { duration: 3000 });
          this.seasonService.load(s.id);
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    });
  }

  openAddInhaalAvond(): void {
    if (!this.schedule) return;
    const scheduleId = this.schedule.id;
    const ref = this.dialog.open(AddInhaalAvondDialogComponent);
    ref.afterClosed().subscribe((date: string | undefined) => {
      if (!date) return;
      this.scheduleService.addInhaalAvond(scheduleId, date).subscribe({
        next: (s) => {
          this.schedule = s;
          this.activeTab = s.evenings.length - 1;
          this.snackBar.open('Inhaalavond toegevoegd!', 'OK', { duration: 2000 });
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    });
  }

  openImportSeason(): void {
    const ref = this.dialog.open(ImportSeasonDialogComponent);
    ref.afterClosed().subscribe((result: { file: File; competitionName: string; season: string } | undefined) => {
      if (!result) return;
      this.scheduleService.importSeason(result.file, result.competitionName, result.season).subscribe({
        next: (s) => {
          this.snackBar.open('Seizoen geïmporteerd!', 'OK', { duration: 3000 });
          this.seasonService.load(s.id);
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    });
  }
}
