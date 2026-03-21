import { Component, inject, OnInit } from '@angular/core';
import { RouterOutlet, RouterLink, RouterLinkActive, Router, NavigationEnd } from '@angular/router';
import { CommonModule } from '@angular/common';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { filter } from 'rxjs';
import { SeasonService } from './services/season.service';
import { environment } from '../environments/environment';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    RouterOutlet, RouterLink, RouterLinkActive, CommonModule,
    MatToolbarModule, MatButtonModule, MatIconModule,
    MatSelectModule, MatFormFieldModule,
  ],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss',
})
export class AppComponent implements OnInit {
  protected seasonService = inject(SeasonService);
  protected version = environment.version;
  private router = inject(Router);

  isMobile = false;

  ngOnInit(): void {
    this.seasonService.load();
    this.router.events.pipe(filter(e => e instanceof NavigationEnd)).subscribe(() => {
      this.isMobile = this.router.url.startsWith('/m');
    });
  }
}
