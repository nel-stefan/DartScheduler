import { Component, inject, OnInit, signal } from '@angular/core';

import { FormsModule } from '@angular/forms';
import { MatDialogModule, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { MatInputModule } from '@angular/material/input';
import { SeasonStatService } from '../services/season-stat.service';
import { displayName } from '../utils/display-name';

export interface EveningStatDialogData {
  scheduleId: string;
  players: { id: string; name: string }[];
  /** Pre-select a player (standings use-case). */
  preselectedPlayerId?: string;
}

@Component({
  selector: 'app-evening-stat-dialog',
  imports: [FormsModule, MatDialogModule, MatButtonModule, MatFormFieldModule, MatSelectModule, MatInputModule],
  styles: [
    `
      .fields {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 8px;
        margin-top: 8px;
      }
    `,
  ],
  template: `
    <h2 mat-dialog-title>180s / Hoge Finish — {{ playerLabel }}</h2>
    <mat-dialog-content style="min-width:320px;padding-top:8px">
      @if (!data.preselectedPlayerId) {
        <mat-form-field style="width:100%" subscriptSizing="dynamic">
          <mat-label>Speler</mat-label>
          <mat-select [ngModel]="selectedPlayerId()" (ngModelChange)="selectedPlayerId.set($event); loadStat()">
            @for (p of data.players; track p) {
              <mat-option [value]="p.id">
                {{ fmt(p.name) }}
              </mat-option>
            }
          </mat-select>
        </mat-form-field>
      }

      @if (selectedPlayerId()) {
        <div class="fields">
          <mat-form-field subscriptSizing="dynamic">
            <mat-label>180s</mat-label>
            <input matInput type="number" [ngModel]="oneEighties()" (ngModelChange)="oneEighties.set($event)" min="0" />
          </mat-form-field>
          <mat-form-field subscriptSizing="dynamic">
            <mat-label>Hoogste Finish</mat-label>
            <input
              matInput
              type="number"
              [ngModel]="highestFinish()"
              (ngModelChange)="highestFinish.set($event)"
              min="0"
              max="170"
            />
          </mat-form-field>
        </div>
      }
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annuleren</button>
      <button mat-flat-button color="primary" [disabled]="saving() || !selectedPlayerId()" (click)="save()">
        {{ saving() ? 'Opslaan…' : 'Opslaan' }}
      </button>
    </mat-dialog-actions>
  `,
})
export class EveningStatDialogComponent implements OnInit {
  data = inject<EveningStatDialogData>(MAT_DIALOG_DATA);
  dialogRef = inject(MatDialogRef<EveningStatDialogComponent>);
  private svc = inject(SeasonStatService);

  selectedPlayerId = signal(this.data.preselectedPlayerId ?? '');
  oneEighties = signal(0);
  highestFinish = signal(0);
  saving = signal(false);

  fmt = displayName;

  get playerLabel(): string {
    if (!this.selectedPlayerId()) return '';
    const p = this.data.players.find((x) => x.id === this.selectedPlayerId());
    return p ? displayName(p.name) : '';
  }

  ngOnInit(): void {
    if (this.selectedPlayerId()) this.loadStat();
  }

  loadStat(): void {
    if (!this.selectedPlayerId()) return;
    this.svc.getBySchedule(this.data.scheduleId).subscribe((stats) => {
      const s = stats.find((x) => x.playerId === this.selectedPlayerId());
      this.oneEighties.set(s?.oneEighties ?? 0);
      this.highestFinish.set(s?.highestFinish ?? 0);
    });
  }

  save(): void {
    this.saving.set(true);
    this.svc.upsert(this.data.scheduleId, this.selectedPlayerId(), this.oneEighties(), this.highestFinish()).subscribe({
      next: () => {
        this.saving.set(false);
        this.dialogRef.close(true);
      },
      error: () => {
        this.saving.set(false);
      },
    });
  }
}
