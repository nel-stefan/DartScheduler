import { Routes } from '@angular/router';
import { authGuard } from './guards/auth.guard';
import { roleGuard } from './guards/role.guard';

export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () => import('./components/login/login.component').then((m) => m.LoginComponent),
  },
  {
    path: '',
    redirectTo: 'schema',
    pathMatch: 'full',
  },
  {
    path: 'schema',
    canActivate: [authGuard],
    loadComponent: () => import('./components/overview/overview.component').then((m) => m.OverviewComponent),
  },
  {
    path: 'spelers',
    canActivate: [authGuard],
    loadComponent: () => import('./components/spelers/spelers.component').then((m) => m.SpelersComponent),
  },
  {
    path: 'evening/:id',
    canActivate: [authGuard],
    loadComponent: () => import('./components/evening-view/evening-view.component').then((m) => m.EveningViewComponent),
  },
  {
    path: 'standings',
    canActivate: [roleGuard('admin')],
    loadComponent: () => import('./components/standings/standings.component').then((m) => m.StandingsComponent),
  },
  {
    path: 'info',
    canActivate: [roleGuard('admin')],
    loadComponent: () => import('./components/info/info.component').then((m) => m.InfoComponent),
  },
  {
    path: 'beheer',
    canActivate: [roleGuard('admin')],
    loadComponent: () => import('./components/beheer/beheer.component').then((m) => m.BeheerComponent),
  },
  {
    path: 'gebruikers',
    canActivate: [roleGuard('admin')],
    loadComponent: () => import('./components/gebruikers/gebruikers.component').then((m) => m.GebruikersComponent),
  },
  {
    path: 'm',
    canActivate: [authGuard],
    loadChildren: () => import('./mobile/mobile.routes').then((m) => m.mobileRoutes),
  },
  { path: '**', redirectTo: 'schema' },
];
