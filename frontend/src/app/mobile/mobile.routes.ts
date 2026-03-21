import { Routes } from '@angular/router';

export const mobileRoutes: Routes = [
  {
    path: '',
    loadComponent: () => import('./mobile-shell.component').then(m => m.MobileShellComponent),
    children: [
      { path: '', redirectTo: 'avond', pathMatch: 'full' },
      { path: 'avond', loadComponent: () => import('./mobile-avond.component').then(m => m.MobileAvondComponent) },
      { path: 'stand', loadComponent: () => import('./mobile-stand.component').then(m => m.MobileStandComponent) },
    ],
  },
  { path: 'score/:id', loadComponent: () => import('./mobile-score.component').then(m => m.MobileScoreComponent) },
];
