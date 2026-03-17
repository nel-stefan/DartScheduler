import { Component, inject, DestroyRef, OnInit } from '@angular/core';
import { RouterOutlet, RouterLink, RouterLinkActive, Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { BreakpointObserver } from '@angular/cdk/layout';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
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

  private breakpoints = inject(BreakpointObserver);
  private router      = inject(Router);
  private destroyRef  = inject(DestroyRef);

  ngOnInit(): void {
    this.seasonService.load();
    this.breakpoints.observe('(max-width: 767px)')
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(state => {
        const onMobile = this.router.url.startsWith('/m');
        if (state.matches && !onMobile) this.router.navigate(['/m']);
        if (!state.matches && onMobile) this.router.navigate(['/']);
      });
  }
}
