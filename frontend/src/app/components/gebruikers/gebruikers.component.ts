import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatCardModule } from '@angular/material/card';
import { MatDialogModule, MatDialog } from '@angular/material/dialog';

interface User {
  id: string;
  username: string;
  role: string;
  createdAt: string;
}

@Component({
  selector: 'app-gebruikers',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatCardModule,
    MatDialogModule,
  ],
  template: `
    <div class="page-container">
      <h2>Gebruikersbeheer</h2>

      <mat-card class="form-card">
        <mat-card-header><mat-card-title>Nieuwe gebruiker</mat-card-title></mat-card-header>
        <mat-card-content>
          <form (ngSubmit)="createUser()" class="user-form">
            <mat-form-field appearance="outline">
              <mat-label>Gebruikersnaam</mat-label>
              <input matInput [(ngModel)]="newUsername" name="username" required />
            </mat-form-field>
            <mat-form-field appearance="outline">
              <mat-label>Wachtwoord</mat-label>
              <input matInput type="password" [(ngModel)]="newPassword" name="password" required />
            </mat-form-field>
            <mat-form-field appearance="outline">
              <mat-label>Rol</mat-label>
              <mat-select [(ngModel)]="newRole" name="role" required>
                <mat-option value="viewer">Viewer</mat-option>
                <mat-option value="maintainer">Maintainer</mat-option>
                <mat-option value="admin">Admin</mat-option>
              </mat-select>
            </mat-form-field>
            <button mat-raised-button color="primary" type="submit">Aanmaken</button>
          </form>
          @if (createError()) {
            <p class="error-msg">{{ createError() }}</p>
          }
        </mat-card-content>
      </mat-card>

      <table mat-table [dataSource]="users()" class="user-table">
        <ng-container matColumnDef="username">
          <th mat-header-cell *matHeaderCellDef>Gebruikersnaam</th>
          <td mat-cell *matCellDef="let u">{{ u.username }}</td>
        </ng-container>
        <ng-container matColumnDef="role">
          <th mat-header-cell *matHeaderCellDef>Rol</th>
          <td mat-cell *matCellDef="let u">
            <mat-select [value]="u.role" (selectionChange)="updateRole(u, $event.value)">
              <mat-option value="viewer">Viewer</mat-option>
              <mat-option value="maintainer">Maintainer</mat-option>
              <mat-option value="admin">Admin</mat-option>
            </mat-select>
          </td>
        </ng-container>
        <ng-container matColumnDef="password">
          <th mat-header-cell *matHeaderCellDef>Wachtwoord</th>
          <td mat-cell *matCellDef="let u">
            <input type="password" placeholder="Nieuw wachtwoord" #pwInput class="pw-input" />
            <button mat-icon-button (click)="updatePassword(u, pwInput.value); pwInput.value = ''" title="Opslaan">
              <mat-icon>save</mat-icon>
            </button>
          </td>
        </ng-container>
        <ng-container matColumnDef="actions">
          <th mat-header-cell *matHeaderCellDef></th>
          <td mat-cell *matCellDef="let u">
            <button mat-icon-button color="warn" (click)="deleteUser(u)" title="Verwijderen">
              <mat-icon>delete</mat-icon>
            </button>
          </td>
        </ng-container>
        <tr mat-header-row *matHeaderRowDef="columns"></tr>
        <tr mat-row *matRowDef="let row; columns: columns;"></tr>
      </table>
    </div>
  `,
  styles: [`
    .page-container { padding: 24px; max-width: 900px; margin: 0 auto; }
    .form-card { margin-bottom: 24px; }
    .user-form { display: flex; gap: 12px; flex-wrap: wrap; align-items: center; }
    .user-table { width: 100%; }
    .error-msg { color: red; margin-top: 8px; }
    .pw-input { border: 1px solid #ccc; border-radius: 4px; padding: 4px 8px; font-size: 14px; }
  `],
})
export class GebruikersComponent implements OnInit {
  users = signal<User[]>([]);
  createError = signal('');
  columns = ['username', 'role', 'password', 'actions'];

  newUsername = '';
  newPassword = '';
  newRole = 'viewer';

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    this.load();
  }

  private async load(): Promise<void> {
    const users = await firstValueFrom(this.http.get<User[]>('/api/users'));
    this.users.set(users);
  }

  async createUser(): Promise<void> {
    this.createError.set('');
    try {
      await firstValueFrom(
        this.http.post('/api/users', {
          username: this.newUsername,
          password: this.newPassword,
          role: this.newRole,
        })
      );
      this.newUsername = '';
      this.newPassword = '';
      this.newRole = 'viewer';
      await this.load();
    } catch (e: any) {
      this.createError.set(e?.error || 'Fout bij aanmaken gebruiker.');
    }
  }

  async updateRole(user: User, role: string): Promise<void> {
    await firstValueFrom(this.http.put(`/api/users/${user.id}`, { role }));
    await this.load();
  }

  async updatePassword(user: User, password: string): Promise<void> {
    if (!password) return;
    await firstValueFrom(this.http.put(`/api/users/${user.id}`, { password }));
  }

  async deleteUser(user: User): Promise<void> {
    if (!confirm(`Gebruiker "${user.username}" verwijderen?`)) return;
    await firstValueFrom(this.http.delete(`/api/users/${user.id}`));
    await this.load();
  }
}
