import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { PlayerService } from '../../services/player.service';

@Component({
  selector: 'app-upload',
  standalone: true,
  imports: [CommonModule, MatSnackBarModule, MatButtonModule, MatIconModule, MatCardModule],
  template: `
    <mat-card>
      <mat-card-header>
        <mat-card-title>Import Players</mat-card-title>
        <mat-card-subtitle>Upload an Excel file with columns: Name, Email, Sponsor</mat-card-subtitle>
      </mat-card-header>
      <mat-card-content style="padding-top:16px">
        <input #fileInput type="file" accept=".xlsx,.xls" hidden (change)="onFileSelected($event)">
        <button mat-raised-button color="primary" (click)="fileInput.click()">
          <mat-icon>upload_file</mat-icon> Choose Excel File
        </button>
        <span *ngIf="selectedFile" style="margin-left:12px">{{ selectedFile.name }}</span>
      </mat-card-content>
      <mat-card-actions>
        <button mat-raised-button color="accent" [disabled]="!selectedFile || loading" (click)="upload()">
          {{ loading ? 'Uploading…' : 'Upload' }}
        </button>
      </mat-card-actions>
    </mat-card>
  `,
})
export class UploadComponent {
  private playerService = inject(PlayerService);
  private snackBar = inject(MatSnackBar);

  selectedFile: File | null = null;
  loading = false;

  onFileSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.selectedFile = input.files?.[0] ?? null;
  }

  upload(): void {
    if (!this.selectedFile) return;
    this.loading = true;
    this.playerService.import(this.selectedFile).subscribe({
      next: (res) => {
        this.snackBar.open(`Imported ${res.imported} players`, 'OK', { duration: 3000 });
        this.selectedFile = null;
        this.loading = false;
      },
      error: (err) => {
        this.snackBar.open(`Error: ${err.message}`, 'Close', { duration: 5000 });
        this.loading = false;
      },
    });
  }
}
