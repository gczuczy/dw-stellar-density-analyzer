import { CanActivateFn, Router }  from '@angular/router';
import { inject }                 from '@angular/core';
import { AuthService }            from './auth.service';

/**
 * Functional route guard (Angular 15+ style – no class needed).
 *
 * • If the user is logged in          → allows navigation.
 * • If the user is NOT logged in      → triggers the PKCE login flow.
 *   The library stores the intended path in OAuth state so that after
 *   a successful token exchange we can navigate back here.
 *
 * Usage in route definitions:
 *   canActivate: [authGuard]
 */
export const authGuard: CanActivateFn = (_route, _state) => {
  const authService = inject(AuthService);
  const router      = inject(Router);

  if (authService.isLoggedIn) {
    return true;
  }

  // Store the target URL so we can return after login.
  // angular-oauth2-oidc's initCodeFlowWithState already does this
  // but we also persist it ourselves as a safety net.
  sessionStorage.setItem('_redirectUri', _state.url);

  // Kick off the PKCE flow; the library will redirect the browser.
  authService.login();

  // Return false – navigation is aborted here; the page will reload
  // after the OAuth redirect anyway.
  return false;
};
