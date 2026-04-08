import {
  Component,
  inject,
  OnInit,
  Inject,
  ElementRef,
  ChangeDetectorRef,
  viewChild,
  signal,
  Signal,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialog, MatDialogModule, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatTabsModule } from '@angular/material/tabs';
import { MatDividerModule } from '@angular/material/divider';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { ScheduleService } from '../../services/schedule.service';
import { PlayerService } from '../../services/player.service';
import { SeasonService } from '../../services/season.service';
import { SystemService } from '../../services/system.service';
import { SeasonSummary, GenerateScheduleRequest } from '../../models';
import { environment } from '../../../environments/environment';

// ---------------------------------------------------------------------------
// Loading-dialog (shown while the scheduler runs)
// ---------------------------------------------------------------------------

@Component({
  selector: 'app-loading-dialog',
  imports: [MatProgressSpinnerModule, MatProgressBarModule, CommonModule],
  template: `
    <div
      style="display:flex;flex-direction:column;align-items:center;gap:20px;padding:40px 56px;text-align:center;min-width:280px"
    >
      @if (data.percent() < 5) {
        <mat-spinner diameter="56"></mat-spinner>
      } @else {
        <div style="width:220px">
          <mat-progress-bar mode="determinate" [value]="data.percent()"></mat-progress-bar>
          <span style="font-size:13px;color:#555;margin-top:6px;display:block">{{ data.percent() }}%</span>
        </div>
      }
      <div>
        <p style="margin:0;font-size:16px;font-weight:500">Schema wordt berekend…</p>
        <p style="margin:6px 0 0;font-size:13px;color:#9e9e9e">Dit kan een minuut duren.</p>
      </div>
    </div>
  `,
})
export class LoadingDialogComponent {
  data = inject<{ percent: Signal<number> }>(MAT_DIALOG_DATA);
}

// ---------------------------------------------------------------------------
// Constraint-violation-dialog (shown when the scheduler returns HTTP 422)
// ---------------------------------------------------------------------------

@Component({
  selector: 'app-constraint-violation-dialog',
  imports: [MatDialogModule, MatButtonModule],
  template: `
    <h2 mat-dialog-title style="color:#b71c1c">Schema voldoet niet aan constraints</h2>
    <mat-dialog-content>
      <p style="margin:0 0 12px;font-size:14px;color:#555">
        Het gegenereerde schema voldoet niet aan de volgende harde constraints:
      </p>
      <ul style="margin:0;padding-left:20px;font-size:14px;line-height:1.8">
        @for (line of lines; track line) {
          <li>{{ line }}</li>
        }
      </ul>
      <p style="margin:16px 0 0;font-size:13px;color:#9e9e9e">
        Probeer het schema opnieuw te genereren. Als het probleem aanhoudt, pas dan het aantal speelavonden aan.
      </p>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Sluiten</button>
    </mat-dialog-actions>
  `,
})
export class ConstraintViolationDialogComponent {
  data = inject<{ message: string }>(MAT_DIALOG_DATA);
  get lines(): string[] {
    return this.data.message
      .split('\n')
      .map((l) => l.replace(/^[•\-]\s*/, '').trim())
      .filter((l) => l.length > 0 && !l.startsWith('het gegenereerde schema'));
  }
}

// ---------------------------------------------------------------------------
// Generate-dialog
// ---------------------------------------------------------------------------

@Component({
  selector: 'app-generate-dialog',
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatButtonModule,
    MatIconModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
  ],
  styles: [
    `
      .slot-row {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 4px 0;
      }
      .slot-date {
        color: #555;
        min-width: 180px;
        font-size: 13px;
      }
      .slot-nr {
        min-width: 64px;
        font-size: 13px;
        font-weight: 500;
      }
      .insert-row {
        display: flex;
        align-items: center;
        gap: 4px;
        height: 20px;
        opacity: 0.25;
        transition: opacity 0.15s;
        cursor: pointer;
      }
      .insert-row:hover {
        opacity: 1;
      }
      .insert-line {
        flex: 1;
        height: 1px;
        background: #bdbdbd;
      }
      .insert-btn {
        width: 20px !important;
        height: 20px !important;
        line-height: 20px !important;
      }
      .insert-btn mat-icon {
        font-size: 16px;
        width: 16px;
        height: 16px;
        line-height: 16px;
        color: #757575;
      }
    `,
  ],
  template: `
    <h2 mat-dialog-title>Schema genereren</h2>
    <mat-dialog-content style="min-width:480px">
      <form [formGroup]="form" style="display:grid;grid-template-columns:1fr 1fr;gap:8px 16px;padding-top:8px">
        <mat-form-field style="grid-column:1/-1"
          ><mat-label>Naam competitie</mat-label>
          <input matInput formControlName="competitionName" />
        </mat-form-field>
        <mat-form-field style="grid-column:1/-1"
          ><mat-label>Seizoen (bijv. 2026-2027)</mat-label>
          <input matInput formControlName="season" placeholder="2026-2027" />
        </mat-form-field>
        <mat-form-field
          ><mat-label>Aantal avonden (totaal)</mat-label>
          <input matInput type="number" formControlName="numEvenings" />
        </mat-form-field>
        <mat-form-field
          ><mat-label>Interval (dagen)</mat-label>
          <input matInput type="number" formControlName="intervalDays" />
        </mat-form-field>
        <mat-form-field style="grid-column:1/-1"
          ><mat-label>Startdatum (YYYY-MM-DD)</mat-label>
          <input matInput formControlName="startDate" placeholder="2026-09-01" />
        </mat-form-field>
      </form>

      <!-- Avondenlijst -->
      @if (slots.length > 0) {
        <div style="margin-top:16px">
          <div style="font-weight:500;margin-bottom:8px;font-size:14px">
            Avondtype instellen
            <span style="color:#888;font-size:12px;font-weight:400;margin-left:8px">
              {{ regularCount }} speelavonden · {{ inhaalCount }} inhaalavonden · {{ vrijCount }} vrij
            </span>
          </div>
          <div style="max-height:320px;overflow-y:auto">
            @for (s of slots; track s.nr; let i = $index) {
              <div class="slot-row">
                <span class="slot-nr">Avond {{ s.nr }}</span>
                <span class="slot-date">{{ s.date | date: 'EEE d MMM yyyy' }}</span>
                <mat-form-field style="min-width:130px" subscriptSizing="dynamic">
                  <mat-select [(value)]="s.type">
                    <mat-option value="normaal">Normaal</mat-option>
                    <mat-option value="inhaal">Inhaalavond</mat-option>
                    <mat-option value="vrij">Vrije avond</mat-option>
                  </mat-select>
                </mat-form-field>
              </div>
              <div class="insert-row" (click)="insertVrij(i)" title="Vakantieweek invoegen">
                <div class="insert-line"></div>
                <button type="button" mat-icon-button class="insert-btn">
                  <mat-icon>add_circle_outline</mat-icon>
                </button>
                <div class="insert-line"></div>
              </div>
            }
          </div>
        </div>
      }
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annuleren</button>
      <button mat-raised-button color="primary" [disabled]="form.invalid || regularCount === 0" (click)="submit()">
        Genereren
      </button>
    </mat-dialog-actions>
  `,
})
export class GenerateDialogComponent implements OnInit {
  private dialogRef = inject(MatDialogRef<GenerateDialogComponent>);
  fb = inject(FormBuilder);

  constructor(@Inject(MAT_DIALOG_DATA) public data: null) {}

  form = this.fb.group({
    competitionName: ['2026-2027', Validators.required],
    season: ['2026-2027', Validators.required],
    numEvenings: [30, [Validators.required, Validators.min(1)]],
    startDate: ['2026-09-01', Validators.required],
    intervalDays: [7, [Validators.required, Validators.min(1)]],
  });

  slots: { nr: number; date: Date; type: string }[] = [];

  ngOnInit(): void {
    // Sync competitionName ↔ season so both fields always match.
    this.form.get('competitionName')!.valueChanges.subscribe((v) => {
      this.form.get('season')!.setValue(v, { emitEvent: false });
    });
    this.form.get('season')!.valueChanges.subscribe((v) => {
      this.form.get('competitionName')!.setValue(v, { emitEvent: false });
    });

    // Rebuild slot list when scheduling parameters change.
    ['numEvenings', 'startDate', 'intervalDays'].forEach((field) => {
      this.form.get(field)!.valueChanges.subscribe(() => this.rebuildSlots());
    });

    this.rebuildSlots();
  }

  get regularCount(): number {
    return this.slots.filter((s) => s.type === 'normaal').length;
  }
  get inhaalCount(): number {
    return this.slots.filter((s) => s.type === 'inhaal').length;
  }
  get vrijCount(): number {
    return this.slots.filter((s) => s.type === 'vrij').length;
  }

  private rebuildSlots(): void {
    const v = this.form.value;
    const n = v.numEvenings ?? 0;
    const start = v.startDate ? new Date(v.startDate) : null;
    const interval = v.intervalDays ?? 7;
    if (n > 0 && start && !isNaN(start.getTime())) {
      // Last 4 slots default to 'inhaal'; the rest are 'normaal'.
      this.slots = Array.from({ length: n }, (_, i) => {
        const d = new Date(start);
        d.setDate(d.getDate() + i * interval);
        const nr = i + 1;
        return { nr, date: d, type: [8, 15, 23, 30].includes(nr) ? 'inhaal' : 'normaal' };
      });
    } else {
      this.slots = [];
    }
  }

  /** Insert a vrij (vacation) slot after the slot at the given 0-based index. */
  insertVrij(afterIndex: number): void {
    const start = this.form.value.startDate ? new Date(this.form.value.startDate) : new Date();
    const interval = this.form.value.intervalDays ?? 7;

    const newSlots = [...this.slots];
    newSlots.splice(afterIndex + 1, 0, { nr: 0, date: new Date(), type: 'vrij' });

    // Renumber all slots and recompute dates from the start date.
    this.slots = newSlots.map((s, i) => {
      const d = new Date(start);
      d.setDate(d.getDate() + i * interval);
      return { ...s, nr: i + 1, date: d };
    });
  }

  submit(): void {
    if (!this.form.valid || this.regularCount === 0) return;
    const v = this.form.value;
    const inhaalNrs = this.slots.filter((s) => s.type === 'inhaal').map((s) => s.nr);
    const vrijeNrs = this.slots.filter((s) => s.type === 'vrij').map((s) => s.nr);
    this.dialogRef.close({
      competitionName: v.competitionName,
      season: v.season,
      numEvenings: this.slots.length,
      startDate: v.startDate,
      intervalDays: v.intervalDays,
      inhaalNrs,
      vrijeNrs,
    } as GenerateScheduleRequest);
  }
}

// ---------------------------------------------------------------------------
// Import-season-dialog
// ---------------------------------------------------------------------------

@Component({
  selector: 'app-import-season-dialog',
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    MatIconModule,
  ],
  template: `
    <h2 mat-dialog-title>Oud seizoen importeren</h2>
    <mat-dialog-content>
      <form [formGroup]="form" style="display:flex;flex-direction:column;gap:12px;min-width:340px;padding-top:8px">
        <mat-form-field
          ><mat-label>Naam competitie</mat-label>
          <input matInput formControlName="competitionName" />
        </mat-form-field>
        <mat-form-field
          ><mat-label>Seizoen (bijv. 2025)</mat-label>
          <input matInput formControlName="season" placeholder="2025" />
        </mat-form-field>
        <div>
          <input #fileInput type="file" accept=".xlsx,.xls" hidden (change)="onFile($event)" />
          <button mat-stroked-button type="button" (click)="fileInput.click()">
            <mat-icon>upload_file</mat-icon> Excel kiezen
          </button>
          @if (file) {
            <span style="margin-left:8px;color:#555;font-size:13px">{{ file.name }}</span>
          }
        </div>
      </form>
      <p style="color:#757575;font-size:12px;margin-top:12px">
        Ondersteunde formaten:<br />
        • <strong>Speelschema matrix</strong>: rij 1 = datums, cellen = "nr - nr" of "nr (naam) - nr (naam)". Kolommen
        met "INHAAL" worden als inhaalavonden geïmporteerd.<br />
        • <strong>Platte tabel</strong>: kolommen avond, datum, nr a, naam a, nr b, naam b, leg1, beurten1, …
      </p>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annuleren</button>
      <button mat-raised-button color="primary" [disabled]="form.invalid || !file" (click)="submit()">
        Importeren
      </button>
    </mat-dialog-actions>
  `,
})
export class ImportSeasonDialogComponent {
  private dialogRef = inject(MatDialogRef<ImportSeasonDialogComponent>);
  fb = inject(FormBuilder);
  file: File | null = null;

  form = this.fb.group({
    competitionName: ['', Validators.required],
    season: ['', Validators.required],
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
// Beheer-component
// ---------------------------------------------------------------------------

@Component({
  selector: 'app-beheer',
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    MatSnackBarModule,
    MatButtonModule,
    MatIconModule,
    MatCardModule,
    MatDialogModule,
    MatDividerModule,
    MatTooltipModule,
    MatFormFieldModule,
    MatInputModule,
    MatTabsModule,
  ],
  styles: [
    `
      .section-title {
        font-size: 18px;
        font-weight: 500;
        margin: 0 0 16px;
      }
      .seasons-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
      }
      .season-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 10px 16px;
        background: #fafafa;
        border: 1px solid #e0e0e0;
        border-radius: 6px;
      }
      .season-name {
        font-weight: 500;
      }
      .season-meta {
        font-size: 12px;
        color: #757575;
        margin-top: 2px;
      }
      .rename-input {
        font-size: 14px;
        font-weight: 500;
        font-family: inherit;
        border: none;
        border-bottom: 2px solid #795548;
        outline: none;
        background: transparent;
        padding: 2px 0;
        min-width: 180px;
      }
      .import-row {
        display: flex;
        align-items: center;
        gap: 12px;
        flex-wrap: wrap;
      }
      .action-bar {
        display: flex;
        gap: 12px;
        flex-wrap: wrap;
        margin-bottom: 16px;
      }
      .server-meta {
        display: flex;
        align-items: center;
        gap: 16px;
        margin-bottom: 16px;
      }
      .version-chip {
        background: #e8f5e9;
        color: #2e7d32;
        border-radius: 12px;
        padding: 4px 12px;
        font-size: 13px;
        font-weight: 500;
      }
      .log-box {
        background: #1e1e1e;
        color: #d4d4d4;
        font-family: monospace;
        font-size: 12px;
        line-height: 1.5;
        padding: 12px 16px;
        border-radius: 6px;
        max-height: 480px;
        overflow-y: auto;
        white-space: pre-wrap;
        word-break: break-all;
      }
      .log-empty {
        color: #9e9e9e;
        font-style: italic;
        font-size: 13px;
      }
      @keyframes spin {
        from {
          transform: rotate(0deg);
        }
        to {
          transform: rotate(360deg);
        }
      }
    `,
  ],
  template: `
    <div style="max-width:700px">
      <!-- Seizoenen -->
      <mat-card style="margin-bottom:24px">
        <mat-card-header>
          <mat-card-title>Seizoenen</mat-card-title>
        </mat-card-header>
        <mat-card-content style="padding-top:16px">
          <div class="action-bar">
            <button mat-raised-button color="primary" (click)="openGenerate()">
              <mat-icon>auto_awesome</mat-icon> Schema genereren
            </button>
            <button mat-stroked-button (click)="openImportSeason()">
              <mat-icon>history</mat-icon> Oud seizoen importeren
            </button>
          </div>

          <mat-divider style="margin-bottom:16px"></mat-divider>

          @if (seasons().length === 0) {
            <p style="color:#888;text-align:center;padding:16px 0;margin:0">Nog geen seizoenen aangemaakt.</p>
          }

          <div class="seasons-list">
            @for (s of seasons(); track s) {
              <div class="season-row">
                <div style="flex:1;min-width:0">
                  @if (editingSeasonId() !== s.id) {
                    <div class="season-name">
                      {{ s.competitionName }}
                      @if (s.active) {
                        <span style="font-size:11px;color:#f9a825;font-weight:600;margin-left:6px">actief</span>
                      }
                    </div>
                  }
                  @if (editingSeasonId() === s.id) {
                    <input
                      #renameInput
                      class="rename-input"
                      [ngModel]="renameDraft()"
                      (ngModelChange)="renameDraft.set($event)"
                      (keydown.enter)="saveRename(s)"
                      (keydown.escape)="cancelRename()"
                      (blur)="saveRename(s)"
                    />
                  }
                  <div class="season-meta">Seizoen {{ s.season }} · {{ s.eveningCount }} avonden</div>
                </div>
                @if (editingSeasonId() !== s.id) {
                  @if (s.active) {
                    <mat-icon style="color:#f9a825;margin:0 4px;vertical-align:middle" matTooltip="Actief seizoen"
                      >star</mat-icon
                    >
                  } @else {
                    <button mat-icon-button (click)="setActiveSeason(s)" matTooltip="Als actief instellen">
                      <mat-icon>star_border</mat-icon>
                    </button>
                  }
                }
                @if (editingSeasonId() !== s.id) {
                  <button mat-icon-button (click)="startRename(s)" matTooltip="Naam aanpassen">
                    <mat-icon>edit</mat-icon>
                  </button>
                }
                @if (editingSeasonId() !== s.id) {
                  <button
                    mat-icon-button
                    [disabled]="regeneratingId() !== ''"
                    (click)="regenerateSeason(s)"
                    matTooltip="Schema herberekenen"
                  >
                    @if (regeneratingId() === s.id) {
                      <mat-icon style="animation:spin 1s linear infinite">sync</mat-icon>
                    } @else {
                      <mat-icon>sync</mat-icon>
                    }
                  </button>
                }
                @if (editingSeasonId() !== s.id) {
                  <button
                    mat-icon-button
                    color="warn"
                    [disabled]="regeneratingId() !== ''"
                    (click)="deleteSeason(s)"
                    matTooltip="Seizoen verwijderen"
                  >
                    <mat-icon>delete</mat-icon>
                  </button>
                }
              </div>
            }
          </div>
        </mat-card-content>
      </mat-card>

      <!-- Spelers importeren -->
      <mat-card>
        <mat-card-header>
          <mat-card-title>Spelers importeren</mat-card-title>
          <mat-card-subtitle>Upload een Excel-bestand met de ledenlijst</mat-card-subtitle>
        </mat-card-header>
        <mat-card-content style="padding-top:16px">
          <div class="import-row">
            <input #fileInput type="file" accept=".xlsx,.xls" hidden (change)="onFileSelected($event)" />
            <button mat-raised-button color="primary" (click)="fileInput.click()">
              <mat-icon>upload_file</mat-icon> Excel-bestand kiezen
            </button>
            @if (selectedFile) {
              <span style="color:#555">{{ selectedFile.name }}</span>
            }
            <button mat-raised-button color="accent" [disabled]="!selectedFile || loading()" (click)="upload()">
              {{ loading() ? 'Bezig…' : 'Importeren' }}
            </button>
          </div>
        </mat-card-content>
      </mat-card>

      <!-- Server -->
      <mat-card style="margin-top:24px">
        <mat-card-header>
          <mat-card-title>Server</mat-card-title>
        </mat-card-header>
        <mat-card-content style="padding-top:16px">
          <div class="server-meta">
            <span class="version-chip">{{ version }}</span>
            <button mat-stroked-button (click)="refreshLogs()"><mat-icon>refresh</mat-icon> Vernieuwen</button>
          </div>
          @if (logsLoading()) {
            <div style="color:#9e9e9e;font-size:13px">Laden...</div>
          }
          @if (!logsLoading()) {
            <mat-tab-group>
              <mat-tab [label]="'Routing/API (' + httpLogs.length + ')'">
                @if (httpLogs.length === 0) {
                  <div class="log-empty" style="padding:12px 0">Geen HTTP logs.</div>
                } @else {
                  <div class="log-box">
                    @for (line of httpLogs; track line) {
                      <div [style.color]="httpLogColor(line)">{{ line }}</div>
                    }
                  </div>
                }
              </mat-tab>
              <mat-tab [label]="'Debug/Error (' + errorLogs.length + ')'">
                @if (errorLogs.length === 0) {
                  <div class="log-empty" style="padding:12px 0">Geen fout-logs.</div>
                } @else {
                  <div class="log-box">
                    @for (line of errorLogs; track line) {
                      <div style="color:#f44336">{{ line }}</div>
                    }
                  </div>
                }
              </mat-tab>
              <mat-tab [label]="'Info (' + infoLogs.length + ')'">
                @if (infoLogs.length === 0) {
                  <div class="log-empty" style="padding:12px 0">Geen info-logs.</div>
                } @else {
                  <div class="log-box">
                    @for (line of infoLogs; track line) {
                      <div>{{ line }}</div>
                    }
                  </div>
                }
              </mat-tab>
            </mat-tab-group>
          }
        </mat-card-content>
      </mat-card>
    </div>
  `,
})
export class BeheerComponent implements OnInit {
  private scheduleService = inject(ScheduleService);
  private playerService = inject(PlayerService);
  private seasonService = inject(SeasonService);
  private systemService = inject(SystemService);
  private snackBar = inject(MatSnackBar);
  private dialog = inject(MatDialog);
  private cdr = inject(ChangeDetectorRef);

  readonly fileInputRef = viewChild.required<ElementRef<HTMLInputElement>>('fileInput');

  readonly renameInputRef = viewChild<ElementRef<HTMLInputElement>>('renameInput');

  seasons = signal<SeasonSummary[]>([]);
  selectedFile: File | null = null;
  loading = signal(false);
  editingSeasonId = signal('');
  renameDraft = signal('');
  regeneratingId = signal('');

  version = environment.version;
  logs = signal<string[]>([]);
  logsLoading = signal(false);

  get httpLogs(): string[] {
    return this.logs().filter((l) => l.includes('[HTTP]'));
  }

  get errorLogs(): string[] {
    return this.logs().filter((l) => l.includes('[ERROR]'));
  }

  get infoLogs(): string[] {
    return this.logs().filter((l) => !l.includes('[HTTP]') && !l.includes('[ERROR]'));
  }

  httpLogColor(line: string): string {
    // Format: "2026/04/08 13:20:21 [HTTP] GET /path 200 duration"
    const idx = line.indexOf('[HTTP]');
    const parts = idx >= 0 ? line.slice(idx).split(' ') : [];
    const status = parseInt(parts[3], 10);
    if (status >= 500) return '#f44336';
    if (status >= 400) return '#ff9800';
    if (status >= 200) return '#4caf50';
    return '#d4d4d4';
  }

  ngOnInit(): void {
    this.loadSeasons();
    this.refreshLogs();
  }

  refreshLogs(): void {
    this.logsLoading.set(true);
    this.systemService.getLogs().subscribe({
      next: ({ logs }) => {
        this.logs.set(logs);
        this.logsLoading.set(false);
      },
      error: () => {
        this.logsLoading.set(false);
      },
    });
  }

  loadSeasons(): void {
    this.scheduleService.listSeasons().subscribe({
      next: (s) => {
        this.seasons.set(s);
      },
      error: () => {},
    });
  }

  /** Extracts a readable message from an HTTP error response. */
  private errorText(err: { error?: string; message?: string }): string {
    return (typeof err.error === 'string' ? err.error.trim() : null) || err.message || 'Onbekende fout';
  }

  openGenerate(): void {
    // Check for players first so the user gets a clear explanation instead of a cryptic error.
    this.playerService.list().subscribe({
      next: (players) => {
        const participants = players.filter((p) => !p.nr.toLowerCase().includes('-s'));
        if (participants.length < 2) {
          this.snackBar.open(
            'Er zijn geen spelers geïmporteerd. Importeer eerst spelers via "Spelers importeren".',
            'OK',
            { duration: 6000 },
          );
          return;
        }
        this.openGenerateDialog();
      },
      error: () => this.openGenerateDialog(), // if the check fails, just open the dialog
    });
  }

  private openGenerateDialog(): void {
    const ref = this.dialog.open(GenerateDialogComponent, { data: null });
    ref.afterClosed().subscribe((req: GenerateScheduleRequest | undefined) => {
      if (!req) return;
      const progressPct = signal(0);
      const loadingRef = this.dialog.open(LoadingDialogComponent, {
        disableClose: true,
        data: { percent: progressPct },
      });
      const pollId = setInterval(() => {
        this.scheduleService.getProgress().subscribe((p) => progressPct.set(p.percent));
      }, 150);
      this.scheduleService.generate(req).subscribe({
        next: (s) => {
          clearInterval(pollId);
          loadingRef.close();
          this.snackBar.open('Schema gegenereerd!', 'OK', { duration: 3000 });
          this.seasonService.load(s.id);
          this.loadSeasons();
        },
        error: (err) => {
          clearInterval(pollId);
          loadingRef.close();
          if (err.status === 422) {
            this.showConstraintViolationDialog(this.errorText(err));
          } else {
            this.snackBar.open(`Fout: ${this.errorText(err)}`, 'Sluiten', { duration: 8000 });
          }
        },
      });
    });
  }

  private showConstraintViolationDialog(message: string): void {
    this.dialog.open(ConstraintViolationDialogComponent, { data: { message }, maxWidth: '520px' });
  }

  openImportSeason(): void {
    const ref = this.dialog.open(ImportSeasonDialogComponent);
    ref.afterClosed().subscribe((result: { file: File; competitionName: string; season: string } | undefined) => {
      if (!result) return;
      this.scheduleService.importSeason(result.file, result.competitionName, result.season).subscribe({
        next: (s) => {
          this.snackBar.open('Seizoen geïmporteerd!', 'OK', { duration: 3000 });
          this.seasonService.load(s.id);
          this.loadSeasons();
        },
        error: (err) => this.snackBar.open(`Fout: ${this.errorText(err)}`, 'Sluiten', { duration: 5000 }),
      });
    });
  }

  startRename(s: SeasonSummary): void {
    this.renameDraft.set(s.competitionName);
    this.editingSeasonId.set(s.id);
    setTimeout(() => this.renameInputRef()?.nativeElement.select());
  }

  saveRename(s: SeasonSummary): void {
    if (this.editingSeasonId() !== s.id) return;
    this.editingSeasonId.set('');
    const name = this.renameDraft().trim();
    if (!name || name === s.competitionName) return;
    this.scheduleService.renameSchedule(s.id, name).subscribe({
      next: () => {
        s.competitionName = name;
        this.seasonService.load();
      },
      error: (err) => this.snackBar.open(`Fout: ${this.errorText(err)}`, 'Sluiten', { duration: 5000 }),
    });
  }

  cancelRename(): void {
    this.editingSeasonId.set('');
  }

  deleteSeason(s: SeasonSummary): void {
    if (!confirm(`Seizoen "${s.competitionName}" verwijderen? Dit verwijdert ook alle avonden en wedstrijden.`)) return;
    this.scheduleService.deleteSchedule(s.id).subscribe({
      next: () => {
        this.snackBar.open('Seizoen verwijderd', 'OK', { duration: 2000 });
        this.seasonService.load();
        this.loadSeasons();
      },
      error: (err) => this.snackBar.open(`Fout: ${this.errorText(err)}`, 'Sluiten', { duration: 5000 }),
    });
  }

  setActiveSeason(s: SeasonSummary): void {
    this.scheduleService.setActive(s.id).subscribe({
      next: () => {
        this.snackBar.open(`"${s.competitionName}" is nu het actieve seizoen`, 'OK', { duration: 2500 });
        this.seasonService.load(s.id);
        this.loadSeasons();
      },
      error: (err) => this.snackBar.open(`Fout: ${this.errorText(err)}`, 'Sluiten', { duration: 5000 }),
    });
  }

  regenerateSeason(s: SeasonSummary): void {
    if (!confirm(`Schema voor "${s.competitionName}" opnieuw berekenen? De wedstrijdindeling wordt vervangen.`)) return;
    this.regeneratingId.set(s.id);
    const progressPct = signal(0);
    const loadingRef = this.dialog.open(LoadingDialogComponent, { disableClose: true, data: { percent: progressPct } });
    const pollId = setInterval(() => {
      this.scheduleService.getProgress().subscribe((p) => progressPct.set(p.percent));
    }, 500);
    this.scheduleService.regenerate(s.id).subscribe({
      next: (sched) => {
        clearInterval(pollId);
        loadingRef.close();
        this.regeneratingId.set('');
        this.snackBar.open('Schema herberekend!', 'OK', { duration: 3000 });
        this.seasonService.load(sched.id);
        this.loadSeasons();
      },
      error: (err) => {
        clearInterval(pollId);
        loadingRef.close();
        this.regeneratingId.set('');
        if (err.status === 422) {
          this.showConstraintViolationDialog(this.errorText(err));
        } else {
          this.snackBar.open(`Fout: ${this.errorText(err)}`, 'Sluiten', { duration: 5000 });
        }
      },
    });
  }

  onFileSelected(event: Event): void {
    this.selectedFile = (event.target as HTMLInputElement).files?.[0] ?? null;
    this.cdr.detectChanges();
  }

  upload(): void {
    if (!this.selectedFile) return;
    this.loading.set(true);
    this.playerService.import(this.selectedFile).subscribe({
      next: (res) => {
        this.snackBar.open(`${res.imported} spelers geïmporteerd`, 'OK', { duration: 3000 });
        this.selectedFile = null;
        this.fileInputRef().nativeElement.value = '';
        this.loading.set(false);
      },
      error: (err) => {
        this.snackBar.open(`Fout: ${err.message ?? err.statusText}`, 'Sluiten', { duration: 5000 });
        this.selectedFile = null;
        this.fileInputRef().nativeElement.value = '';
        this.loading.set(false);
      },
    });
  }
}
