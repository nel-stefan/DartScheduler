import { Component } from '@angular/core';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-mobile-shell',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive, MatIconModule],
  styles: [`
    :host { display: block; }

    .mobile-layout {
      display: flex;
      flex-direction: column;
      height: calc(100dvh - 64px);
    }

    .mobile-content {
      flex: 1;
      overflow-y: auto;
      -webkit-overflow-scrolling: touch;
    }

    .mobile-bottom-nav {
      display: flex;
      flex-shrink: 0;
      height: calc(56px + env(safe-area-inset-bottom));
      padding-bottom: env(safe-area-inset-bottom);
      background: #1b5e20;
      border-top: 1px solid rgba(0,0,0,.2);

      a {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 2px;
        color: rgba(255,255,255,.65);
        text-decoration: none;
        font-size: 11px;
        letter-spacing: .3px;
        transition: background .15s;

        mat-icon { font-size: 22px; width: 22px; height: 22px; }

        &.active {
          color: #fff;
          background: rgba(255,255,255,.12);
        }
      }
    }
  `],
  template: `
    <div class="mobile-layout">
      <div class="mobile-content">
        <router-outlet />
      </div>
      <nav class="mobile-bottom-nav">
        <a routerLink="/m/schema" routerLinkActive="active">
          <mat-icon>event_note</mat-icon>
          <span>Schema</span>
        </a>
        <a routerLink="/m/standings" routerLinkActive="active">
          <mat-icon>leaderboard</mat-icon>
          <span>Klassement</span>
        </a>
      </nav>
    </div>
  `,
})
export class MobileShellComponent {}
