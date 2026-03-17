import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { Router } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { distinctUntilChanged, filter } from 'rxjs';
import { MatListModule } from '@angular/material/list';
import { MatIconModule } from '@angular/material/icon';
import { MatDividerModule } from '@angular/material/divider';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { SeasonService } from '../services/season.service';
import { ScheduleService } from '../services/schedule.service';
import { Evening, Schedule } from '../models';

@Component({
  selector: 'app-mobile-evenings',
  standalone: true,
  imports: [CommonModule, DatePipe, MatListModule, MatIconModule, MatDividerModule, MatProgressSpinnerModule],
  styles: [`
    .header {
      padding: 16px 16px 8px;
      font-size: 13px;
      color: #616161;
      font-weight: 500;
      letter-spacing: .3px;
      text-transform: uppercase;
    }

    .evening-item {
      cursor: pointer;
      padding: 12px 16px;
      display: flex;
      align-items: center;
      gap: 12px;
      min-height: 64px;
      border-bottom: 1px solid rgba(0,0,0,.08);
      transition: background .12s;

      &:active { background: rgba(0,0,0,.05); }
    }

    .evening-info { flex: 1; min-width: 0; }

    .evening-title {
      font-size: 15px;
      font-weight: 500;
      color: #212121;
      display: flex;
      align-items: center;
      gap: 6px;
    }

    .evening-date {
      font-size: 13px;
      color: #757575;
      margin-top: 2px;
    }

    .status-badge {
      font-size: 12px;
      font-weight: 600;
      padding: 3px 8px;
      border-radius: 12px;
      white-space: nowrap;

      &.status-done    { background: #e8f5e9; color: #2e7d32; }
      &.status-partial { background: #fff3e0; color: #e65100; }
      &.status-open    { background: #f5f5f5; color: #757575; }
      &.status-inhaal  { background: #f3e5f5; color: #7b1fa2; }
    }

    .chevron { color: #bdbdbd; }

    .spinner-wrap { display: flex; justify-content: center; padding: 40px; }

    .empty {
      text-align: center;
      color: #9e9e9e;
      padding: 40px 16px;
      font-size: 14px;
    }

    .inhaal-icon { font-size: 16px; width: 16px; height: 16px; vertical-align: middle; color: #7b1fa2; }
  `],
  template: `
    <div *ngIf="loading" class="spinner-wrap">
      <mat-spinner diameter="40" />
    </div>

    <ng-container *ngIf="!loading">
      <div class="header" *ngIf="schedule">{{ schedule.competitionName }} — {{ schedule.season }}</div>

      <div *ngIf="evenings.length === 0" class="empty">Geen avonden gevonden.</div>

      <div *ngFor="let ev of evenings" class="evening-item" (click)="go(ev)">
        <div class="evening-info">
          <div class="evening-title">
            <mat-icon class="inhaal-icon" *ngIf="ev.isInhaalAvond">replay</mat-icon>
            <span>{{ ev.isInhaalAvond ? 'Inhaalavond' : 'Avond ' + ev.number }}</span>
          </div>
          <div class="evening-date">{{ ev.date | date:'EEEE d MMMM yyyy' }}</div>
        </div>
        <span class="status-badge" [ngClass]="statusClass(ev)">{{ statusLabel(ev) }}</span>
        <mat-icon class="chevron">chevron_right</mat-icon>
      </div>
    </ng-container>
  `,
})
export class MobileEveningsComponent implements OnInit {
  private scheduleService = inject(ScheduleService);
  private seasonService   = inject(SeasonService);
  private router          = inject(Router);
  private destroyRef      = inject(DestroyRef);

  schedule: Schedule | null = null;
  evenings: Evening[] = [];
  loading = false;

  ngOnInit(): void {
    this.seasonService.selectedId$.pipe(
      takeUntilDestroyed(this.destroyRef),
      filter(id => !!id),
      distinctUntilChanged(),
    ).subscribe(id => this.load(id!));
  }

  private load(id: string): void {
    this.loading = true;
    this.scheduleService.getById(id).subscribe({
      next: (s) => {
        this.schedule = s;
        this.evenings = s.evenings ?? [];
        this.loading = false;
      },
      error: () => { this.loading = false; },
    });
  }

  go(ev: Evening): void {
    this.router.navigate(['/m/evening', ev.id]);
  }

  statusLabel(ev: Evening): string {
    if (ev.isInhaalAvond) return 'Inhaal';
    const played = ev.matches?.filter(m => m.played).length ?? 0;
    const total  = ev.matches?.length ?? 0;
    if (total === 0) return '—';
    if (played === total) return 'Volledig';
    if (played === 0) return 'Open';
    return `${played}/${total}`;
  }

  statusClass(ev: Evening): string {
    if (ev.isInhaalAvond) return 'status-inhaal';
    const played = ev.matches?.filter(m => m.played).length ?? 0;
    const total  = ev.matches?.length ?? 0;
    if (total > 0 && played === total) return 'status-done';
    if (played > 0) return 'status-partial';
    return 'status-open';
  }
}
