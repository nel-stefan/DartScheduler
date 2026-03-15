import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { PlayerStats, DutyStats } from '../models';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class ScoreService {
  private http = inject(HttpClient);
  private base = environment.apiBaseUrl;

  submitResult(matchId: string, data: {
    leg1Winner: string; leg1Turns: number;
    leg2Winner: string; leg2Turns: number;
    leg3Winner: string; leg3Turns: number;
    reportedBy: string; rescheduleDate: string;
    secretaryNr: string; counterNr: string;
  }): Observable<void> {
    return this.http.put<void>(`${this.base}/matches/${matchId}/score`, data);
  }

  getStats(scheduleId?: string): Observable<PlayerStats[]> {
    const params = scheduleId ? `?scheduleId=${scheduleId}` : '';
    return this.http.get<PlayerStats[]>(`${this.base}/stats${params}`);
  }

  getDutyStats(scheduleId?: string): Observable<DutyStats[]> {
    const params = scheduleId ? `?scheduleId=${scheduleId}` : '';
    return this.http.get<DutyStats[]>(`${this.base}/stats/duties${params}`);
  }
}
