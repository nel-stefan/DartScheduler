import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Player } from '../models';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class PlayerService {
  private http = inject(HttpClient);
  private base = environment.apiBaseUrl;

  list(): Observable<Player[]> {
    return this.http.get<Player[]>(`${this.base}/players`);
  }

  import(file: File): Observable<{ imported: number }> {
    const fd = new FormData();
    fd.append('file', file);
    return this.http.post<{ imported: number }>(`${this.base}/import`, fd);
  }
}
