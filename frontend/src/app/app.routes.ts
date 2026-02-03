import { Routes }              from '@angular/router';
import { authGuard }           from './auth/auth.guard';

// ── Eager imports (small components – no need to lazy-load) ─────────────────
import { HomeComponent }       from './components/home/home.component';
import { BarfooComponent }     from './components/barfoo/barfoo.component';
import { FoobarComponent }     from './components/foobar/foobar.component';
import { SettingsComponent }   from './components/settings/settings.component';
import { SidemenuComponent }   from './components/sidemenu/sidemenu.component';
import { SidemenuAlphaComponent } from './components/sidemenu/sidemenu-alpha.component';
import { SidemenuBetaComponent }  from './components/sidemenu/sidemenu-beta.component';

export const routes: Routes = [
  // ── public ──────────────────────────────────────────────────────────────
  { path: '',        component: HomeComponent,    title: 'Home' },
  { path: 'barfoo',  component: BarfooComponent,  title: 'Barfoo' },

  // ── protected ───────────────────────────────────────────────────────────
  { path: 'foobar',   component: FoobarComponent,   canActivate: [authGuard], title: 'Foobar' },
  { path: 'settings', component: SettingsComponent, canActivate: [authGuard], title: 'Settings' },

  // ── protected – nested (side-menu layout) ──────────────────────────────
  {
    path: 'sidemenu',
    component: SidemenuComponent,   // layout shell (sidebar + router-outlet)
    canActivate: [authGuard],
    title: 'Side Menu',
    children: [
      { path: '',      redirectTo: 'alpha', pathMatch: 'full' },
      { path: 'alpha', component: SidemenuAlphaComponent, title: 'Alpha' },
      { path: 'beta',  component: SidemenuBetaComponent,  title: 'Beta'  },
    ]
  },

  // ── catch-all ───────────────────────────────────────────────────────────
  { path: '**', redirectTo: '' }
];
