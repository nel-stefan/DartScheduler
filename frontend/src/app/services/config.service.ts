import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject } from 'rxjs';
import { environment } from '../../environments/environment';

interface AppConfig {
  appTitle: string;
  clubName: string;
  primaryColor: string;
}

@Injectable({ providedIn: 'root' })
export class ConfigService {
  private http = inject(HttpClient);

  readonly appTitle$ = new BehaviorSubject<string>('DartScheduler');
  readonly clubName$ = new BehaviorSubject<string>('');
  readonly primaryColor$ = new BehaviorSubject<string>('');

  load(): void {
    this.http.get<AppConfig>(`${environment.apiBaseUrl}/config`).subscribe({
      next: (cfg) => {
        this.appTitle$.next(cfg.appTitle);
        this.clubName$.next(cfg.clubName);
        this.primaryColor$.next(cfg.primaryColor ?? '');
      },
      error: () => {},
    });
  }
}
