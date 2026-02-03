import { Injectable }                       from '@angular/core';
import { OAuthService, AuthConfig }        from 'angular-oauth2-oidc';
import { Observable, from }               from 'rxjs';
import { map, switchMap, tap }            from 'rxjs/operators';
import { Router }                         from '@angular/router';
import { ApiService }                     from '../services/api.service';

/**
 * The flat key→value shape that `/api/oauth/config` must return.
 *
 * Example response from the backend:
 * {
 *   "clientId":         "my-spa-client",
 *   "issuer":           "https://auth.example.com/",
 *   "redirectUri":      "http://localhost:4200",
 *   "scope":            "openid profile email"
 * }
 */
export interface OAuthBackendConfig {
  clientId:            string;
  redirectUri:         string;
  scope:               string;
  issuer:              string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {

  constructor(
    private oauth  : OAuthService,
    private api    : ApiService,
    private router : Router
  ) {}

  /**
   * Called once by APP_INITIALIZER.
   *
   * 1. GETs /api/oauth/config  →  flat JSON  →  AuthConfig
   * 2. Configures OAuthService (PKCE is automatic when responseType = 'code')
   * 3. Loads discovery document and tries login (exchanges code if present)
   * 4. After exchange, navigates back to the originally-requested path
   */
  initialize(): Observable<void> {
    return this.api.getConfig('/api/oauth/config').pipe(

      tap((cfg: Record<string, string>) => {
        const authConfig: AuthConfig = {
          issuer:       cfg['issuer'],
          clientId:     cfg['clientId'],
          redirectUri:  cfg['redirectUri']  || window.location.origin,
          scope:        cfg['scope']        || 'openid profile email',
          responseType: 'code',   // PKCE is activated automatically
          showDebugInformation: false,
        };
        this.oauth.configure(authConfig);
      }),

      switchMap(() =>
        // Load discovery document and try login in one call
        from(this.oauth.loadDiscoveryDocumentAndTryLogin())
      ),

      tap(() => {
        // Redirect back to the page the user originally wanted to visit
        const target = sessionStorage.getItem('_redirectUri');
        if (target && this.oauth.hasValidAccessToken()) {
          sessionStorage.removeItem('_redirectUri');
          this.router.navigateByUrl(target);
        }
      }),

      map(() => undefined)
    );
  }

  // ── queries ───────────────────────────────────────────────────────────────

  get isLoggedIn(): boolean {
    return this.oauth.hasValidAccessToken();
  }

  /** Decoded ID-token payload, or null. */
  get userInfo(): Record<string, unknown> | null {
    const claims = this.oauth.getIdentityClaims();
    return claims ? claims as Record<string, unknown> : null;
  }

  get accessToken(): string | null {
    return this.oauth.getAccessToken();
  }

  // ── actions ───────────────────────────────────────────────────────────────

  /** Start the PKCE Authorization-Code flow (browser redirect to IdP). */
  login(): void {
    // Store where we came from
    sessionStorage.setItem('_redirectUri', window.location.pathname);
    this.oauth.initCodeFlow();
  }

  /** Logout and clear tokens. */
  logout(): void {
    this.oauth.logOut();
  }
}
