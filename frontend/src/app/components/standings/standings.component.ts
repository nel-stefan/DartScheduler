import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatSortModule, Sort } from '@angular/material/sort';
import { ScoreService } from '../../services/score.service';
import { ExportService } from '../../services/export.service';
import { PlayerStats } from '../../models';

@Component({
  selector: 'app-standings',
  standalone: true,
  imports: [CommonModule, RouterModule, MatButtonModule, MatCardModule,
            MatTableModule, MatSortModule],
  template: `
    <div style="padding:24px">
      <div style="display:flex;align-items:center;gap:16px;margin-bottom:16px">
        <button mat-button routerLink="/">← Back</button>
        <h2 style="margin:0">Standings</h2>
        <button mat-stroked-button (click)="exportService.downloadExcel()">Export Excel</button>
        <button mat-stroked-button (click)="exportService.downloadPdf()">Export PDF</button>
      </div>

      <mat-card>
        <mat-card-content>
          <table mat-table [dataSource]="stats" matSort (matSortChange)="onSort($event)" style="width:100%">
            <ng-container matColumnDef="rank">
              <th mat-header-cell *matHeaderCellDef>#</th>
              <td mat-cell *matCellDef="let s; let i = index">{{ i + 1 }}</td>
            </ng-container>
            <ng-container matColumnDef="name">
              <th mat-header-cell *matHeaderCellDef mat-sort-header="name">Player</th>
              <td mat-cell *matCellDef="let s">{{ s.player.name }}</td>
            </ng-container>
            <ng-container matColumnDef="played">
              <th mat-header-cell *matHeaderCellDef mat-sort-header="played">Played</th>
              <td mat-cell *matCellDef="let s">{{ s.played }}</td>
            </ng-container>
            <ng-container matColumnDef="wins">
              <th mat-header-cell *matHeaderCellDef mat-sort-header="wins">W</th>
              <td mat-cell *matCellDef="let s">{{ s.wins }}</td>
            </ng-container>
            <ng-container matColumnDef="draws">
              <th mat-header-cell *matHeaderCellDef>D</th>
              <td mat-cell *matCellDef="let s">{{ s.draws }}</td>
            </ng-container>
            <ng-container matColumnDef="losses">
              <th mat-header-cell *matHeaderCellDef>L</th>
              <td mat-cell *matCellDef="let s">{{ s.losses }}</td>
            </ng-container>
            <ng-container matColumnDef="pf">
              <th mat-header-cell *matHeaderCellDef>PF</th>
              <td mat-cell *matCellDef="let s">{{ s.pointsFor }}</td>
            </ng-container>
            <ng-container matColumnDef="pa">
              <th mat-header-cell *matHeaderCellDef>PA</th>
              <td mat-cell *matCellDef="let s">{{ s.pointsAgainst }}</td>
            </ng-container>
            <tr mat-header-row *matHeaderRowDef="cols"></tr>
            <tr mat-row *matRowDef="let row; columns: cols;"></tr>
          </table>
        </mat-card-content>
      </mat-card>
    </div>
  `,
})
export class StandingsComponent implements OnInit {
  private scoreService = inject(ScoreService);
  exportService = inject(ExportService);

  stats: PlayerStats[] = [];
  cols = ['rank', 'name', 'played', 'wins', 'draws', 'losses', 'pf', 'pa'];

  ngOnInit(): void {
    this.scoreService.getStats().subscribe((s) => {
      this.stats = s.sort((a, b) => b.wins - a.wins);
    });
  }

  onSort(sort: Sort): void {
    this.stats = [...this.stats].sort((a, b) => {
      const dir = sort.direction === 'asc' ? 1 : -1;
      switch (sort.active) {
        case 'name': return dir * a.player.name.localeCompare(b.player.name);
        case 'wins': return dir * (a.wins - b.wins);
        case 'played': return dir * (a.played - b.played);
        default: return 0;
      }
    });
  }
}
