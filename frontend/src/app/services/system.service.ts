import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class SystemService {
  private http = inject(HttpClient);
  private base = environment.apiBaseUrl;

  getLogs(): Observable<{ logs: string[] }> {
    return this.http.get<{ logs: string[] }>(`${this.base}/system/logs`);
  }
}
