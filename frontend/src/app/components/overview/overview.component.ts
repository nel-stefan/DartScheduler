import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialog, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { ScheduleService } from '../../services/schedule.service';
import { PlayerService } from '../../services/player.service';
import { Schedule, Player, GenerateScheduleRequest } from '../../models';

@Component({
  selector: 'app-generate-dialog',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatDialogModule, MatButtonModule,
            MatFormFieldModule, MatInputModule],
  template: `
    <h2 mat-dialog-title>Generate Schedule</h2>
    <mat-dialog-content>
      <form [formGroup]="form" style="display:flex;flex-direction:column;gap:12px;min-width:320px;padding-top:8px">
        <mat-form-field><mat-label>Competition Name</mat-label>
          <input matInput formControlName="competitionName">
        </mat-form-field>
        <mat-form-field><mat-label>Number of Evenings</mat-label>
          <input matInput type="number" formControlName="numEvenings">
        </mat-form-field>
        <mat-form-field><mat-label>Start Date (YYYY-MM-DD)</mat-label>
          <input matInput formControlName="startDate" placeholder="2026-04-01">
        </mat-form-field>
        <mat-form-field><mat-label>Interval (days)</mat-label>
          <input matInput type="number" formControlName="intervalDays">
        </mat-form-field>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Cancel</button>
      <button mat-raised-button color="primary" [disabled]="form.invalid" (click)="submit()">Generate</button>
    </mat-dialog-actions>
  `,
})
export class GenerateDialogComponent {
  private dialogRef = inject(MatDialogRef<GenerateDialogComponent>);
  fb = inject(FormBuilder);

  form = this.fb.group({
    competitionName: ['Liga 2026', Validators.required],
    numEvenings: [20, [Validators.required, Validators.min(1)]],
    startDate: ['2026-04-01', Validators.required],
    intervalDays: [7, [Validators.required, Validators.min(1)]],
  });

  submit(): void {
    if (this.form.valid) this.dialogRef.close(this.form.value as GenerateScheduleRequest);
  }
}

@Component({
  selector: 'app-overview',
  standalone: true,
  imports: [CommonModule, RouterModule, MatButtonModule, MatCardModule,
            MatTableModule, MatSnackBarModule, MatDialogModule],
  template: `
    <div style="padding:24px">
      <div style="display:flex;align-items:center;gap:16px;margin-bottom:24px">
        <h1 style="margin:0">{{ schedule?.competitionName ?? 'DartScheduler' }}</h1>
        <button mat-raised-button color="primary" (click)="openGenerate()">Generate Schedule</button>
        <button mat-stroked-button routerLink="/upload">Import Players</button>
        <button mat-stroked-button routerLink="/standings">Standings</button>
      </div>

      <p *ngIf="!schedule">No schedule yet. Import players then generate a schedule.</p>

      <mat-card *ngIf="schedule">
        <mat-card-content>
          <table mat-table [dataSource]="schedule.evenings" style="width:100%">
            <ng-container matColumnDef="number">
              <th mat-header-cell *matHeaderCellDef>#</th>
              <td mat-cell *matCellDef="let ev">{{ ev.number }}</td>
            </ng-container>
            <ng-container matColumnDef="date">
              <th mat-header-cell *matHeaderCellDef>Date</th>
              <td mat-cell *matCellDef="let ev">{{ ev.date | date:'mediumDate' }}</td>
            </ng-container>
            <ng-container matColumnDef="matches">
              <th mat-header-cell *matHeaderCellDef>Matches</th>
              <td mat-cell *matCellDef="let ev">{{ ev.matches?.length ?? 0 }}</td>
            </ng-container>
            <ng-container matColumnDef="actions">
              <th mat-header-cell *matHeaderCellDef></th>
              <td mat-cell *matCellDef="let ev">
                <button mat-button color="primary" [routerLink]="['/evening', ev.id]">View</button>
              </td>
            </ng-container>
            <tr mat-header-row *matHeaderRowDef="cols"></tr>
            <tr mat-row *matRowDef="let row; columns: cols;"></tr>
          </table>
        </mat-card-content>
      </mat-card>
    </div>
  `,
})
export class OverviewComponent implements OnInit {
  private scheduleService = inject(ScheduleService);
  private snackBar = inject(MatSnackBar);
  private dialog = inject(MatDialog);

  schedule: Schedule | null = null;
  players: Player[] = [];
  cols = ['number', 'date', 'matches', 'actions'];

  ngOnInit(): void {
    this.loadSchedule();
  }

  loadSchedule(): void {
    this.scheduleService.get().subscribe({
      next: (s) => (this.schedule = s),
      error: () => {},
    });
  }

  openGenerate(): void {
    const ref = this.dialog.open(GenerateDialogComponent);
    ref.afterClosed().subscribe((req: GenerateScheduleRequest | undefined) => {
      if (!req) return;
      this.scheduleService.generate(req).subscribe({
        next: (s) => {
          this.schedule = s;
          this.snackBar.open('Schedule generated!', 'OK', { duration: 3000 });
        },
        error: (err) => this.snackBar.open(`Error: ${err.message}`, 'Close', { duration: 5000 }),
      });
    });
  }
}
