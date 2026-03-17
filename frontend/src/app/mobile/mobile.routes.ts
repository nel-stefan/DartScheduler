import { Routes } from '@angular/router';
import { MobileShellComponent } from './mobile-shell.component';

export const mobileRoutes: Routes = [
  {
    path: '',
    component: MobileShellComponent,
    children: [
      { path: '', redirectTo: 'schema', pathMatch: 'full' },
      {
        path: 'schema',
        loadComponent: () =>
          import('./mobile-evenings.component').then(m => m.MobileEveningsComponent),
      },
      {
        path: 'evening/:id',
        loadComponent: () =>
          import('./mobile-matches.component').then(m => m.MobileMatchesComponent),
      },
      {
        path: 'match/:id',
        loadComponent: () =>
          import('./mobile-score.component').then(m => m.MobileScoreComponent),
      },
      {
        path: 'standings',
        loadComponent: () =>
          import('./mobile-standings.component').then(m => m.MobileStandingsComponent),
      },
    ],
  },
];
