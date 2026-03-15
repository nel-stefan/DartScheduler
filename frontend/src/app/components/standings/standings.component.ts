import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { distinctUntilChanged } from 'rxjs';
import { CommonModule } from '@angular/common';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatTabsModule } from '@angular/material/tabs';
import { MatIconModule } from '@angular/material/icon';
import { ScoreService } from '../../services/score.service';
import { SeasonService } from '../../services/season.service';
import { PlayerStats, DutyStats } from '../../models';

@Component({
  selector: 'app-standings',
  standalone: true,
  imports: [CommonModule, MatButtonModule, MatCardModule,
            MatTableModule, MatTabsModule, MatIconModule],
  styles: [`
    .section-title {
      font-size: 18px;
      font-weight: 500;
      margin: 0 0 12px 0;
    }
    table { width: 100%; }
    .rank-col { width: 40px; font-weight: 600; color: #616161; }
    .pts-col  { font-weight: 600; color: #2e7d32; }
    .highlight-row td { background: #f9fbe7; }
  `],
  template: `
    <div style="padding:24px">
      <h2 style="margin:0 0 20px 0">Klassement</h2>

      <!-- Class standings tabs -->
      <mat-tab-group animationDuration="150ms" color="primary">

        <mat-tab *ngFor="let cls of classes" [label]="cls.label">
          <mat-card style="border-radius:0 0 8px 8px;border-top:none">
            <mat-card-content>
              <table mat-table [dataSource]="cls.stats" style="width:100%">

                <ng-container matColumnDef="rank">
                  <th mat-header-cell *matHeaderCellDef class="rank-col">#</th>
                  <td mat-cell *matCellDef="let s; let i = index" class="rank-col">{{ i + 1 }}</td>
                </ng-container>

                <ng-container matColumnDef="nr">
                  <th mat-header-cell *matHeaderCellDef style="width:48px">Nr</th>
                  <td mat-cell *matCellDef="let s">{{ s.player.nr }}</td>
                </ng-container>

                <ng-container matColumnDef="name">
                  <th mat-header-cell *matHeaderCellDef>Naam</th>
                  <td mat-cell *matCellDef="let s"><strong>{{ s.player.name }}</strong></td>
                </ng-container>

                <ng-container matColumnDef="played">
                  <th mat-header-cell *matHeaderCellDef style="width:70px;text-align:center">Gespeeld</th>
                  <td mat-cell *matCellDef="let s" style="text-align:center">{{ s.played }}</td>
                </ng-container>

                <ng-container matColumnDef="wins">
                  <th mat-header-cell *matHeaderCellDef style="width:60px;text-align:center">
                    <mat-icon style="color:#2e7d32;vertical-align:middle;font-size:18px">emoji_events</mat-icon>
                  </th>
                  <td mat-cell *matCellDef="let s" style="text-align:center" class="pts-col">{{ s.wins }}</td>
                </ng-container>

                <ng-container matColumnDef="losses">
                  <th mat-header-cell *matHeaderCellDef style="width:60px;text-align:center">
                    <mat-icon style="color:#c62828;vertical-align:middle;font-size:18px">cancel</mat-icon>
                  </th>
                  <td mat-cell *matCellDef="let s" style="text-align:center;color:#c62828">{{ s.losses }}</td>
                </ng-container>

                <ng-container matColumnDef="pf">
                  <th mat-header-cell *matHeaderCellDef style="width:60px;text-align:center"
                      title="Gewonnen legs">Legs voor</th>
                  <td mat-cell *matCellDef="let s" style="text-align:center">{{ s.pointsFor }}</td>
                </ng-container>

                <ng-container matColumnDef="pa">
                  <th mat-header-cell *matHeaderCellDef style="width:60px;text-align:center"
                      title="Verloren legs">Legs tegen</th>
                  <td mat-cell *matCellDef="let s" style="text-align:center">{{ s.pointsAgainst }}</td>
                </ng-container>

                <tr mat-header-row *matHeaderRowDef="matchCols"></tr>
                <tr mat-row *matRowDef="let row; columns: matchCols;"
                    [class.highlight-row]="row.wins === cls.topWins && row.wins > 0"></tr>
              </table>

              <p *ngIf="cls.stats.length === 0" style="color:#9e9e9e;text-align:center;padding:24px 0">
                Nog geen gespeelde wedstrijden in deze klasse.
              </p>
            </mat-card-content>
          </mat-card>
        </mat-tab>

        <!-- Duty stats tab -->
        <mat-tab label="Schrijver / Teller">
          <mat-card style="border-radius:0 0 8px 8px;border-top:none">
            <mat-card-content>
              <p style="color:#616161;font-size:13px;margin:12px 0 8px 0">
                Totaal aantal keer als schrijver of teller ingezet (gecombineerd).
              </p>
              <table mat-table [dataSource]="dutyStats" style="width:100%">

                <ng-container matColumnDef="rank">
                  <th mat-header-cell *matHeaderCellDef class="rank-col">#</th>
                  <td mat-cell *matCellDef="let s; let i = index" class="rank-col">{{ i + 1 }}</td>
                </ng-container>

                <ng-container matColumnDef="nr">
                  <th mat-header-cell *matHeaderCellDef style="width:48px">Nr</th>
                  <td mat-cell *matCellDef="let s">{{ s.player.nr }}</td>
                </ng-container>

                <ng-container matColumnDef="name">
                  <th mat-header-cell *matHeaderCellDef>Naam</th>
                  <td mat-cell *matCellDef="let s"><strong>{{ s.player.name }}</strong></td>
                </ng-container>

                <ng-container matColumnDef="class">
                  <th mat-header-cell *matHeaderCellDef style="width:80px">Klasse</th>
                  <td mat-cell *matCellDef="let s">{{ s.player.class || '—' }}</td>
                </ng-container>

                <ng-container matColumnDef="count">
                  <th mat-header-cell *matHeaderCellDef style="width:80px;text-align:center">Keer</th>
                  <td mat-cell *matCellDef="let s" style="text-align:center;font-weight:600">{{ s.count }}</td>
                </ng-container>

                <tr mat-header-row *matHeaderRowDef="dutyCols"></tr>
                <tr mat-row *matRowDef="let row; columns: dutyCols;"></tr>
              </table>

              <p *ngIf="dutyStats.length === 0" style="color:#9e9e9e;text-align:center;padding:24px 0">
                Nog geen schrijvers of tellers geregistreerd.
              </p>
            </mat-card-content>
          </mat-card>
        </mat-tab>

      </mat-tab-group>
    </div>
  `,
})
export class StandingsComponent implements OnInit {
  private scoreService  = inject(ScoreService);
  private seasonService = inject(SeasonService);
  private destroyRef    = inject(DestroyRef);

  classes:   { label: string; stats: PlayerStats[]; topWins: number }[] = [];
  dutyStats: DutyStats[] = [];

  matchCols = ['rank', 'nr', 'name', 'played', 'wins', 'losses', 'pf', 'pa'];
  dutyCols  = ['rank', 'nr', 'name', 'class', 'count'];

  ngOnInit(): void {
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      distinctUntilChanged(),
    ).subscribe(() => this.loadStats());
  }

  private loadStats(): void {
    const sid = this.seasonService.selectedId$.value || undefined;
    this.scoreService.getStats(sid).subscribe((s) => {
      this.classes = this.buildClasses(s);
    });
    this.scoreService.getDutyStats(sid).subscribe((d) => {
      this.dutyStats = d.sort((a, b) => b.count - a.count);
    });
  }

  private buildClasses(allStats: PlayerStats[]): { label: string; stats: PlayerStats[]; topWins: number }[] {
    const classValues = [...new Set(allStats.map(s => s.player.class || ''))].sort();
    if (classValues.every(c => c === '')) {
      return [{ label: 'Alle spelers', stats: this.sortedStats(allStats), topWins: this.maxWins(allStats) }];
    }
    const result = classValues
      .filter(c => c !== '')
      .map(c => {
        const filtered = allStats.filter(s => (s.player.class || '') === c);
        return { label: `Klasse ${c}`, stats: this.sortedStats(filtered), topWins: this.maxWins(filtered) };
      });
    const noClass = allStats.filter(s => !s.player.class);
    if (noClass.length > 0) {
      result.push({ label: 'Overig', stats: this.sortedStats(noClass), topWins: this.maxWins(noClass) });
    }
    return result;
  }

  private sortedStats(stats: PlayerStats[]): PlayerStats[] {
    return [...stats].sort((a, b) => {
      if (b.wins !== a.wins) return b.wins - a.wins;
      const diffA = a.pointsFor - a.pointsAgainst;
      const diffB = b.pointsFor - b.pointsAgainst;
      return diffB - diffA;
    });
  }

  private maxWins(stats: PlayerStats[]): number {
    return stats.reduce((m, s) => Math.max(m, s.wins), 0);
  }
}
