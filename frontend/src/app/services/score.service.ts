import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { PlayerStats, DutyStats } from '../models';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class ScoreService {
  private http = inject(HttpClient);
  private base = environment.apiBaseUrl;

  submitResult(
    matchId: string,
    data: {
      leg1Winner: string;
      leg1Turns: number;
      leg2Winner: string;
      leg2Turns: number;
      leg3Winner: string;
      leg3Turns: number;
      playerA180s: number;
      playerB180s: number;
      playerAHighestFinish: number;
      playerBHighestFinish: number;
      reportedBy: string;
      rescheduleDate: string;
      secretaryNr: string;
      counterNr: string;
      playedDate: string;
    }
  ): Observable<void> {
    return this.http.put<void>(`${this.base}/matches/${matchId}/score`, data);
  }

  reportAbsent(eveningId: string, playerId: string, reportedBy: string): Observable<void> {
    return this.http.post<void>(`${this.base}/evenings/${eveningId}/report-absent`, { playerId, reportedBy });
  }

  getStats(scheduleId?: string, listId?: string | null): Observable<PlayerStats[]> {
    let params = new HttpParams();
    if (scheduleId) params = params.set('scheduleId', scheduleId);
    if (listId) params = params.set('listId', listId);
    return this.http.get<PlayerStats[]>(`${this.base}/stats`, { params });
  }

  getDutyStats(scheduleId?: string, listId?: string | null): Observable<DutyStats[]> {
    let params = new HttpParams();
    if (scheduleId) params = params.set('scheduleId', scheduleId);
    if (listId) params = params.set('listId', listId);
    return this.http.get<DutyStats[]>(`${this.base}/stats/duties`, { params });
  }
}
