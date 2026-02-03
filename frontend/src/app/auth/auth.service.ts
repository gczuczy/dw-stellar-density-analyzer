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
 *   "redirectUri":      "https://localhost:4200/callback",
 *   "scope":            "openid profile email",
 *   "authorizationUrl": "https://auth.example.com/authorize",
 *   "tokenUrl":         "https://auth.example.com/oauth/token",
 *   "jwksUrl":          "https://auth.example.com/.well-known/jwks.json"
 * }
 */
export interface OAuthBackendConfig {
  clientId:            string;
  authorizationUrl:    string;
  tokenUrl:            string;
  redirectUri:         string;
  scope:               string;
  issuer?:             string;
  jwksUrl?:            string;
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
   * 2. Configures OAuthService  (PKCE is automatic when responseType = 'code')
   * 3. tryLogin()  –  if ?code= is present the library exchanges it silently
   * 4. After exchange, navigates back to the originally-requested path
   *    (stored in sessionStorage by authGuard)
   */
  initialize(): Observable<void> {
    return this.api.getConfig('/api/oauth/config').pipe(

      tap((cfg: Record<string, string>) => {
        const authConfig: AuthConfig = {
          clientId:              cfg['clientId'],
          issuer:                cfg['issuer']            || '',
          redirectUri:           cfg['redirectUri']       || (window.location.origin + '/callback'),
          scope:                 cfg['scope']             || 'openid profile email',
          responseType:          'code',   // PKCE is activated automatically
          authorizationEndpoint: cfg['authorizationUrl'],
          tokenEndpoint:         cfg['tokenUrl'],
          jwksUrl:               cfg['jwksUrl'],          // optional
        };
        this.oauth.configure(authConfig);
      }),

      switchMap(() =>
        from(this.oauth.tryLogin({ disableOidcChecks: true }))
      ),

      tap(() => {
        // Redirect back to the page the user originally wanted to visit
        const target = sessionStorage.getItem('_redirectUri');
        if (target) {
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
    const raw = this.oauth.getIdToken();
    if (!raw) { return null; }
    try {
      return JSON.parse(atob(raw.split('.')[1]));
    } catch {
      return null;
    }
  }

  get accessToken(): string | null {
    return this.oauth.getAccessToken();
  }

  // ── actions ───────────────────────────────────────────────────────────────

  /** Start the PKCE Authorization-Code flow (browser redirect to IdP). */
  login(): void {
    this.oauth.initCodeFlowWithState(window.location.pathname);
  }

  /** Logout + optional token revocation. */
  logout(): void {
    this.oauth.logoutWithRedirect();
  }
}
