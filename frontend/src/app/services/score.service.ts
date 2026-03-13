import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { PlayerStats } from '../models';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class ScoreService {
  private http = inject(HttpClient);
  private base = environment.apiBaseUrl;

  submitScore(matchId: string, scoreA: number, scoreB: number): Observable<void> {
    return this.http.put<void>(`${this.base}/matches/${matchId}/score`, { scoreA, scoreB });
  }

  getStats(): Observable<PlayerStats[]> {
    return this.http.get<PlayerStats[]>(`${this.base}/stats`);
  }
}
