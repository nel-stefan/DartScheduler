import { Injectable, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';

export type Role = 'viewer' | 'maintainer' | 'admin';

export interface Identity {
  username: string;
  role: Role;
}

const TOKEN_KEY = 'dart_token';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly _identity = signal<Identity | null>(null);

  readonly identity = this._identity.asReadonly();
  readonly isLoggedIn = computed(() => this._identity() !== null);
  readonly role = computed(() => this._identity()?.role ?? null);
  readonly isAdmin = computed(() => this._identity()?.role === 'admin');
  readonly isMaintainerOrAdmin = computed(() =>
    this._identity()?.role === 'maintainer' || this._identity()?.role === 'admin'
  );

  constructor(private http: HttpClient) {}

  async init(): Promise<void> {
    try {
      const me = await firstValueFrom(this.http.get<Identity>('/api/auth/me'));
      this._identity.set(me);
    } catch {
      this._identity.set(null);
    }
  }

  async login(username: string, password: string): Promise<void> {
    const res = await firstValueFrom(
      this.http.post<{ token: string; username: string; role: Role }>('/api/auth/login', {
        username,
        password,
      })
    );
    localStorage.setItem(TOKEN_KEY, res.token);
    this._identity.set({ username: res.username, role: res.role });
  }

  logout(): void {
    localStorage.removeItem(TOKEN_KEY);
    this._identity.set(null);
  }

  getToken(): string | null {
    return localStorage.getItem(TOKEN_KEY);
  }
}
