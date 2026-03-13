import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, ActivatedRoute, Router } from '@angular/router';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { ScoreService } from '../../services/score.service';

@Component({
  selector: 'app-score-entry',
  standalone: true,
  imports: [CommonModule, RouterModule, ReactiveFormsModule,
            MatButtonModule, MatCardModule, MatFormFieldModule, MatInputModule, MatSnackBarModule],
  template: `
    <div style="padding:24px;max-width:480px">
      <button mat-button (click)="back()">← Back</button>
      <h2>Enter Score</h2>
      <mat-card>
        <mat-card-content style="padding-top:16px">
          <form [formGroup]="form" style="display:flex;flex-direction:column;gap:12px">
            <mat-form-field>
              <mat-label>Score Player A</mat-label>
              <input matInput type="number" formControlName="scoreA">
            </mat-form-field>
            <mat-form-field>
              <mat-label>Score Player B</mat-label>
              <input matInput type="number" formControlName="scoreB">
            </mat-form-field>
          </form>
        </mat-card-content>
        <mat-card-actions>
          <button mat-raised-button color="primary" [disabled]="form.invalid || loading" (click)="submit()">
            {{ loading ? 'Saving…' : 'Save Score' }}
          </button>
        </mat-card-actions>
      </mat-card>
    </div>
  `,
})
export class ScoreEntryComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private scoreService = inject(ScoreService);
  private snackBar = inject(MatSnackBar);
  fb = inject(FormBuilder);

  matchId = '';
  loading = false;

  form = this.fb.group({
    scoreA: [null as number | null, [Validators.required, Validators.min(0)]],
    scoreB: [null as number | null, [Validators.required, Validators.min(0)]],
  });

  ngOnInit(): void {
    this.matchId = this.route.snapshot.paramMap.get('id')!;
  }

  submit(): void {
    const { scoreA, scoreB } = this.form.value;
    if (scoreA == null || scoreB == null) return;
    this.loading = true;
    this.scoreService.submitScore(this.matchId, scoreA, scoreB).subscribe({
      next: () => {
        this.snackBar.open('Score saved!', 'OK', { duration: 2000 });
        this.back();
      },
      error: (err) => {
        this.snackBar.open(`Error: ${err.message}`, 'Close', { duration: 5000 });
        this.loading = false;
      },
    });
  }

  back(): void {
    this.router.navigate(['/']);
  }
}
