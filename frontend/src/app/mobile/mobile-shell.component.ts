import { Component } from '@angular/core';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';

@Component({
    selector: 'app-mobile-shell',
    imports: [RouterOutlet, RouterLink, RouterLinkActive, MatIconModule],
    styles: [`
    :host { display: flex; flex-direction: column; height: 100dvh; overflow: hidden; }
    .mobile-content { flex: 1; overflow-y: auto; -webkit-overflow-scrolling: touch; background: #f5f0ee; }
    .mobile-nav {
      display: flex; flex-shrink: 0;
      height: calc(60px + env(safe-area-inset-bottom));
      padding-bottom: env(safe-area-inset-bottom);
      background: #4e342e;
      border-top: 1px solid rgba(0,0,0,.2);
      a {
        flex: 1; display: flex; flex-direction: column; align-items: center;
        justify-content: center; gap: 2px; color: rgba(255,255,255,.6);
        text-decoration: none; font-size: 11px; letter-spacing: .3px;
        mat-icon { font-size: 24px; width: 24px; height: 24px; }
        &.active { color: #fff; background: rgba(255,255,255,.12); }
      }
    }
  `],
    template: `
    <div class="mobile-content"><router-outlet /></div>
    <nav class="mobile-nav">
      <a routerLink="avond" routerLinkActive="active"><mat-icon>sports_bar</mat-icon>Avond</a>
      <a routerLink="stand" routerLinkActive="active"><mat-icon>leaderboard</mat-icon>Stand</a>
    </nav>
  `
})
export class MobileShellComponent {}
