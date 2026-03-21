import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Schedule, SeasonSummary, GenerateScheduleRequest, ScheduleInfo, Evening } from '../models';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class ScheduleService {
  private http = inject(HttpClient);
  private base = environment.apiBaseUrl;

  generate(req: GenerateScheduleRequest): Observable<Schedule> {
    return this.http.post<Schedule>(`${this.base}/schedule/generate`, req);
  }

  get(): Observable<Schedule> {
    return this.http.get<Schedule>(`${this.base}/schedule`);
  }

  getById(id: string): Observable<Schedule> {
    return this.http.get<Schedule>(`${this.base}/schedules/${id}`);
  }

  listSeasons(): Observable<SeasonSummary[]> {
    return this.http.get<SeasonSummary[]>(`${this.base}/schedules`);
  }

  addInhaalAvond(scheduleId: string, date: string): Observable<Schedule> {
    return this.http.post<Schedule>(`${this.base}/schedules/${scheduleId}/inhaal-avond`, { date });
  }

  renameSchedule(id: string, competitionName: string): Observable<void> {
    return this.http.patch<void>(`${this.base}/schedules/${id}`, { competitionName });
  }

  deleteSchedule(id: string): Observable<void> {
    return this.http.delete<void>(`${this.base}/schedules/${id}`);
  }

  deleteEvening(scheduleId: string, eveningId: string): Observable<void> {
    return this.http.delete<void>(`${this.base}/schedules/${scheduleId}/evenings/${eveningId}`);
  }

  getEvening(id: string): Observable<Evening> {
    return this.http.get<Evening>(`${this.base}/schedule/evening/${id}`);
  }

  getInfo(scheduleId: string): Observable<ScheduleInfo> {
    return this.http.get<ScheduleInfo>(`${this.base}/schedules/${scheduleId}/info`);
  }

  importSeason(file: File, competitionName: string, season: string): Observable<Schedule> {
    const form = new FormData();
    form.append('file', file);
    form.append('competitionName', competitionName);
    form.append('season', season);
    return this.http.post<Schedule>(`${this.base}/schedules/import-season`, form);
  }
}
