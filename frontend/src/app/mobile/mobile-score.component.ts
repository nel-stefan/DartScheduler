import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatRadioModule } from '@angular/material/radio';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatDividerModule } from '@angular/material/divider';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { ScoreService } from '../services/score.service';
import { Match, Player } from '../models';

@Component({
  selector: 'app-mobile-score',
  standalone: true,
  imports: [
    CommonModule, ReactiveFormsModule,
    MatButtonModule, MatIconModule, MatRadioModule,
    MatSelectModule, MatFormFieldModule, MatInputModule,
    MatDividerModule, MatSnackBarModule,
  ],
  styles: [`
    .page-header {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 8px 8px 4px;
      background: #fff;
      border-bottom: 1px solid rgba(0,0,0,.1);
      position: sticky;
      top: 0;
      z-index: 10;
    }

    .header-title {
      font-size: 16px;
      font-weight: 500;
      color: #212121;
    }

    .header-matchup {
      font-size: 12px;
      color: #757575;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .form-body { padding: 0 0 16px; }

    .section-label {
      font-size: 11px;
      font-weight: 600;
      letter-spacing: .6px;
      text-transform: uppercase;
      color: #757575;
      padding: 14px 16px 4px;
    }

    .leg-section {
      padding: 8px 16px 12px;
      background: #fff;
      margin-bottom: 2px;
    }

    .radio-group {
      display: flex;
      flex-direction: column;
      gap: 0;
    }

    .radio-option {
      display: flex;
      align-items: center;
      min-height: 48px;
      border-bottom: 1px solid rgba(0,0,0,.06);

      &:last-child { border-bottom: none; }
    }

    .turns-row {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-top: 8px;

      label { font-size: 13px; color: #616161; }

      mat-form-field { flex: 1; }
    }

    .stats-section {
      padding: 8px 16px 12px;
      background: #fff;
      margin-bottom: 2px;
    }

    .stats-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 8px;
    }

    .admin-section {
      padding: 8px 16px 12px;
      background: #fff;
      margin-bottom: 2px;
    }

    .admin-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 8px;
    }

    .actions {
      display: flex;
      gap: 12px;
      padding: 16px;
      background: #fff;
      border-top: 1px solid rgba(0,0,0,.08);
    }

    .no-state {
      text-align: center;
      color: #9e9e9e;
      padding: 40px 16px;
      font-size: 14px;
    }
  `],
  template: `
    <ng-container *ngIf="!match">
      <div class="no-state">Geen wedstrijd geselecteerd.</div>
    </ng-container>

    <ng-container *ngIf="match">
      <div class="page-header">
        <button mat-icon-button (click)="back()">
          <mat-icon>arrow_back</mat-icon>
        </button>
        <div>
          <div class="header-title">Score invoeren</div>
          <div class="header-matchup">{{ nameOf(match.playerA) }} vs {{ nameOf(match.playerB) }}</div>
        </div>
      </div>

      <form [formGroup]="form" class="form-body">

        <!-- Leg 1 -->
        <div class="section-label">Leg 1</div>
        <div class="leg-section">
          <mat-radio-group formControlName="leg1Winner" class="radio-group">
            <div class="radio-option">
              <mat-radio-button [value]="match.playerA">{{ nameOf(match.playerA) }}</mat-radio-button>
            </div>
            <div class="radio-option">
              <mat-radio-button [value]="match.playerB">{{ nameOf(match.playerB) }}</mat-radio-button>
            </div>
          </mat-radio-group>
          <div class="turns-row">
            <label>Beurten:</label>
            <mat-form-field subscriptSizing="dynamic">
              <mat-select formControlName="leg1Turns">
                <mat-option [value]="null">—</mat-option>
                <mat-option *ngFor="let t of turnsOptions" [value]="t">{{ t }}</mat-option>
              </mat-select>
            </mat-form-field>
          </div>
        </div>

        <!-- Leg 2 -->
        <div class="section-label">Leg 2</div>
        <div class="leg-section">
          <mat-radio-group formControlName="leg2Winner" class="radio-group">
            <div class="radio-option">
              <mat-radio-button [value]="match.playerA">{{ nameOf(match.playerA) }}</mat-radio-button>
            </div>
            <div class="radio-option">
              <mat-radio-button [value]="match.playerB">{{ nameOf(match.playerB) }}</mat-radio-button>
            </div>
          </mat-radio-group>
          <div class="turns-row">
            <label>Beurten:</label>
            <mat-form-field subscriptSizing="dynamic">
              <mat-select formControlName="leg2Turns">
                <mat-option [value]="null">—</mat-option>
                <mat-option *ngFor="let t of turnsOptions" [value]="t">{{ t }}</mat-option>
              </mat-select>
            </mat-form-field>
          </div>
        </div>

        <!-- Leg 3 (only when 1-1) -->
        <ng-container *ngIf="needsLeg3">
          <div class="section-label">Leg 3</div>
          <div class="leg-section">
            <mat-radio-group formControlName="leg3Winner" class="radio-group">
              <div class="radio-option">
                <mat-radio-button [value]="match.playerA">{{ nameOf(match.playerA) }}</mat-radio-button>
              </div>
              <div class="radio-option">
                <mat-radio-button [value]="match.playerB">{{ nameOf(match.playerB) }}</mat-radio-button>
              </div>
            </mat-radio-group>
            <div class="turns-row">
              <label>Beurten:</label>
              <mat-form-field subscriptSizing="dynamic">
                <mat-select formControlName="leg3Turns">
                  <mat-option [value]="null">—</mat-option>
                  <mat-option *ngFor="let t of turnsOptions" [value]="t">{{ t }}</mat-option>
                </mat-select>
              </mat-form-field>
            </div>
          </div>
        </ng-container>

        <!-- Stats -->
        <div class="section-label">Statistieken</div>
        <div class="stats-section">
          <div class="stats-grid">
            <mat-form-field subscriptSizing="dynamic">
              <mat-label>180's {{ shortName(match.playerA) }}</mat-label>
              <input matInput type="number" min="0" formControlName="playerA180s">
            </mat-form-field>
            <mat-form-field subscriptSizing="dynamic">
              <mat-label>180's {{ shortName(match.playerB) }}</mat-label>
              <input matInput type="number" min="0" formControlName="playerB180s">
            </mat-form-field>
            <mat-form-field subscriptSizing="dynamic">
              <mat-label>H.Finish {{ shortName(match.playerA) }}</mat-label>
              <input matInput type="number" min="0" formControlName="playerAHighestFinish">
            </mat-form-field>
            <mat-form-field subscriptSizing="dynamic">
              <mat-label>H.Finish {{ shortName(match.playerB) }}</mat-label>
              <input matInput type="number" min="0" formControlName="playerBHighestFinish">
            </mat-form-field>
          </div>
        </div>

        <!-- Admin -->
        <div class="section-label">Administratie</div>
        <div class="admin-section">
          <div class="admin-grid">
            <mat-form-field subscriptSizing="dynamic">
              <mat-label>Schrijver</mat-label>
              <mat-select formControlName="secretaryNr">
                <mat-option value="">—</mat-option>
                <mat-option *ngFor="let p of players" [value]="p.nr">{{ p.nr }} {{ p.name }}</mat-option>
              </mat-select>
            </mat-form-field>
            <mat-form-field subscriptSizing="dynamic">
              <mat-label>Teller</mat-label>
              <mat-select formControlName="counterNr">
                <mat-option value="">—</mat-option>
                <mat-option *ngFor="let p of players" [value]="p.nr">{{ p.nr }} {{ p.name }}</mat-option>
              </mat-select>
            </mat-form-field>
          </div>
        </div>

      </form>

      <div class="actions">
        <button mat-stroked-button (click)="back()" style="flex:1">Annuleren</button>
        <button mat-raised-button color="primary" (click)="submit()"
          [disabled]="!isValid() || saving" style="flex:2">
          {{ saving ? 'Opslaan...' : 'Opslaan' }}
        </button>
      </div>
    </ng-container>
  `,
})
export class MobileScoreComponent implements OnInit {
  private router       = inject(Router);
  private fb           = inject(FormBuilder);
  private scoreService = inject(ScoreService);
  private snackBar     = inject(MatSnackBar);

  match:    Match | null = null;
  eveningId = '';
  players:  Player[] = [];
  saving = false;

  turnsOptions = Array.from({ length: 18 }, (_, i) => i + 3);

  form = this.fb.group({
    leg1Winner: ['', Validators.required],
    leg1Turns:  [null as number | null],
    leg2Winner: ['', Validators.required],
    leg2Turns:  [null as number | null],
    leg3Winner: [''],
    leg3Turns:  [null as number | null],
    playerA180s:          [0],
    playerB180s:          [0],
    playerAHighestFinish: [0],
    playerBHighestFinish: [0],
    secretaryNr: [''],
    counterNr:   [''],
  });

  private playerMap = new Map<string, Player>();

  ngOnInit(): void {
    const state = history.state;
    if (!state?.match) {
      this.router.navigate(['/m/schema']);
      return;
    }
    this.match     = state.match as Match;
    this.eveningId = state.eveningId as string;
    this.players   = (state.players as Player[]) ?? [];
    for (const p of this.players) this.playerMap.set(p.id, p);

    const m = this.match;
    this.form.patchValue({
      leg1Winner: m.leg1Winner || '',
      leg1Turns:  m.leg1Turns  || null,
      leg2Winner: m.leg2Winner || '',
      leg2Turns:  m.leg2Turns  || null,
      leg3Winner: m.leg3Winner || '',
      leg3Turns:  m.leg3Turns  || null,
      playerA180s:          m.playerA180s          ?? 0,
      playerB180s:          m.playerB180s          ?? 0,
      playerAHighestFinish: m.playerAHighestFinish ?? 0,
      playerBHighestFinish: m.playerBHighestFinish ?? 0,
      secretaryNr: m.secretaryNr || '',
      counterNr:   m.counterNr   || '',
    });
  }

  get needsLeg3(): boolean {
    const v = this.form.value;
    return !!(v.leg1Winner && v.leg2Winner && v.leg1Winner !== v.leg2Winner);
  }

  isValid(): boolean {
    const v = this.form.value;
    if (!v.leg1Winner || !v.leg2Winner) return false;
    if (this.needsLeg3 && !v.leg3Winner) return false;
    return true;
  }

  nameOf(id: string): string {
    const p = this.playerMap.get(id);
    if (!p) return '—';
    const parts = p.name.split(', ');
    const first = parts.length === 2 ? parts[1].split(' ')[0] : p.name;
    return `${first} - ${p.nr}`;
  }

  shortName(id: string): string {
    const p = this.playerMap.get(id);
    if (!p) return '?';
    const parts = p.name.split(', ');
    return parts.length === 2 ? parts[1].split(' ')[0] : p.name;
  }

  submit(): void {
    if (!this.isValid() || !this.match) return;
    const v = this.form.value;
    this.saving = true;
    this.scoreService.submitResult(this.match.id, {
      leg1Winner: v.leg1Winner!,
      leg1Turns:  v.leg1Turns  ?? 0,
      leg2Winner: v.leg2Winner!,
      leg2Turns:  v.leg2Turns  ?? 0,
      leg3Winner: v.leg3Winner ?? '',
      leg3Turns:  v.leg3Turns  ?? 0,
      playerA180s:          v.playerA180s          ?? 0,
      playerB180s:          v.playerB180s          ?? 0,
      playerAHighestFinish: v.playerAHighestFinish ?? 0,
      playerBHighestFinish: v.playerBHighestFinish ?? 0,
      reportedBy:    '',
      rescheduleDate: '',
      secretaryNr: v.secretaryNr ?? '',
      counterNr:   v.counterNr   ?? '',
    }).subscribe({
      next: () => {
        this.saving = false;
        this.snackBar.open('Resultaat opgeslagen!', 'OK', { duration: 2500 });
        this.router.navigate(['/m/evening', this.eveningId]);
      },
      error: () => {
        this.saving = false;
        this.snackBar.open('Fout bij opslaan.', 'Sluiten', { duration: 4000 });
      },
    });
  }

  back(): void {
    if (this.eveningId) {
      this.router.navigate(['/m/evening', this.eveningId]);
    } else {
      this.router.navigate(['/m/schema']);
    }
  }
}
