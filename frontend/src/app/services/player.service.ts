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

  update(player: Player): Observable<Player> {
    return this.http.put<Player>(`${this.base}/players/${player.id}`, player);
  }

  getBuddies(playerId: string): Observable<string[]> {
    return this.http.get<string[]>(`${this.base}/players/${playerId}/buddies`);
  }

  setBuddies(playerId: string, buddyIds: string[]): Observable<void> {
    return this.http.put<void>(`${this.base}/players/${playerId}/buddies`, { buddyIds });
  }

  import(file: File): Observable<{ imported: number }> {
    const fd = new FormData();
    fd.append('file', file);
    return this.http.post<{ imported: number }>(`${this.base}/import`, fd);
  }
}
