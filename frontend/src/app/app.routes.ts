import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () => import('./components/overview/overview.component').then((m) => m.OverviewComponent),
  },
  {
    path: 'spelers',
    loadComponent: () => import('./components/spelers/spelers.component').then((m) => m.SpelersComponent),
  },
  {
    path: 'evening/:id',
    loadComponent: () => import('./components/evening-view/evening-view.component').then((m) => m.EveningViewComponent),
  },
  {
    path: 'standings',
    loadComponent: () => import('./components/standings/standings.component').then((m) => m.StandingsComponent),
  },
  {
    path: 'info',
    loadComponent: () => import('./components/info/info.component').then((m) => m.InfoComponent),
  },
  {
    path: 'beheer',
    loadComponent: () => import('./components/beheer/beheer.component').then((m) => m.BeheerComponent),
  },
  {
    path: 'm',
    loadChildren: () => import('./mobile/mobile.routes').then((m) => m.mobileRoutes),
  },
  { path: '**', redirectTo: '' },
];
