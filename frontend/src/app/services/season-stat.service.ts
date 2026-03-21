import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { EveningPlayerStat } from '../models';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class SeasonStatService {
  private http = inject(HttpClient);
  private base = environment.apiBaseUrl;

  getBySchedule(scheduleId: string): Observable<EveningPlayerStat[]> {
    return this.http.get<EveningPlayerStat[]>(`${this.base}/schedules/${scheduleId}/player-stats`);
  }

  upsert(scheduleId: string, playerId: string, oneEighties: number, highestFinish: number): Observable<void> {
    return this.http.put<void>(`${this.base}/schedules/${scheduleId}/player-stats/${playerId}`, { oneEighties, highestFinish });
  }
}
