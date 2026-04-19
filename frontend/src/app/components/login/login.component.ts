import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule, MatCardModule, MatFormFieldModule, MatInputModule, MatButtonModule],
  template: `
    <div class="login-container">
      <mat-card class="login-card">
        <mat-card-header>
          <mat-card-title>Inloggen</mat-card-title>
        </mat-card-header>
        <mat-card-content>
          <form (ngSubmit)="submit()" #f="ngForm">
            <mat-form-field appearance="outline" class="full-width">
              <mat-label>Gebruikersnaam</mat-label>
              <input matInput [(ngModel)]="username" name="username" required autocomplete="username" />
            </mat-form-field>
            <mat-form-field appearance="outline" class="full-width">
              <mat-label>Wachtwoord</mat-label>
              <input matInput type="password" [(ngModel)]="password" name="password" required autocomplete="current-password" />
            </mat-form-field>
            @if (error()) {
              <p class="error-msg">{{ error() }}</p>
            }
            <button mat-raised-button color="primary" type="submit" [disabled]="loading()">
              Inloggen
            </button>
          </form>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .login-container {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 100vh;
    }
    .login-card {
      width: 360px;
      padding: 16px;
    }
    .full-width {
      width: 100%;
      margin-bottom: 12px;
    }
    .error-msg {
      color: red;
      margin-bottom: 8px;
    }
  `],
})
export class LoginComponent {
  username = '';
  password = '';
  loading = signal(false);
  error = signal('');

  constructor(private auth: AuthService, private router: Router) {}

  async submit(): Promise<void> {
    this.error.set('');
    this.loading.set(true);
    try {
      await this.auth.login(this.username, this.password);
      this.router.navigate(['/schema']);
    } catch {
      this.error.set('Ongeldige gebruikersnaam of wachtwoord.');
    } finally {
      this.loading.set(false);
    }
  }
}
