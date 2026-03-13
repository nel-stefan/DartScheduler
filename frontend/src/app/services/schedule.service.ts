import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Schedule, GenerateScheduleRequest } from '../models';
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
}
