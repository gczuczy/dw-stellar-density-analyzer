import { Provider, APP_INITIALIZER }   from '@angular/core';
import { provideOAuthClient }          from 'angular-oauth2-oidc';
import { AuthService }                 from './auth.service';
import { ThemeService }                from './theme.service';

/**
 * Call this from the root providers[] in main.ts.
 *
 * 1. provideOAuthClient() – the official standalone provider from
 *    angular-oauth2-oidc v15+.  Registers OAuthService, OAuthStorage,
 *    and the event bus in one call.
 * 2. APP_INITIALIZER – runs before the first route is resolved:
 *      a) Forces ThemeService instantiation (sets data-bs-theme before first paint)
 *      b) Calls AuthService.initialize() which fetches /api/oauth/config,
 *         configures OAuthService, and finishes any pending code-exchange.
 */
export function provideOAuthService(): Provider[] {
  return [
    provideOAuthClient(),          // ← replaces manual OAuthService + OAuthStorage

    {
      provide:    APP_INITIALIZER,
      multi:      true,
      useFactory: (auth: AuthService, _theme: ThemeService) => {
        // ThemeService constructor already applied data-bs-theme;
        // injecting it here only to guarantee early instantiation.
        return () => auth.initialize().toPromise();
      },
      deps: [AuthService, ThemeService]
    }
  ];
}
