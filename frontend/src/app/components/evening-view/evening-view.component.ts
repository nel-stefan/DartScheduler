import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, ActivatedRoute } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { HttpClient } from '@angular/common/http';
import { Evening, Player } from '../../models';
import { PlayerService } from '../../services/player.service';
import { environment } from '../../../environments/environment';

@Component({
  selector: 'app-evening-view',
  standalone: true,
  imports: [CommonModule, RouterModule, MatButtonModule, MatCardModule, MatTableModule],
  template: `
    <div style="padding:24px">
      <button mat-button routerLink="/">← Back</button>
      <h2>Evening {{ evening?.number }} – {{ evening?.date | date:'mediumDate' }}</h2>

      <mat-card *ngIf="evening">
        <mat-card-content>
          <table mat-table [dataSource]="evening.matches" style="width:100%">
            <ng-container matColumnDef="playerA">
              <th mat-header-cell *matHeaderCellDef>Player A</th>
              <td mat-cell *matCellDef="let m">{{ playerName(m.playerA) }}</td>
            </ng-container>
            <ng-container matColumnDef="playerB">
              <th mat-header-cell *matHeaderCellDef>Player B</th>
              <td mat-cell *matCellDef="let m">{{ playerName(m.playerB) }}</td>
            </ng-container>
            <ng-container matColumnDef="score">
              <th mat-header-cell *matHeaderCellDef>Score</th>
              <td mat-cell *matCellDef="let m">
                {{ m.played ? (m.scoreA + ' – ' + m.scoreB) : '—' }}
              </td>
            </ng-container>
            <ng-container matColumnDef="actions">
              <th mat-header-cell *matHeaderCellDef></th>
              <td mat-cell *matCellDef="let m">
                <button mat-button color="primary" *ngIf="!m.played"
                  [routerLink]="['/score', m.id]">Enter Score</button>
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
export class EveningViewComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private http = inject(HttpClient);
  private playerService = inject(PlayerService);

  evening: Evening | null = null;
  players: Player[] = [];
  cols = ['playerA', 'playerB', 'score', 'actions'];

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id')!;
    this.http.get<Evening>(`${environment.apiBaseUrl}/schedule/evening/${id}`).subscribe((ev) => {
      this.evening = ev;
    });
    this.playerService.list().subscribe((ps) => (this.players = ps));
  }

  playerName(id: string): string {
    return this.players.find((p) => p.id === id)?.name ?? id.slice(0, 8);
  }
}
