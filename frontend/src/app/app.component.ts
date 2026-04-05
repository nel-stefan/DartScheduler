import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { RouterOutlet, RouterLink, RouterLinkActive, Router, NavigationEnd } from '@angular/router';
import { CommonModule, AsyncPipe } from '@angular/common';
import { Title } from '@angular/platform-browser';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { filter } from 'rxjs';
import { SeasonService } from './services/season.service';
import { ConfigService } from './services/config.service';
import { environment } from '../environments/environment';

@Component({
  selector: 'app-root',
  imports: [
    RouterOutlet,
    RouterLink,
    RouterLinkActive,
    CommonModule,
    AsyncPipe,
    MatToolbarModule,
    MatButtonModule,
    MatIconModule,
    MatSelectModule,
    MatFormFieldModule,
  ],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss',
})
export class AppComponent implements OnInit {
  protected seasonService = inject(SeasonService);
  protected configService = inject(ConfigService);
  protected version = environment.version;
  private router = inject(Router);
  private titleService = inject(Title);
  private destroyRef = inject(DestroyRef);

  isMobile = false;

  ngOnInit(): void {
    this.seasonService.load();
    this.configService.load();
    this.configService.appTitle$
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((title) => this.titleService.setTitle(title));
    this.router.events
      .pipe(
        filter((e) => e instanceof NavigationEnd),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe(() => {
        this.isMobile = this.router.url.startsWith('/m');
      });
  }
}
